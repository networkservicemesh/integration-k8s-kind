// Copyright (c) 2024 Nordix and/or its affiliates.
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

package ovsextra_test

import (
	"flag"
	"testing"

	"github.com/networkservicemesh/integration-tests/extensions/parallel"
	"github.com/networkservicemesh/integration-tests/suites/features_ovs"
)

var smartVFFlag = flag.Bool("smart", false, "selects smartVF tests")

// Disabled tests for kind:
// SmartVF to SmartVF Connection - ../features/webhook-smartvf
type kindFeatOvsSuite struct {
	features_ovs.Suite
}

func (s *kindFeatOvsSuite) BeforeTest(suiteName, testName string) {
	if testName == "TestWebhook_smartvf" {
		s.T().Skip()
	}
}

func TestRunFeatureOvsSuite(t *testing.T) {
	if !*smartVFFlag {
		parallel.Run(t, new(kindFeatOvsSuite), "TestScale_from_zero", "TestNse_composition", "TestSelect_forwarder")
	} else {
		parallel.Run(t, new(features_ovs.Suite), "TestScale_from_zero", "TestNse_composition", "TestSelect_forwarder")
	}
}
