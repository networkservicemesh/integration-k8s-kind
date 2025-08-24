// Copyright (c) 2023-2025 Nordix Foundation.
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

package single

import (
	"flag"
	"strings"
	"testing"

	"github.com/networkservicemesh/integration-tests/extensions/parallel"
	"github.com/networkservicemesh/integration-tests/suites/ovs"
)

var smartVFFlag = flag.Bool("smart", false, "selects smartVF tests")

// Disabled tests:
// SmartVF to SmartVF Connection - ../use-cases/SmartVF2SmartVF
// Temporary disabled tests:
// Kernel to Kernel Connection over VLAN Trunking - ../use-cases/Kernel2KernelVLAN
type kindOvsSuite struct {
	ovs.Suite
}

func (s *kindOvsSuite) BeforeTest(suiteName, testName string) {
	switch strings.ToLower(testName) {
	case
		"testsmartvf2smartvf",
		"testkernel2kernelvlan":
		s.T().Skip()
	}
}

func TestRunOvsSuite(t *testing.T) {
	if !*smartVFFlag {
		parallel.Run(t, new(kindOvsSuite))
	} else {
		parallel.Run(t, new(ovs.Suite))
	}
}
