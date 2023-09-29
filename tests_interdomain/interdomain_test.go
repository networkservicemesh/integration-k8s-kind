// Copyright (c) 2021-2022 Doc.ai and/or its affiliates.
//
// Copyright (c) 2022-2023 Cisco and/or its affiliates.
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

	"github.com/networkservicemesh/integration-tests/extensions/parallel"
	"github.com/networkservicemesh/integration-tests/suites/multicluster"
)

func TestRunMulticlusterSuite(t *testing.T) {
	parallel.Run(t, new(multicluster.Suite),
		"TestFloating_vl3_basic",
		"TestFloating_vl3_scale_from_zero",
		"TestFloating_vl3_dns",
		"TestFloating_nse_composition",
	)
}

//func TestRunBasicInterdomainSuite(t *testing.T) {
//	suite.Run(t, new(interdomain.Suite))
//}
//
//func TestRunMulticlusterHealSuite(t *testing.T) {
//	suite.Run(t, new(multicluster_heal.Suite))
//}
