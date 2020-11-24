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
	"testing"
	"time"

	"github.com/networkservicemesh/integration-k8s-kind/k8s"
	"github.com/networkservicemesh/integration-k8s-kind/k8s/require"
	"github.com/networkservicemesh/integration-k8s-kind/spire"

	"github.com/edwarnicke/exechelper"
	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/suite"
)

const defaultNamespace = "default"

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

	s.Require().NoError(spire.Setup(s.options...))

	// Setup NSM
	s.Require().NoError(exechelper.Run("kubectl apply -f ./deployments/namespace.yaml", s.options...))

	s.Require().NoError(exechelper.Run("kubectl apply -f ./deployments/registry-service.yaml", s.options...))
	s.Require().NoError(exechelper.Run("kubectl apply -f ./deployments/registry-memory.yaml", s.options...))
	s.Require().NoError(exechelper.Run("kubectl wait --for=condition=ready pod -l app=nsm-registry --namespace nsm-system", s.options...))

	s.Require().NoError(exechelper.Run("kubectl apply -f ./deployments/nsmgr.yaml", s.options...))
	s.Require().NoError(exechelper.Run("kubectl wait --for=condition=ready pod -l app=nsmgr --namespace nsm-system", s.options...))

	s.Require().NoError(exechelper.Run("kubectl apply -f ./deployments/fake-cross-nse.yaml", s.options...))
	s.Require().NoError(exechelper.Run("kubectl wait --for=condition=ready pod -l app=fake-cross-nse --namespace nsm-system", s.options...))
}

func (s *BasicTestsSuite) TearDownSuite() {
	s.Require().NoError(spire.Delete(s.options...))
	s.Require().NoError(exechelper.Run("kubectl delete -f ./deployments/namespace.yaml", s.options...))
}

func (s *BasicTestsSuite) TearDownTest() {
	k8s.ShowLogs(defaultNamespace, s.options...)
}

func (s *BasicTestsSuite) TestDeployAlpine() {
	defer require.NoRestarts(s.T())

	s.Require().NoError(exechelper.Run("kubectl apply -f ./deployments/alpine.yaml", s.options...))
	s.Require().NoError(exechelper.Run("kubectl wait --for=condition=ready pod -l app=alpine", s.options...))
}

func (s *BasicTestsSuite) TestNSM_Local() {
	defer require.NoRestarts(s.T())

	ns, cleanup, err := k8s.NewNamespace()
	s.Require().NoError(err)
	defer cleanup()
	s.Require().NoError(spire.RegisterNamespace(ns, s.options...))

	nodes, err := k8s.Nodes()
	s.Require().NoError(err)
	s.Require().Greater(len(nodes), 0)

	s.Require().NoError(k8s.ApplyDeployment("./deployments/nse.yaml", k8s.SetNode(nodes[0].Labels["kubernetes.io/hostname"]), k8s.SetNamespace(ns)))
	s.Require().NoError(exechelper.Run("kubectl wait --for=condition=ready pod -l app=nse -n"+ns, s.options...))
	s.Require().NoError(k8s.ApplyDeployment("./deployments/nsc.yaml", k8s.SetNode(nodes[0].Labels["kubernetes.io/hostname"]), k8s.SetNamespace(ns)))
	s.Require().NoError(exechelper.Run("kubectl wait --for=condition=ready pod -l app=nsc -n"+ns, s.options...))

	time.Sleep(time.Second * 15) // https://github.com/networkservicemesh/sdk/issues/593

	s.Require().NoError(k8s.WaitLogsMatch("app=nsc", "All client init operations are done.", ns, time.Minute/2))
}

func (s *BasicTestsSuite) TestNSM_Remote() {
	defer require.NoRestarts(s.T())

	ns, cleanup, err := k8s.NewNamespace()
	s.Require().NoError(err)
	defer cleanup()
	s.Require().NoError(spire.RegisterNamespace(ns, s.options...))

	nodes, err := k8s.Nodes()
	s.Require().NoError(err)
	s.Require().Greater(len(nodes), 1)

	s.Require().NoError(k8s.ApplyDeployment("./deployments/nse.yaml", k8s.SetNode(nodes[0].Labels["kubernetes.io/hostname"]), k8s.SetNamespace(ns)))
	s.Require().NoError(exechelper.Run("kubectl wait --for=condition=ready pod -l app=nse -n"+ns, s.options...))
	s.Require().NoError(k8s.ApplyDeployment("./deployments/nsc.yaml", k8s.SetNode(nodes[1].Labels["kubernetes.io/hostname"]), k8s.SetNamespace(ns)))
	s.Require().NoError(exechelper.Run("kubectl wait --for=condition=ready pod -l app=nsc -n"+ns, s.options...))

	time.Sleep(time.Second * 15) // https://github.com/networkservicemesh/sdk/issues/593

	s.Require().NoError(k8s.WaitLogsMatch("app=nsc", "All client init operations are done.", ns, time.Minute/2))
}

func TestRunBasicSuite(t *testing.T) {
	suite.Run(t, &BasicTestsSuite{})
}
