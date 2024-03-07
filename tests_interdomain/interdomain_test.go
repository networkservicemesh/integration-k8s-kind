// Copyright (c) 2021-2022 Doc.ai and/or its affiliates.
//
// Copyright (c) 2022-2024 Cisco and/or its affiliates.
//
// Copyright (c) 2024 Pragmagic Inc. and/or its affiliates.
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

package interdomain

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/networkservicemesh/integration-tests/extensions/parallel"
	"github.com/networkservicemesh/integration-tests/suites/interdomain/suites/basic"
	"github.com/networkservicemesh/integration-tests/suites/interdomain/suites/heal"
	"github.com/networkservicemesh/integration-tests/suites/interdomain/suites/ipsec"
	"github.com/networkservicemesh/integration-tests/suites/interdomain/suites/multiservicemesh"
)

func TestRunBasicInterdomainSuite(t *testing.T) {
	excludedTests := []string{
		"TestFloating_vl3_basic",
		"TestFloating_vl3_scale_from_zero",
		"TestFloating_vl3_dns",
		"TestFloating_nse_composition"}

	parallel.Run(t, new(basic.Suite), parallel.WithExcludedTests(excludedTests))
}

func TestRunInterdomainIPSecSuite(t *testing.T) {
	parallel.Run(t, new(ipsec.Suite))
}

func TestRunInterdomainHealSuite(t *testing.T) {
	suite.Run(t, new(heal.Suite))
}

func TestRunMultiServiceMeshSuite(t *testing.T) {
	suite.Run(t, new(multiservicemesh.Suite))
}
