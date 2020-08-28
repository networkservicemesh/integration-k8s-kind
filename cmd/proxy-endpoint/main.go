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

// Package main define a nsc application
package main

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/networkservicemesh/api/pkg/api/networkservice"
	api_registry "github.com/networkservicemesh/api/pkg/api/registry"
	"github.com/networkservicemesh/sdk/pkg/networkservice/chains/registry"
	"github.com/networkservicemesh/sdk/pkg/registry/core/adapters"
	"github.com/networkservicemesh/sdk/pkg/registry/core/chain"
	"github.com/networkservicemesh/sdk/pkg/registry/core/nextwrap"
	"github.com/networkservicemesh/sdk/pkg/tools/callback"
	"github.com/networkservicemesh/sdk/pkg/tools/debug"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"net/url"
	"os"
	"time"

	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/svid/x509svid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/kelseyhightower/envconfig"
	"github.com/networkservicemesh/sdk/pkg/tools/grpcutils"
	"github.com/sirupsen/logrus"

	"github.com/networkservicemesh/sdk/pkg/tools/jaeger"
	"github.com/networkservicemesh/sdk/pkg/tools/log"
	"github.com/networkservicemesh/sdk/pkg/tools/signalctx"
	"github.com/networkservicemesh/sdk/pkg/tools/spanhelper"
)

// Config - configuration for cmd-nsmgr
type Config struct {
	Name             string        `default:"proxy-endpoint" desc:"Namespace of Network service "`
	ListenOn         url.URL       `default:"tcp://127.0.0.1:6001" desc:"Network Service Endpoint Listen point" split_words:"true"`
	ConnectTo        url.URL       `default:"unix:///var/lib/networkservicemesh/nsm.io.sock" desc:"url to connect to NSM" split_words:"true"`
	MaxTokenLifetime time.Duration `default:"24h" desc:"maximum lifetime of tokens" split_words:"true"`
}

func main() {
	// ********************************************************************************
	// Configure signal handling context
	// ********************************************************************************
	ctx := signalctx.WithSignals(context.Background())
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	// ********************************************************************************
	// Debug self if necessary
	// ********************************************************************************
	if err := debug.Self(); err != nil {
		log.Entry(ctx).Infof("%s", err)
	}

	// ********************************************************************************
	// Setup logger
	// ********************************************************************************
	logrus.Info("Starting NetworkServiceMesh Proxy Endpoint ...")
	logrus.SetFormatter(&nested.Formatter{})
	logrus.SetLevel(logrus.TraceLevel)

	ctx = log.WithField(ctx, "cmd", os.Args[:1])

	// ********************************************************************************
	// Configure open tracing
	// ********************************************************************************
	var span opentracing.Span
	// Enable Jaeger
	if jaeger.IsOpentracingEnabled() {
		jaegerCloser := jaeger.InitJaeger("proxy-endpoint")
		defer func() { _ = jaegerCloser.Close() }()
		span = opentracing.StartSpan("proxy-endpoint")
	}
	cmdSpan := spanhelper.NewSpanHelper(ctx, span, "proxy-endpoint")

	// ********************************************************************************
	// Get config from environment
	// ********************************************************************************
	rootConf := &Config{}
	if err := envconfig.Usage("nsm", rootConf); err != nil {
		logrus.Fatal(err)
	}
	if err := envconfig.Process("nsm", rootConf); err != nil {
		logrus.Fatalf("error processing rootConf from env: %+v", err)
	}

	go func() {
		logrus.Infof("starting endpoint #2...")
		e := StartNSMProxyEndpoint(ctx, rootConf)

		if e != nil {
			logrus.Fatalf("exit due %v", e)
		}
	}()

	// Startup is finished
	cmdSpan.Finish()

	logrus.Infof("Proxy endpoint started...")
	// Wait for cancel event to terminate
	<-ctx.Done()
}

// IdentityByEndpointID - return identity by :endpoint-id
func IdentityByEndpointID(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		err := errors.New("no metadata provided")
		logrus.Error(err)
		return "", err
	}
	return md.Get("endpoint-id")[0], nil
}

// WithCallbackEndpointID - pass with :endpoint-id a correct endpoint identity.
func WithCallbackEndpointID(ctx context.Context, endpoint string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "endpoint-id", endpoint)
}

type proxyEndpoint struct {
	registry.Registry
	nsmClient        networkservice.NetworkServiceClient
	nsmMonitorClient networkservice.MonitorConnectionClient
}

func (p *proxyEndpoint) MonitorConnections(selector *networkservice.MonitorScopeSelector, server networkservice.MonitorConnection_MonitorConnectionsServer) error {
	client, err := p.nsmMonitorClient.MonitorConnections(server.Context(), selector)
	if err != nil {
		return err
	}
	for {
		select {
		case <-server.Context().Done():
			return server.Context().Err()
		default:
			{
				evt, recverr := client.Recv()
				if recverr != nil {
					return recverr
				}
				sendErr := server.Send(evt)
				if sendErr != nil {
					return sendErr
				}
			}
		}
	}
}

func (p *proxyEndpoint) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	// Performing nsmClient connection request
	conn, connerr := p.nsmClient.Request(ctx, request)
	if connerr != nil {
		logrus.Errorf("Failed to request network service with %v: err %v", request, connerr)
		return conn, connerr
	}

	logrus.Infof("Network service established with %v\n Connection:%v", request, conn)
	return conn, connerr
}

func (p *proxyEndpoint) Close(ctx context.Context, connection *networkservice.Connection) (*empty.Empty, error) {
	return p.nsmClient.Close(ctx, connection)
}

// NewNSMProxyClient - creates a client connection to NSMGr
func StartNSMProxyEndpoint(ctx context.Context, rootConf *Config) error {
	// ********************************************************************************
	// Get a x509Source
	// ********************************************************************************
	logrus.Infof("Retriving Spire Source")
	source, err := workloadapi.NewX509Source(ctx)
	if err != nil {
		logrus.Fatalf("error getting x509 source: %+v", err)
	}
	var svid *x509svid.SVID
	svid, err = source.GetX509SVID()
	if err != nil {
		logrus.Fatalf("error getting x509 svid: %+v", err)
	}
	logrus.Infof("sVID: %q", svid.ID)

	// ********************************************************************************
	// Connect to NSManager
	// ********************************************************************************
	logrus.Infof("Proxy Endpoint: Connecting to Network Service Manager %v", rootConf.ConnectTo.String())
	var clientCC *grpc.ClientConn
	clientCC, err = grpc.DialContext(ctx,
		grpcutils.URLToTarget(&rootConf.ConnectTo),
		grpc.WithTransportCredentials(
			credentials.NewTLS(
				tlsconfig.MTLSClientConfig(source, source, tlsconfig.AuthorizeAny()))),
		grpc.WithDefaultCallOptions(
			grpc.WaitForReady(true)))
	if err != nil {
		e := errors.Errorf("failed to dial NSM: %v", err)
		return e
	}

	grpclog.SetLoggerV2(grpclog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout))

	//****
	logrus.Infof("Checking NSMgr is responding")

	registryCC := api_registry.NewNetworkServiceEndpointRegistryClient(clientCC)

	stream, serr := registryCC.Find(context.Background(), &api_registry.NetworkServiceEndpointQuery{
		NetworkServiceEndpoint: &api_registry.NetworkServiceEndpoint{

		},
		Watch: false,
	})
	if serr != nil {
		logrus.Fatal("Failed to query NSmgr")
	}
	logrus.Infof("Endpoints found: %v", api_registry.ReadNetworkServiceEndpointList(stream))

	ep := &proxyEndpoint{
		nsmClient:        networkservice.NewNetworkServiceClient(clientCC),
		nsmMonitorClient: networkservice.NewMonitorConnectionClient(clientCC),
	}
	// ************************************************
	// Setup registry
	// ************************************************
	nsRegistry := adapters.NetworkServiceClientToServer(
		nextwrap.NewNetworkServiceRegistryClient(
			api_registry.NewNetworkServiceRegistryClient(clientCC)))

	nsmgrRegServer := adapters.NetworkServiceEndpointClientToServer(
		nextwrap.NewNetworkServiceEndpointRegistryClient(
			api_registry.NewNetworkServiceEndpointRegistryClient(clientCC)))

	callbackServer := &callbackRegistryServer{
		clientCC: clientCC,
		server:   callback.NewServer(IdentityByEndpointID),
		servers:  map[string]*callbackClientEntry{},
		source:   source,
	}

	nseRegistry := chain.NewNetworkServiceEndpointRegistryServer(callbackServer, nsmgrRegServer)

	// ************************************************
	// Construct endpoint registry
	// ************************************************
	ep.Registry = registry.NewServer(nsRegistry, nseRegistry)

	// ************************************************
	// Construct GRPC server with no security
	// ************************************************
	server := grpc.NewServer()

	// Construct context to pass identity to server.
	ep.Registry.Register(server)
	networkservice.RegisterNetworkServiceServer(server, ep)
	networkservice.RegisterMonitorConnectionServer(server, ep)

	// Handle callback registration
	callback.RegisterCallbackServiceServer(server, callbackServer.server)

	// Listen to handle new endpoint connections.
	_ = grpcutils.ListenAndServe(ctx, &rootConf.ListenOn, server)
	logrus.Infof("Endpoint is serving at %v", rootConf.ListenOn)
	return nil
}
