package main

import (
	"context"
	"errors"
	"net"
	"net/url"
	"os"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	api_registry "github.com/networkservicemesh/api/pkg/api/registry"

	"github.com/networkservicemesh/sdk/pkg/tools/debug"
	"github.com/networkservicemesh/sdk/pkg/tools/log"
	"github.com/networkservicemesh/sdk/pkg/tools/signalctx"
)

// Config is configuration for cmd-testing-registry-client
type Config struct {
	ConnectTo                  url.URL `desc:"url to the local registry that handles this domain" split_words:"true"`
	NetworkServiceName         string  `default:"icmp-responder" desc:"url to the local registry that handles this domain" split_words:"true"`
	NetworkServiceEndpointName string  `default:"icmp-responder-nse" desc:"url to the local registry that handles this domain" split_words:"true"`
}

// WaitForPortAvailable waits while the port will is available. Throws exception if the context is done.
func WaitForPortAvailable(ctx context.Context, protoType, registryAddress string, idleSleep time.Duration) error {
	if idleSleep < 0 {
		return errors.New("idleSleep must be positive")
	}
	logrus.Infof("Waiting for liveness probe: %s:%s", protoType, registryAddress)
	last := time.Now()

	for {
		select {
		case <-ctx.Done():
			return errors.New("timeout waiting for: " + protoType + ":" + registryAddress)
		default:
			var d net.Dialer
			conn, err := d.DialContext(ctx, protoType, registryAddress)
			if conn != nil {
				_ = conn.Close()
			}
			if err == nil {
				return nil
			}
			if time.Since(last) > time.Minute {
				logrus.Infof("Waiting for liveness probe: %s:%s", protoType, registryAddress)
				last = time.Now()
			}
			// Sleep to not overflow network
			<-time.After(idleSleep)
		}
	}
}

func main() {
	// Setup context to catch signals
	ctx := signalctx.WithSignals(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Setup logging
	logrus.SetFormatter(&nested.Formatter{})
	logrus.SetLevel(logrus.TraceLevel)
	ctx = log.WithField(ctx, "cmd", os.Args[0])

	// Debug self if necessary
	if err := debug.Self(); err != nil {
		log.Entry(ctx).Infof("%s", err)
	}
	// Get config from environment
	config := &Config{}
	if err := envconfig.Usage("nse", config); err != nil {
		logrus.Fatal(err)
	}
	if err := envconfig.Process("nse", config); err != nil {
		logrus.Fatalf("error processing config from env: %+v", err)
	}

	log.Entry(ctx).Infof("Config: %#v", config)
	// Get a X509Source
	source, err := workloadapi.NewX509Source(ctx)
	if err != nil {
		logrus.Fatalf("error getting x509 source: %+v", err)
	}

	svid, err := source.GetX509SVID()

	if err != nil {
		logrus.Fatalf("error getting x509 svid: %+v", err)
	}
	logrus.Infof("SVID: %q", svid.ID)

	getTarget := func() string {
		if config.ConnectTo.Scheme == "tcp" {
			return config.ConnectTo.Host
		}
		return config.ConnectTo.String()
	}

	var cc *grpc.ClientConn

	cc, err = grpc.DialContext(ctx,
		getTarget(),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(credentials.NewTLS(tlsconfig.MTLSClientConfig(source, source, tlsconfig.AuthorizeAny()))),
	)

	if err != nil {
		logrus.Fatalf(err.Error())
	}

	_, err = api_registry.NewNetworkServiceRegistryClient(cc).Register(context.Background(), &api_registry.NetworkService{
		Name:    config.NetworkServiceName,
		Payload: "IP",
	})
	if err != nil {
		logrus.Fatalf(err.Error())
	}

	nse, err := api_registry.NewNetworkServiceEndpointRegistryClient(cc).Register(context.Background(), &api_registry.NetworkServiceEndpoint{
		Name:                config.NetworkServiceEndpointName,
		NetworkServiceNames: []string{config.NetworkServiceName},
		ExpirationTime:      &timestamp.Timestamp{Seconds: time.Now().Add(time.Hour * 24).Unix()},
	})

	logrus.Infof("NSE: %+v", nse)

	if err != nil {
		logrus.Fatalf(err.Error())
	}
	logrus.Info("Done!")
	<-ctx.Done()
}
