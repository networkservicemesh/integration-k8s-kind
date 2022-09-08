// Copyright (c) 2021-2022 Doc.ai and/or its affiliates.
// Copyright (c) 2022 Nordix and/or its affiliates.
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

package main

import (
	"testing"

	"github.com/networkservicemesh/integration-tests/suites/k8s_monolith"
	"github.com/stretchr/testify/suite"
)

func TestK8sMonolithSuiteSingle(t *testing.T) {
	suite.Run(t, new(k8s_monolith.Suite))
}
