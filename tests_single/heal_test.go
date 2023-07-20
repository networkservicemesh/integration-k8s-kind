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

package single

import (
	"testing"

	"github.com/networkservicemesh/integration-tests/extensions/parallel"
	"github.com/networkservicemesh/integration-tests/suites/heal"
)

func TestRunHealSuite(t *testing.T) {
	parallel.Run(t, new(heal.Suite),
		"TestLocal_forwarder_remote_forwarder",
		"TestLocal_nsm_system_restart",
		"TestLocal_nsmgr_local_forwarder_memif",
		"TestLocal_nsmgr_remote_nsmgr",
		"TestLocal_nsmgr_restart",
		"TestRegistry_remote_forwarder",
		"TestRegistry_remote_nsmgr",
		"TestRegistry_restart",
		"TestRemote_forwarder_death",
		"TestRemote_forwarder_death_ip",
		"TestRemote_nsm_system_restart_memif_ip",
		"TestRemote_nsmgr_death",
		"TestRemote_nsmgr_remote_endpoint",
		"TestRemote_nsmgr_restart",
		"TestRemote_nsmgr_restart_ip",
		"TestSpire_server_agent_restart",
		"TestSpire_server_restart",
		"TestSpire_upgrade",
		"TestVl3_nscs_death",
		"TestVl3_nse_death",
		"TestLocal_nsmgr_local_nse_memif",
		"TestRegistry_local_endpoint")
}
