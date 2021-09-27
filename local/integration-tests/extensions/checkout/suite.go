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

// Package checkout contains a suite that checkouts missed repository.
package checkout

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/networkservicemesh/gotestmd/pkg/suites/shell"
)

// Suite clones the repository if it is not presented on the running file system.
type Suite struct {
	shell.Suite
	Repository string
	Dir        string
	Version    string
}

const urlFormat = "https://github.com/%v.git"

// SetupSuite clones repository if it is not presented on the local machine.
func (s *Suite) SetupSuite() {
	r := s.Runner(s.Dir)
	u := fmt.Sprintf(urlFormat, s.Repository)
	_, dir := path.Split(s.Repository)
	repoDir := filepath.Join(r.Dir(), dir)
	// #nosec
	if _, err := os.Open(repoDir); err != nil {
		r.Run("git clone " + u)
		r.Run("cd " + repoDir)
		r.Run("git checkout " + s.Version)
	}
}
