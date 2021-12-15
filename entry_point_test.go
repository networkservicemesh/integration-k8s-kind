// Copyright (c) 2021 Doc.ai and/or its affiliates.
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

package test

import (
	"io/ioutil"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/networkservicemesh/integration-tests/suites/basic"
	"github.com/networkservicemesh/integration-tests/suites/features"
	"github.com/networkservicemesh/integration-tests/suites/heal"
	"github.com/networkservicemesh/integration-tests/suites/memory"
	"github.com/networkservicemesh/integration-tests/suites/scalability"
)

func TestRunHealSuite(t *testing.T) {
	suite.Run(t, new(heal.Suite))
}

func TestRunFeatureSuite(t *testing.T) {
	suite.Run(t, new(features.Suite))
}

func TestRunBasicSuite(t *testing.T) {
	suite.Run(t, new(basic.Suite))
}

func TestRunMemorySuite(t *testing.T) {
	suite.Run(t, new(memory.Suite))
}

func setScalabilityTestParams(netsvc, nse, nsc int, remote bool) string {
	filename := "../deployments-k8s/examples/scalability/cases/set_params.sh"
	filecontent := `#!/bin/bash
TEST_NS_COUNT=` + strconv.Itoa(netsvc) + `
TEST_NSE_COUNT=` + strconv.Itoa(nse) + `
TEST_NSC_COUNT=` + strconv.Itoa(nsc) + `
TEST_REMOTE_CASE=` + strconv.FormatBool(remote)
	err := ioutil.WriteFile(filename, []byte(filecontent), 0600)
	if err != nil {
		panic(err)
	}
	testName := "ns=" + strconv.Itoa(netsvc) + ",nse=" + strconv.Itoa(nse) + ",nsc=" + strconv.Itoa(nsc) + ",remote=" + strconv.FormatBool(remote)
	return testName
}

func TestRunScalabilitySuite(t *testing.T) {
	t.Run("ns=1,nse=1,nsc=1,remote=false", func(t *testing.T) { suite.Run(t, new(scalability.Suite)) })
	t.Run(setScalabilityTestParams(1, 1, 5, true), func(t *testing.T) { suite.Run(t, new(scalability.Suite)) })
}
