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

// Package nsmgr - allow to start nsmgr
package nsmgr

import (
	"github.com/edwarnicke/exechelper"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/networkservicemesh/integration-k8s-kind/k8s"
)

// Config - configuration for cmd-nsmgr
type Config struct {
	Cleanup   bool   `default:"true" desc:"Perform full NSMGR cleanup" split_words:"true"`
	Namespace string `default:"nsmgr" desc:"Namespace of nsmgr container" split_words:"true"`
}

// Delete deletes nsmgr namespace with all depended resources
func Delete(options ...*exechelper.Option) error {
	if config().Cleanup {
		if err := exechelper.Run("kubectl delete -f ./deployments/nsmgr/nsmgr.yaml --wait --force --grace-period=0", options...); err != nil {
			return errors.Wrap(err, "cannot delete nsmgr deployments")
		}
		return exechelper.Run("kubectl delete -f ./deployments/nsmgr/nsmgr-namespace.yaml --wait", options...)
	}
	return nil
}

// Setup setups spire in spire namespace
func Setup(options ...*exechelper.Option) error {
	if !k8s.NamespaceExists(config().Namespace) {
		if err := exechelper.Run("kubectl apply -f ./deployments/nsmgr/nsmgr-namespace.yaml", options...); err != nil {
			return errors.Wrap(err, "cannot create spire namespace")
		}
	}

	if err := exechelper.Run("kubectl apply -f ./deployments/nsmgr", options...); err != nil {
		return errors.Wrap(err, "cannot deploy spire deployments")
	}
	return nil
}

// config - read configuration from environment.
func config() *Config {
	conf := &Config{}
	if err := envconfig.Process("nsmgr", conf); err != nil {
		logrus.Fatalf("error processing conf from env: %+v", err)
	}
	return conf
}
