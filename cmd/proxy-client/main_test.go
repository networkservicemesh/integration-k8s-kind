package main_test

import (
	"context"
	"github.com/networkservicemesh/api/pkg/api/networkservice"
	main "github.com/networkservicemesh/integration-k8s-kind/cmd/proxy-client"
	"github.com/networkservicemesh/sdk/pkg/networkservice/chains/endpoint"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/authorize"
	"github.com/networkservicemesh/sdk/pkg/tools/grpcutils"
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
)

func TokenGenerator(peerAuthInfo credentials.AuthInfo) (token string, expireTime time.Time, err error) {
	return "TestToken", time.Date(3000, 1, 1, 1, 1, 1, 1, time.UTC), nil
}
func TestProxyClient(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get a X509Source
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

	grpclog.SetLoggerV2(grpclog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout))

	mgr := endpoint.NewServer(ctx, "test", authorize.NewServer(), spiffejwt.TokenGeneratorFunc(source, time.Second*1000))

	mgrUrl := &url.URL{Scheme: "tcp", Host: "127.0.0.1:0"}
	endpoint.Serve(ctx, mgrUrl, mgr, grpc.Creds(credentials.NewTLS(tlsconfig.MTLSServerConfig(source, source, tlsconfig.AuthorizeAny()))))

	cfg := &main.Config{
		Name:      "proxy-nsc",
		ConnectTo: *mgrUrl,
		ListenOn:  url.URL{Scheme: "tcp", Host: "127.0.0.1:0"}, // Some random port we would like to connect
	}
	client, e := main.NewNSMProxyClient(ctx, cfg)
	require.NoError(t, e)

	errChan := main.RunProxyClient(ctx, cfg, client)
	require.NotNil(t, errChan)

	// Obtain a connection
	var clientCC *grpc.ClientConn
	clientCC, err = grpc.DialContext(ctx,
		grpcutils.URLToTarget(&cfg.ListenOn),
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.WaitForReady(true)))

	require.NoError(t, err)

	defer func() { _ = clientCC.Close() }()

	cCon := networkservice.NewNetworkServiceClient(clientCC)

	var conn *networkservice.Connection
	conn, err = cCon.Request(ctx, &networkservice.NetworkServiceRequest{})

	require.NoError(t, err)
	require.NotNil(t, conn)

	require.Equal(t, 2, len(conn.Path.PathSegments))
}
