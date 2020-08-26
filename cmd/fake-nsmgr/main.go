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
	ConnectTo                      url.URL `desc:"url to the local registry that handles this domain" split_words:"true"`
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
	if err := envconfig.Usage("fake-nsmgr", config); err != nil {
		logrus.Fatal(err)
	}
	if err := envconfig.Process("fake-nsmgr", config); err != nil {
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
		config.ConnectTo.String(),
		grpc.WithTransportCredentials(credentials.NewTLS(tlsconfig.MTLSClientConfig(source, source, tlsconfig.AuthorizeAny()))),
		grpc.WithBlock(),
	)
	if err != nil {
		logrus.Fatal(err.Error())
	}

	logrus.Info("Starting search")

	nseStream, err := api_registry.NewNetworkServiceEndpointRegistryClient(cc).Find(context.Background(), &api_registry.NetworkServiceEndpointQuery{
		NetworkServiceEndpoint: &api_registry.NetworkServiceEndpoint{Name: config.FindNetworkServiceEndpointName},
	})

	if err != nil {
		logrus.Fatal(err.Error())
	}
	nses := api_registry.ReadNetworkServiceEndpointList(nseStream)
	if len(nses) == 0 {
		logrus.Fatal("Network Service  Endpoint is not found")
	}
	logrus.Infof("Found an NSE: %+v", nses[0])
	<-ctx.Done()
}
