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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/networkservicemesh/integration-k8s-kind/pullkind"
	"github.com/networkservicemesh/integration-k8s-kind/setup"
)

func Test(t *testing.T) {
	suite.Run(t, new(setup.Suite))

	start := time.Now()
	println("-- START: ", fmt.Sprint(start))

	suite.Run(t, new(pullkind.Suite))

	println("-- END: ", fmt.Sprint(time.Now()))
	println("-- DURATION: ", fmt.Sprint(time.Since(start)))
}
