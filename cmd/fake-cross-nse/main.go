package main

import (
	"context"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/edwarnicke/grpcfd"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	api_registry "github.com/networkservicemesh/api/pkg/api/registry"
	"github.com/networkservicemesh/sdk/pkg/networkservice/chains/client"
	"github.com/networkservicemesh/sdk/pkg/networkservice/chains/endpoint"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/authorize"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/clienturl"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/connect"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/mechanisms/recvfd"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/mechanisms/sendfd"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/adapters"
	"github.com/networkservicemesh/sdk/pkg/registry/common/interpose"
	sendfd2 "github.com/networkservicemesh/sdk/pkg/registry/common/sendfd"
	"github.com/networkservicemesh/sdk/pkg/registry/core/chain"
	"github.com/networkservicemesh/sdk/pkg/tools/addressof"
	"github.com/networkservicemesh/sdk/pkg/tools/grpcutils"
	"github.com/networkservicemesh/sdk/pkg/tools/log"
	"github.com/networkservicemesh/sdk/pkg/tools/signalctx"
	"github.com/networkservicemesh/sdk/pkg/tools/spiffejwt"
)

// Config is configuration for cmd-testing-registry-client
type Config struct {
	Name      string  `default:"cross-nse" desc:"application name" split_words:"true"`
	ConnectTo url.URL `default:"unix:///var/lib/networkservicemesh/nsm.io.sock" desc:"url to connect to NSM" split_words:"true"`
}

func main() {
	ctx := signalctx.WithSignals(context.Background())
	ctx, cancel := context.WithCancel(ctx)

	// Get config from environment
	config := &Config{}
	if err := envconfig.Usage("fake-cross-nse", config); err != nil {
		logrus.Fatal(err)
	}
	if err := envconfig.Process("fake-cross-nse", config); err != nil {
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
	var crossNse endpoint.Endpoint
	crossNse = endpoint.NewServer(ctx,
		config.Name,
		authorize.NewServer(),
		spiffejwt.TokenGeneratorFunc(source, time.Hour*24),
		// Statically set the url we use to the unix file socket for the NSMgr
		recvfd.NewServer(),
		clienturl.NewServer(&config.ConnectTo),
		connect.NewServer(
			ctx,
			client.NewClientFactory(
				config.Name,
				// What to call onHeal
				addressof.NetworkServiceClient(adapters.NewServerToClient(crossNse)),
				spiffejwt.TokenGeneratorFunc(source, time.Hour*24),
				recvfd.NewClient(),
			),
			grpc.WithTransportCredentials(grpcfd.TransportCredentials(credentials.NewTLS(tlsconfig.MTLSClientConfig(source, source, tlsconfig.AuthorizeAny())))),
			grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
		),
		sendfd.NewServer(),
	)

	server := grpc.NewServer(
		grpc.Creds(
			grpcfd.TransportCredentials(
				credentials.NewTLS(
					tlsconfig.MTLSServerConfig(
						source,
						source,
						tlsconfig.AuthorizeAny()),
				),
			),
		),
	)

	crossNse.Register(server)

	tmpDir, err := ioutil.TempDir("", "fake-cross-nse-")
	if err != nil {
		logrus.Fatalf("error creating tmpDir %+v", err)
	}
	defer func(tmpDir string) { _ = os.Remove(tmpDir) }(tmpDir)
	listenOn := &(url.URL{Scheme: "unix", Path: filepath.Join(tmpDir, "listen.on")})
	srvErrCh := grpcutils.ListenAndServe(ctx, listenOn, server)
	exitOnErrCh(ctx, cancel, srvErrCh)

	cc, err := grpc.DialContext(ctx,
		config.ConnectTo.String(),
		grpc.WithTransportCredentials(grpcfd.TransportCredentials(credentials.NewTLS(tlsconfig.MTLSClientConfig(source, source, tlsconfig.AuthorizeAny())))),
		grpc.WithBlock(),
	)
	if err != nil {
		logrus.Fatal(err.Error())
	}

	forwarderRegistrationClient := chain.NewNetworkServiceEndpointRegistryClient(
		sendfd2.NewNetworkServiceEndpointRegistryClient(),
		interpose.NewNetworkServiceEndpointRegistryClient(),
		api_registry.NewNetworkServiceEndpointRegistryClient(cc),
	)

	_, err = forwarderRegistrationClient.Register(context.Background(), &api_registry.NetworkServiceEndpoint{
		Url:  listenOn.String(),
		Name: "fake-cross-nse",
	})
	if err != nil {
		logrus.Fatal(err.Error())
	}

	<-ctx.Done()
}

func exitOnErrCh(ctx context.Context, cancel context.CancelFunc, errCh <-chan error) {
	// If we already have an error, log it and exit
	select {
	case err := <-errCh:
		log.Entry(ctx).Fatal(err)
	default:
	}
	// Otherwise wait for an error in the background to log and cancel
	go func(ctx context.Context, errCh <-chan error) {
		err := <-errCh
		log.Entry(ctx).Error(err)
		cancel()
	}(ctx, errCh)
}
