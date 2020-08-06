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
	"os"
	"path/filepath"
	"testing"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/edwarnicke/exechelper"
	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/suite"
	"k8s.io/client-go/kubernetes"
)

type BasicTestsSuite struct {
	suite.Suite
	options []*exechelper.Option
}

func (s *BasicTestsSuite) SetupSuite() {
	writer := logrus.StandardLogger().Writer()

	s.options = []*exechelper.Option{
		exechelper.WithStderr(writer),
		exechelper.WithStdout(writer),
	}

	s.Require().NoError(exechelper.Run("kubectl apply -f ./deployments/spire/spire-namespace.yaml", s.options...))
	s.Require().NoError(exechelper.Run("kubectl apply -f ./deployments/spire", s.options...))

	s.Require().NoError(exechelper.Run("kubectl wait -n spire --timeout=60s --for=condition=ready pod -l app=spire-agent", s.options...))
	s.Require().NoError(exechelper.Run("kubectl wait -n spire --timeout=60s --for=condition=ready pod -l app=spire-server", s.options...))

	s.Require().NoError(exechelper.Run(`kubectl exec -n spire spire-server-0 --
												/opt/spire/bin/spire-server entry create
												-spiffeID spiffe://example.org/ns/spire/sa/spire-agent
												-selector k8s_sat:cluster:nsm-cluster
												-selector k8s_sat:agent_ns:spire
												-selector k8s_sat:agent_sa:spire-agent
												-node`, s.options...))

	s.Require().NoError(exechelper.Run(`kubectl exec -n spire spire-server-0 --
												/opt/spire/bin/spire-server entry create
												-spiffeID spiffe://example.org/ns/default/sa/default
												-parentID spiffe://example.org/ns/spire/sa/spire-agent
												-selector k8s:ns:default \
												-selector k8s:sa:default`, s.options...))
}

func (s *BasicTestsSuite) TearDownSuite() {
	s.Require().NoError(exechelper.Run("kubectl delete -f ./deployments/spire/spire-namespace.yaml"))
}

func (s *BasicTestsSuite) TearDownTest() {
	s.Require().NoError(exechelper.Run("kubectl delete serviceaccounts --all"))
	s.Require().NoError(exechelper.Run("kubectl delete services --all"))
	s.Require().NoError(exechelper.Run("kubectl delete deployment --all"))
	s.Require().NoError(exechelper.Run("kubectl delete pods --all"))
}

func (s *BasicTestsSuite) TestDeployMemoryRegistry() {
	s.Require().NoError(exechelper.Run("kubectl apply -f ./deployments/memory-registry.yaml", s.options...))
	defer func() {
		s.Require().NoError(exechelper.Run("kubectl describe pod -l app=memory-registry", s.options...))
	}()
	s.Require().NoError(exechelper.Run("kubectl wait --timeout=120s  --for=condition=ready pod -l app=memory-registry", s.options...))
	s.Require().NoError(exechelper.Run("kubectl delete -f ./deployments/memory-registry.yaml", s.options...))
}

func (s *BasicTestsSuite) TestK8sClient() {
	path := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", path)
	s.NoError(err)
	_, err = kubernetes.NewForConfig(config)
	s.NoError(err)
}

func (s *BasicTestsSuite) TestDeployAlpine() {
	s.Require().NoError(exechelper.Run("kubectl apply -f ./deployments/alpine.yaml", s.options...))
	s.Require().NoError(exechelper.Run("kubectl wait --for=condition=ready pod -l app=alpine", s.options...))
	s.Require().NoError(exechelper.Run("kubectl delete -f ./deployments/alpine.yaml", s.options...))
}

func TestRunBasicSuite(t *testing.T) {
	suite.Run(t, &BasicTestsSuite{})
}
