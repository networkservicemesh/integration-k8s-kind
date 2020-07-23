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

package integration_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/networkservicemesh/integration-k8s-kind/tests/integration/nsmtesting"
)

type BasicTestsSuite struct {
	suite.Suite
	*nsmtesting.NSMTesting
}

func (s *BasicTestsSuite) SetupSuite() {
	s.NSMTesting = nsmtesting.New(s.T())
}

func (s *BasicTestsSuite) TestDeployAlpine() {
	s.Exec("kubectl apply -f ../deployments/alpine.yaml")
	defer s.Exec("kubectl delete -f ../deployments/alpine.yaml")
	s.Exec("kubectl wait --for=condition=ready pod -l app=alpine")
}

func TestBasic(t *testing.T) {
	suite.Run(t, &BasicTestsSuite{})
}
