// Copyright (c) 2021-2022 Doc.ai and/or its affiliates.
//
// Copyright (c) 2022 Cisco and/or its affiliates.
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
	"os"
	"testing"

	"github.com/networkservicemesh/integration-tests/extensions/parallel"
	"github.com/networkservicemesh/integration-tests/suites/multicluster"
)

func TestRunMulticlusterSuite(t *testing.T) {
	os.Setenv("KUBECONFIG1", "/tmp/config1")
	os.Setenv("KUBECONFIG2", "/tmp/config2")
	os.Setenv("KUBECONFIG3", "/tmp/config3")
	os.Setenv("CLUSTER1_CIDR", "172.18.1.128/25")
	os.Setenv("CLUSTER2_CIDR", "172.18.2.128/25")
	os.Setenv("CLUSTER3_CIDR", "172.18.3.128/25")
	parallel.Run(t, new(multicluster.Suite))
}
