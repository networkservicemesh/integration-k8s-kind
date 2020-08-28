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
	"github.com/networkservicemesh/sdk/pkg/networkservice/common/updatetoken"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/chain"
	"github.com/networkservicemesh/sdk/pkg/tools/debug"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/svid/x509svid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/api/pkg/api/networkservice/mechanisms/kernel"
	"github.com/networkservicemesh/sdk/pkg/tools/fs"
	"github.com/networkservicemesh/sdk/pkg/tools/grpcutils"
	"github.com/networkservicemesh/sdk/pkg/tools/spiffejwt"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"

	"github.com/networkservicemesh/sdk/pkg/tools/jaeger"
	"github.com/networkservicemesh/sdk/pkg/tools/log"
	"github.com/networkservicemesh/sdk/pkg/tools/signalctx"
	"github.com/networkservicemesh/sdk/pkg/tools/spanhelper"
)

// Config - configuration for cmd-nsmgr
type Config struct {
	Name             string        `default:"proxy-nsc" desc:"Namespace of Network service "`
	ListenOn         url.URL       `default:"tcp://127.0.0.1:6001" desc:"Namespace of Network Service Client" split_words:"true"`
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
	logrus.Info("Starting NetworkServiceMesh Proxy Client ...")
	logrus.SetFormatter(&nested.Formatter{})
	logrus.SetLevel(logrus.TraceLevel)

	ctx = log.WithField(ctx, "cmd", os.Args[:1])

	// ********************************************************************************
	// Configure open tracing
	// ********************************************************************************
	var span opentracing.Span
	// Enable Jaeger
	if jaeger.IsOpentracingEnabled() {
		jaegerCloser := jaeger.InitJaeger("nsc")
		defer func() { _ = jaegerCloser.Close() }()
		span = opentracing.StartSpan("nsc")
	}
	cmdSpan := spanhelper.NewSpanHelper(ctx, span, "nsc")

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

	nsmClient, clientCC, e := NewNSMProxyClient(ctx, rootConf)
	if e != nil {
		logrus.Fatalf("exit due %v", e)
	}
	errCh := RunProxyClient(cmdSpan.Context(), rootConf, nsmClient, clientCC)

	// Startup is finished
	cmdSpan.Finish()

	// Wait for cancel event to terminate
	select {
	case <-ctx.Done():
	case <-errCh:
	}
}

// NewNSMProxyClient - creates a client connection to NSMGr
func NewNSMProxyClient(ctx context.Context, rootConf *Config) (networkservice.NetworkServiceClient, *grpc.ClientConn, error) {
	// ********************************************************************************
	// Get a x509Source
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
	logrus.Infof("sVID: %q", svid.ID)

	// ********************************************************************************
	// Connect to NSManager
	// ********************************************************************************
	logrus.Infof("NSC: Connecting to Network Service Manager %v", rootConf.ConnectTo.String())
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
		return nil, nil, e
	}

	// ********************************************************************************
	// Create Network Service Manager nsmClient
	// ********************************************************************************
	// We need to update path and send it to nsmgr
	return chain.NewNetworkServiceClient(updatetoken.NewClient(spiffejwt.TokenGeneratorFunc(source, rootConf.MaxTokenLifetime)), networkservice.NewNetworkServiceClient(clientCC)), clientCC, nil
}

type proxyEndpointServerImpl struct {
	nsmClient networkservice.NetworkServiceClient
	clientCC  *grpc.ClientConn
}

func (p *proxyEndpointServerImpl) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	for _, m := range request.MechanismPreferences {
		if m.Type == kernel.MECHANISM {
			logrus.Infof("%v", runtime.GOOS)
			if runtime.GOOS != "darwin" && runtime.GOOS != "windows" {
				// Check we are not macos or windows.
				inode, err := fs.GetInode("/proc/self/ns/net")
				if err != nil {
					logrus.Errorf("could not retrieve a linux namespace %v", err)
					return nil, err
				}
				if m.Parameters == nil {
					m.Parameters = map[string]string{}
				}
				m.Parameters[kernel.NetNSInodeKey] = strconv.FormatUint(uint64(inode), 10)
			}

			kernel.ToMechanism(m).SetNetNSURL("unix:///proc/self/ns/net")
		}
	}

	// Performing nsmClient connection request
	conn, connerr := p.nsmClient.Request(ctx, request)
	if connerr != nil {
		logrus.Errorf("Failed to request network service with %v: err %v", request, connerr)
		return conn, connerr
	}

	logrus.Infof("Network service established with %v\n Connection:%v", request, conn)
	return conn, connerr
}

func (p *proxyEndpointServerImpl) Close(ctx context.Context, connection *networkservice.Connection) (*empty.Empty, error) {
	return p.nsmClient.Close(ctx, connection)
}

func (p *proxyEndpointServerImpl) MonitorConnections(selector *networkservice.MonitorScopeSelector, server networkservice.MonitorConnection_MonitorConnectionsServer) error {
	nsmMonitorClient := networkservice.NewMonitorConnectionClient(p.clientCC)
	monClient, err := nsmMonitorClient.MonitorConnections(server.Context(), selector)
	if err != nil {
		return err
	}
	for {
		select {
		case <-server.Context().Done():
			return server.Context().Err()
		default:
			{
				evt, recverr := monClient.Recv()
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

// RunProxyClient - runs a client application with passed configuration over a client to Network Service Manager
func RunProxyClient(ctx context.Context, rootConf *Config, nsmClient networkservice.NetworkServiceClient, clientCC *grpc.ClientConn) <-chan error {
	// ********************************************************************************
	// Initiate connections
	// ********************************************************************************

	var serviceServer = &proxyEndpointServerImpl{
		clientCC:  clientCC,
		nsmClient: nsmClient,
	}
	server := grpc.NewServer()
	networkservice.RegisterNetworkServiceServer(server, serviceServer)
	networkservice.RegisterMonitorConnectionServer(server, serviceServer)

	return grpcutils.ListenAndServe(ctx, &rootConf.ListenOn, server)

}
