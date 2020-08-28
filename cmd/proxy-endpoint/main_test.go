package main_test

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/cls"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/kernel"
	registryapi "github.com/networkservicemesh/api/pkg/api/registry"
	main "github.com/networkservicemesh/integration-k8s-kind/cmd/proxy-endpoint"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/authorize"
	"github.com/networkservicemesh/sdk/pkg/tools/callback"
	"github.com/networkservicemesh/sdk/pkg/tools/spiffejwt"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/svid/x509svid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/networkservicemesh/sdk/pkg/networkservice/common/excludedprefixes"

	"github.com/networkservicemesh/sdk/pkg/networkservice/chains/registry"
	"github.com/networkservicemesh/sdk/pkg/tools/grpcutils"

	"github.com/networkservicemesh/sdk/pkg/registry/common/setid"
	"github.com/networkservicemesh/sdk/pkg/registry/common/seturl"
	chain_registry "github.com/networkservicemesh/sdk/pkg/registry/core/chain"
	"github.com/networkservicemesh/sdk/pkg/registry/core/nextwrap"
	"github.com/networkservicemesh/sdk/pkg/registry/memory"

	"github.com/networkservicemesh/sdk/pkg/networkservice/chains/client"
	"github.com/networkservicemesh/sdk/pkg/networkservice/chains/endpoint"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/connect"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/discover"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/localbypass"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/roundrobin"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/adapters"
	adapter_registry "github.com/networkservicemesh/sdk/pkg/registry/core/adapters"
	"github.com/networkservicemesh/sdk/pkg/tools/addressof"
	"github.com/networkservicemesh/sdk/pkg/tools/token"
)

type nsmgrServer struct {
	endpoint.Endpoint
	registry.Registry
}

func newServer(ctx context.Context, nsmRegistration *registryapi.NetworkServiceEndpoint, authzServer networkservice.NetworkServiceServer, tokenGenerator token.GeneratorFunc, clientDialOptions ...grpc.DialOption) *nsmgrServer {
	rv := &nsmgrServer{}

	var localbypassRegistryServer registryapi.NetworkServiceEndpointRegistryServer

	nsRegistry := memory.NewNetworkServiceRegistryServer()
	nseRegistry := chain_registry.NewNetworkServiceEndpointRegistryServer(
		setid.NewNetworkServiceEndpointRegistryServer(),  // If no remote registry then assign ID.
		memory.NewNetworkServiceEndpointRegistryServer(), // Memory registry to store result inside.
	)
	// Construct Endpoint
	rv.Endpoint = endpoint.NewServer(ctx,
		nsmRegistration.Name,
		authzServer,
		tokenGenerator,
		discover.NewServer(adapter_registry.NetworkServiceServerToClient(nsRegistry),
			adapter_registry.NetworkServiceEndpointServerToClient(nseRegistry)),
		roundrobin.NewServer(),
		localbypass.NewServer(&localbypassRegistryServer),
		excludedprefixes.NewServer(ctx),
		connect.NewServer(
			ctx,
			client.NewClientFactory(nsmRegistration.Name,
				addressof.NetworkServiceClient(
					adapters.NewServerToClient(rv)),
				tokenGenerator),
			clientDialOptions...),
	)

	nsChain := chain_registry.NewNetworkServiceRegistryServer(nsRegistry)
	nseChain := chain_registry.NewNetworkServiceEndpointRegistryServer(
		localbypassRegistryServer,                                           // Store endpoint Id to EndpointURL for local access.
		seturl.NewNetworkServiceEndpointRegistryServer(nsmRegistration.Url), // Remember endpoint URL
		nseRegistry,                                                         // Register NSE inside Remote registry with ID assigned
	)
	rv.Registry = registry.NewServer(nsChain, nseChain)

	return rv
}

func newRemoteNSServer(cc grpc.ClientConnInterface) registryapi.NetworkServiceRegistryServer {
	if cc != nil {
		return adapter_registry.NetworkServiceClientToServer(
			nextwrap.NewNetworkServiceRegistryClient(
				registryapi.NewNetworkServiceRegistryClient(cc)))
	}
	return nil
}
func newRemoteNSEServer(cc grpc.ClientConnInterface) registryapi.NetworkServiceEndpointRegistryServer {
	if cc != nil {
		return adapter_registry.NetworkServiceEndpointClientToServer(
			nextwrap.NewNetworkServiceEndpointRegistryClient(
				registryapi.NewNetworkServiceEndpointRegistryClient(cc)))
	}
	return nil
}

func (n *nsmgrServer) Register(s *grpc.Server) {
	grpcutils.RegisterHealthServices(s, n, n.NetworkServiceEndpointRegistryServer(), n.NetworkServiceRegistryServer())
	networkservice.RegisterNetworkServiceServer(s, n)
	networkservice.RegisterMonitorConnectionServer(s, n)
	registryapi.RegisterNetworkServiceRegistryServer(s, n.Registry.NetworkServiceRegistryServer())
	registryapi.RegisterNetworkServiceEndpointRegistryServer(s, n.Registry.NetworkServiceEndpointRegistryServer())
}

var _ endpoint.Endpoint = &nsmgrServer{}
var _ registry.Registry = &nsmgrServer{}

type endpointImpl struct {
}

func (e *endpointImpl) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	if request.Connection.Context == nil {
		request.Connection.Context = &networkservice.ConnectionContext{
		}
	}
	if request.Connection.Context.ExtraContext == nil {
		request.Connection.Context.ExtraContext = map[string]string{}
	}
	request.Connection.Context.ExtraContext["processed"] = "ok"

	return request.Connection, nil
}

func (e *endpointImpl) Close(ctx context.Context, connection *networkservice.Connection) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func TestProxyEndpoint(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2000*time.Second)
	defer cancel()

	// ********************************************************************************
	// Get a X509Source
	// ********************************************************************************
	source, err := workloadapi.NewX509Source(ctx)
	if err != nil {
		logrus.Fatalf("error getting x509 source: %+v", err)
	}
	var svid *x509svid.SVID
	svid, err = source.GetX509SVID()
	if err != nil {
		logrus.Fatalf("error getting x509 svid: %+v", err)
	}
	logrus.Infof("SVID: %q", svid.ID)

	// Print all grpc operations.
	grpclog.SetLoggerV2(grpclog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout))

	// ********************************************************************************
	// Construct callback server
	// ********************************************************************************
	callbackServer := callback.NewServer(main.IdentityByEndpointID)

	// ********************************************************************************
	// Start nsmgr
	// ********************************************************************************
	mgr := newServer(ctx, &registryapi.NetworkServiceEndpoint{
		Name: "nsmgr",
		Url:  "",
	}, authorize.NewServer(), spiffejwt.TokenGeneratorFunc(source, time.Second*1000),
		callbackServer.WithCallbackDialer(),

		// Security to connect to endpoint
		grpc.WithTransportCredentials(
			credentials.NewTLS(
				tlsconfig.MTLSClientConfig(source, source, tlsconfig.AuthorizeAny()))),
		grpc.WithDefaultCallOptions(
			grpc.WaitForReady(true)))

	// Construct a manager listening on random tcp port
	mgrUrl := &url.URL{Scheme: "tcp", Host: "127.0.0.1:0"}
	server := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsconfig.MTLSServerConfig(source, source, tlsconfig.AuthorizeAny()))))
	mgr.Register(server)

	// Register callback serve to grpc.
	callback.RegisterCallbackServiceServer(server, callbackServer)

	_ = grpcutils.ListenAndServe(ctx, mgrUrl, server)

	// ********************************************************************************
	// Now we have an test nsmgr, and we could register proxy endpoint.
	// ********************************************************************************

	config := &main.Config{
		Name:             "proxy-endpoint",
		ListenOn:         url.URL{Host: "127.0.0.1:0", Scheme: "tcp"},
		ConnectTo:        *mgrUrl,
		MaxTokenLifetime: time.Hour * 24,
	}
	require.NoError(t, main.StartNSMProxyEndpoint(ctx, config))

	// ********************************************************************************
	// Now we have an endpoint up and running, we could connect to it and try perform request
	// ********************************************************************************

	// Construct a local endpoint and register it
	testEp := &endpointImpl{
	}

	epSrv := grpc.NewServer()
	networkservice.RegisterNetworkServiceServer(epSrv, testEp) // testEp will recieve a Request's

	clientCtx := main.WithCallbackEndpointID(ctx, fmt.Sprintf("my_endpoint/client"))

	// Construct connection to Proxy endpoint
	clientCC, err2 := grpc.DialContext(clientCtx, grpcutils.URLToTarget(&config.ListenOn), grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.WaitForReady(true)))
	require.NoError(t, err2)

	// Connect to nsmgr with callback URI
	callbackClient := callback.NewClient(clientCC, epSrv)
	callbackClient.Serve(clientCtx)

	// *****************************************************************************
	// Register endpoint to proxy endpoint and nsmgr
	// *****************************************************************************
	registryClient := registryapi.NewNetworkServiceEndpointRegistryClient(clientCC)

	_, err4 := registryapi.NewNetworkServiceRegistryClient(clientCC).Register(clientCtx, &registryapi.NetworkService{
		Name:    "network-service",
		Payload: "ip",
	})

	require.NoError(t, err4)

	_, err3 := registryClient.Register(clientCtx, &registryapi.NetworkServiceEndpoint{
		Name:                "endpoint",
		NetworkServiceNames: []string{"network-service"},
		Url:                 "callback:my_endpoint/client",
	})
	require.NoError(t, err3)

	// Check we had all stuff inside nsmgr.

	findResult, findErr := registryClient.Find(ctx, &registryapi.NetworkServiceEndpointQuery{
		Watch: false, NetworkServiceEndpoint: &registryapi.NetworkServiceEndpoint{
			NetworkServiceNames: []string{"network-service"},
		}})
	require.NoError(t, findErr)
	endpoints := registryapi.ReadNetworkServiceEndpointList(findResult)
	require.NotNil(t, endpoints)
	require.Equal(t, 1, len(endpoints))

	// Construct a normal client and perform request to our endpoint.
	var nsmgrClient *grpc.ClientConn
	nsmgrClient, err = grpc.DialContext(ctx,
		grpcutils.URLToTarget(mgrUrl),
		grpc.WithTransportCredentials(
			credentials.NewTLS(
				tlsconfig.MTLSClientConfig(source, source, tlsconfig.AuthorizeAny()))),
		grpc.WithDefaultCallOptions(
			grpc.WaitForReady(true)))

	client := client.NewClient(ctx, "nsc", nil, spiffejwt.TokenGeneratorFunc(source, config.MaxTokenLifetime), nsmgrClient)

	var conn *networkservice.Connection
	conn, err = client.Request(ctx, &networkservice.NetworkServiceRequest{
		Connection: &networkservice.Connection{
			NetworkService: "network-service",
			Context: &networkservice.ConnectionContext{},
		},
		MechanismPreferences: []*networkservice.Mechanism{
			{
				Cls:  cls.LOCAL,
				Type: kernel.MECHANISM,
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, conn)
	require.Equal(t, "ok", conn.Context.ExtraContext["processed"])

}
