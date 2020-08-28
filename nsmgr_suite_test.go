// Copyright (c) 2020 Doc.ai and/or its affiliates.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package integration_k8s_kind_test

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/cls"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/kernel"
	"github.com/networkservicemesh/sdk/pkg/networkservice/chains/client"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/clienturl"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/connect"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/adapters"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"
	"github.com/networkservicemesh/sdk/pkg/registry/common/interpose"
	"github.com/networkservicemesh/sdk/pkg/registry/core/chain"
	"github.com/networkservicemesh/sdk/pkg/tools/addressof"
	"github.com/networkservicemesh/sdk/pkg/tools/log"
	"github.com/networkservicemesh/sdk/pkg/tools/token"

	"github.com/networkservicemesh/integration-k8s-kind/pkg/setlogoption"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/registry"
	"github.com/networkservicemesh/sdk/pkg/networkservice/chains/endpoint"
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/authorize"
	"github.com/networkservicemesh/sdk/pkg/tools/callback"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	v1 "k8s.io/api/core/v1"

	"github.com/networkservicemesh/integration-k8s-kind/k8s"
	"github.com/networkservicemesh/integration-k8s-kind/nsmgr"

	"github.com/networkservicemesh/integration-k8s-kind/spire"

	"github.com/edwarnicke/exechelper"
	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/suite"
)

var l sync.Mutex

type NsmgrTestsSuite struct {
	suite.Suite
	options []*exechelper.Option
	writer  *io.PipeWriter
	prefix  string
}

func (s *NsmgrTestsSuite) SetupSuite() {
	l.Lock()
	defer l.Unlock()

	s.writer = logrus.StandardLogger().Writer()

	s.options = []*exechelper.Option{
		exechelper.WithStderr(s.writer),
		exechelper.WithStdout(s.writer),
	}

	// Remove all from default namespace
	s.Require().NoError(k8s.CleanNamespace(k8s.DefaultNamespace))

	s.Require().NoError(spire.Setup(s.options...))

	s.Require().NoError(nsmgr.Setup(s.options...))

	// Extra dependencies
	s.Require().NoError(k8s.KindRequire("proxy-endpoint", "proxy-client"))
}

func (s *NsmgrTestsSuite) TearDownSuite() {
	l.Lock()
	defer l.Unlock()
	_ = exechelper.Run("kubectl get pods -o wide --all-namespaces", s.options...)

	// Show logs from nsmgr
	k8s.ShowLogs("nsmgr", s.options...)

	s.Require().NoError(spire.Delete(s.options...))
	s.Require().NoError(nsmgr.Delete(s.options...))
}

func (s *NsmgrTestsSuite) TearDownTest() {
	s.Require().NoError(k8s.CleanNamespace(k8s.DefaultNamespace))
}

func TokenGenerator(peerAuthInfo credentials.AuthInfo) (t string, expireTime time.Time, err error) {
	return "TestToken", time.Date(3000, 1, 1, 1, 1, 1, 1, time.UTC), nil
}

type localEndpoint struct {
	requests int
	closes   int
}

func (l *localEndpoint) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	l.requests++
	return request.GetConnection(), nil
}

func (l *localEndpoint) Close(ctx context.Context, connection *networkservice.Connection) (*empty.Empty, error) {
	l.closes++
	return &empty.Empty{}, nil
}

// WithCallbackEndpointID - pass with :endpoint-id a correct endpoint identity.
func WithCallbackEndpointID(ctx context.Context, epName string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "endpoint-id", epName)
}

type myEndpouint struct {
	endpoint.Endpoint
}

// NewCrossNSE construct a new Cross connect test NSE
func newCrossNSE(ctx context.Context, name string, connectTo *url.URL, tokenGenerator token.GeneratorFunc, clientDialOptions ...grpc.DialOption) endpoint.Endpoint {
	var crossNSe = &myEndpouint{}
	crossNSe.Endpoint = endpoint.NewServer(ctx,
		name,
		next.NewNetworkServiceServer(
			setlogoption.NewServer(map[string]string{"cmd": "crossNSE"}),
			authorize.NewServer(),
		),
		tokenGenerator,
		// Statically set the url we use to the unix file socket for the NSMgr
		clienturl.NewServer(connectTo),
		connect.NewServer(
			ctx,
			client.NewClientFactory(
				name,
				// What to call onHeal
				addressof.NetworkServiceClient(adapters.NewServerToClient(crossNSe)),
				tokenGenerator,
			),
			clientDialOptions...,
		),
	)

	return crossNSe
}

func (s *NsmgrTestsSuite) TestProxyEndpointClient() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = log.WithField(ctx, "cmd", "TestProxyEndpointClient")

	// ****************************************************************************************
	// Deploy proxy endpoint
	// ****************************************************************************************
	endpointMgrURL, endpointClient := s.deployForwardPod(ctx, "./deployments/proxy-endpoint.yaml", 6001)
	defer func() { _ = endpointClient.Close() }()

	// ****************************************************************************************
	// Deploy proxy client
	// ****************************************************************************************
	_, proxyClient := s.deployForwardPod(ctx, "./deployments/proxy-client.yaml", 6001)
	defer func() { _ = endpointClient.Close() }()

	// ****************************************************************************************
	// Construct a proxy registry
	// ****************************************************************************************
	regClient := registry.NewNetworkServiceEndpointRegistryClient(endpointClient)

	// Register Interpose endpoint
	s.serveEndpoint(ctx, newCrossNSE(ctx, "interpose", &url.URL{Scheme: "tcp", Host: endpointMgrURL}, TokenGenerator, grpc.WithInsecure(), grpc.WithBlock()), endpointClient, "my-interpose")

	// ****************************************************************************************
	// Construct interpose reg client.
	// ****************************************************************************************

	interposeRegClient := chain.NewNetworkServiceEndpointRegistryClient(interpose.NewNetworkServiceEndpointRegistryClient(), regClient)
	interposeEpReg, interposeEpRegErr := interposeRegClient.Register(ctx, &registry.NetworkServiceEndpoint{
		Url: "callback:my-interpose",
	})

	defer func() {
		_, _ = interposeRegClient.Unregister(context.Background(), interposeEpReg)
	}()

	s.Require().NoError(interposeEpRegErr)
	s.Require().NotNil(interposeEpReg)

	// Register network service

	nsRegClient := registry.NewNetworkServiceRegistryClient(endpointClient)
	ns, nsRegErr := nsRegClient.Register(ctx, &registry.NetworkService{
		Name:    "my-service",
		Payload: "none",
	})
	s.Require().NoError(nsRegErr)
	s.Require().NotNil(ns)

	defer func() {
		_, _ = nsRegClient.Unregister(context.Background(), ns)
	}()

	// ****************************************************************************************
	// Register Final endpoint
	// ****************************************************************************************
	ep := endpoint.NewServer(ctx, "my-endpoint",
		next.NewNetworkServiceServer(
			setlogoption.NewServer(map[string]string{"cmd": "localEndpoint"}),
			authorize.NewServer(),
		),
		TokenGenerator, &localEndpoint{})

	s.serveEndpoint(ctx, ep, endpointClient, "my-endpoint")

	epReg, epRegErr := regClient.Register(ctx, &registry.NetworkServiceEndpoint{
		Url:                 "callback:my-endpoint",
		NetworkServiceNames: []string{"my-service"},
	})
	s.Require().NoError(epRegErr)
	s.Require().NotNil(epReg)

	defer func() {
		_, _ = regClient.Unregister(context.Background(), epReg)
	}()

	// Start and connect client, and try connect to endpoint

	reqCtx, reqCan := context.WithTimeout(ctx, 10*time.Second)
	defer reqCan()
	nsClient := client.NewClient(reqCtx, "nsc", nil, TokenGenerator, proxyClient)
	connection, connErr := nsClient.Request(reqCtx, s.newRequest())
	s.Require().NoError(connErr)
	s.Require().NotNil(connection)

	_, closeErr := nsClient.Close(reqCtx, connection)
	s.Require().NoError(closeErr)
}

func (s *NsmgrTestsSuite) newRequest() *networkservice.NetworkServiceRequest {
	return &networkservice.NetworkServiceRequest{
		Connection: &networkservice.Connection{
			Context: &networkservice.ConnectionContext{
				IpContext: &networkservice.IPContext{
					DstIpRequired: true,
					SrcIpRequired: true,
				},
			},
			NetworkService: "my-service",
		},
		MechanismPreferences: []*networkservice.Mechanism{
			{
				Cls:  cls.LOCAL,
				Type: kernel.MECHANISM,
			},
		},
	}
}

func (s *NsmgrTestsSuite) serveEndpoint(ctx context.Context, ep endpoint.Endpoint, endpointClient grpc.ClientConnInterface, id string) {
	server := grpc.NewServer()
	ep.Register(server)
	cbClient := callback.NewClient(endpointClient, server)
	cbClient.Serve(WithCallbackEndpointID(ctx, id))
}

func (s *NsmgrTestsSuite) deployForwardPod(ctx context.Context, deployment string, podPort int) (string, *grpc.ClientConn) {
	// ****************************************************************************************
	// Proxy to endpoint
	// ****************************************************************************************
	dep, err := k8s.ApplyGetDeployment(deployment)
	s.Require().NoError(err)
	s.Require().NoError(k8s.WaitDeploymentReady(dep, 1, time.Second*30))

	var pod *v1.Pod
	pod, err = k8s.DescribePod(k8s.DefaultNamespace, "", dep.Labels)
	s.Require().NoError(err)

	localPort, perr := k8s.GetFreePort()
	s.Require().NoError(perr)

	_ = exechelper.Start(fmt.Sprintf("kubectl port-forward %v %v:%v", pod.Name, localPort, podPort), append(s.options, exechelper.WithContext(ctx))...)
	logrus.Infof("Proxy pod started: %v", pod.Name)

	dialCtx, dialCancel := context.WithTimeout(ctx, time.Second*10)
	defer dialCancel()

	// ****************************************************************************************
	// Connect to proxy endpoint
	// ****************************************************************************************
	podURL := fmt.Sprintf("127.0.0.1:%d", localPort)
	var podClient *grpc.ClientConn
	podClient, err = grpc.DialContext(dialCtx, podURL, grpc.WithInsecure(), grpc.WithBlock())
	s.Require().NoError(err)
	return podURL, podClient
}

func TestRunNsmgrSuite(t *testing.T) {
	suite.Run(t, &NsmgrTestsSuite{prefix: "nsmgr"})
}
