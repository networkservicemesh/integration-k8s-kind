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

// +build calico_test

package main

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/networkservicemesh/integration-tests/suites/basic"
	"github.com/networkservicemesh/integration-tests/suites/heal"
	"github.com/networkservicemesh/integration-tests/suites/memory"
	"github.com/networkservicemesh/integration-tests/suites/observability"
)

func TestRunHealSuite(t *testing.T) {
	suite.Run(t, new(heal.Suite))
}

func TestRunBasicSuite(t *testing.T) {
	suite.Run(t, new(basic.Suite))
}

func TestRunMemorySuite(t *testing.T) {
	suite.Run(t, new(memory.Suite))
}

func TestRunObservabilitySuite(t *testing.T) {
	suite.Run(t, new(observability.Suite))
}