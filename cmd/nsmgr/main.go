package main

import (
	"context"
	"net/url"
	"os"

	nested "github.com/antonfisher/nested-logrus-formatter"
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
	ConnectTO                      url.URL `desc:"url to the local registry that handles this domain" split_words:"true"`
	FindNetworkServiceName         string  `default:"icmp-responder" desc:"url to the local registry that handles this domain" split_words:"true"`
	FindNetworkServiceEndpointName string  `default:"icmp-responder-nse" desc:"url to the local registry that handles this domain" split_words:"true"`
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
	if err := envconfig.Usage("nsmgr", config); err != nil {
		logrus.Fatal(err)
	}
	if err := envconfig.Process("nsmgr", config); err != nil {
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
	cc, err := grpc.DialContext(ctx,
		config.ConnectTO.String(),
		grpc.WithTransportCredentials(credentials.NewTLS(tlsconfig.MTLSClientConfig(source, source, tlsconfig.AuthorizeAny()))),
		grpc.WithBlock(),
	)

	nsStream, err := api_registry.NewNetworkServiceRegistryClient(cc).Find(context.Background(), &api_registry.NetworkServiceQuery{
		NetworkService: &api_registry.NetworkService{Name: config.FindNetworkServiceName},
	})
	if err != nil {
		logrus.Fatal(err.Error())
	}
	services := api_registry.ReadNetworkServiceList(nsStream)
	if len(services) == 0 {
		logrus.Fatal("Network Service is not found")
	}

	nseStream, err := api_registry.NewNetworkServiceRegistryClient(cc).Find(context.Background(), &api_registry.NetworkServiceQuery{
		NetworkService: &api_registry.NetworkService{Name: config.FindNetworkServiceName},
	})
	if err != nil {
		logrus.Fatal(err.Error())
	}
	nses := api_registry.ReadNetworkServiceList(nseStream)
	if len(nses) == 0 {
		logrus.Fatal("Network Service  Endpoint is not found")
	}

	<-ctx.Done()
}
