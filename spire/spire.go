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

// Package spire provides spiffe/spire helper functions
package spire

import (
	"github.com/edwarnicke/exechelper"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/networkservicemesh/integration-k8s-kind/k8s"
)

// Config - configuration for spire
type Config struct {
	Cleanup   bool   `default:"true" desc:"Perform full Spire cleanup" split_words:"true"`
	Namespace string `default:"spire" desc:"Namespace of spire namespace" split_words:"true"`
}

// Delete deletes spire namespace with all depended resources
func Delete(options ...*exechelper.Option) error {
	if config().Cleanup {
		return exechelper.Run("kubectl delete -f ./deployments/spire/spire-namespace.yaml", options...)
	}
	return nil
}

// Setup setups spire in spire namespace
func Setup(options ...*exechelper.Option) error {
	if k8s.NamespaceExists(config().Namespace) {
		return nil
	}
	if err := exechelper.Run("kubectl apply -f ./deployments/spire/spire-namespace.yaml", options...); err != nil {
		return errors.Wrap(err, "cannot create spire namespace")
	}
	if err := exechelper.Run("kubectl apply -f ./deployments/spire", options...); err != nil {
		return errors.Wrap(err, "cannot deploy spire deployments")
	}
	if err := exechelper.Run("kubectl wait -n spire --timeout=60s --for=condition=ready pod -l app=spire-agent", options...); err != nil {
		return errors.Wrap(err, "spire-agent cannot start")
	}
	if err := exechelper.Run("kubectl wait -n spire --timeout=60s --for=condition=ready pod -l app=spire-server", options...); err != nil {
		return errors.Wrap(err, "spire-server cannot start")
	}
	if err := exechelper.Run(`kubectl exec -n spire spire-server-0 --
                                      /opt/spire/bin/spire-server entry create
                                      -spiffeID spiffe://example.org/ns/spire/sa/spire-agent
                                      -selector k8s_sat:cluster:nsm-cluster
                                      -selector k8s_sat:agent_ns:spire
                                      -selector k8s_sat:agent_sa:spire-agent
                                      -node`, options...); err != nil {
		return errors.Wrap(err, "cannot create spire-entry for spire-agent")
	}
	if err := exechelper.Run(`kubectl exec -n spire spire-server-0 --
                                      /opt/spire/bin/spire-server entry create
                                      -spiffeID spiffe://example.org/ns/default/sa/default
                                      -parentID spiffe://example.org/ns/spire/sa/spire-agent
                                      -selector k8s:ns:default  
                                      -selector k8s:sa:default`, options...); err != nil {
		return errors.Wrap(err, "cannot create spire-entry for default namespace")
	}
	if err := exechelper.Run(`kubectl exec -n spire spire-server-0 --
                                      /opt/spire/bin/spire-server entry create
                                      -spiffeID spiffe://example.org/ns/default/sa/default
                                      -parentID spiffe://example.org/ns/spire/sa/spire-agent
                                      -selector k8s:ns:nsmgr 
                                      -selector k8s:sa:default`, options...); err != nil {
		return errors.Wrap(err, "cannot create spire-entry for default namespace")
	}
	return nil
}

// config - read configuration from environment.
func config() *Config {
	conf := &Config{}
	if err := envconfig.Process("spire", conf); err != nil {
		logrus.Fatalf("error processing conf from env: %+v", err)
	}
	return conf
}
