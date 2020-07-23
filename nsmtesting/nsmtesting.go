// Copyright (c) 2020 Doc.ai and/or its affiliates.
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

// Package nsmtesting provides k8s specific testing asserts
package nsmtesting

import (
	"strings"
	"testing"

	"github.com/edwarnicke/exechelper"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes"
)

// NSMTesting is k8s specific asset tool
type NSMTesting struct {
	t      *testing.T
	client *kubernetes.Clientset
}

// New creates new NSMTesting
func New(t *testing.T) *NSMTesting {
	return &NSMTesting{
		t: t,
	}
}

// K8s returns k8s client
func (t *NSMTesting) K8s() *kubernetes.Clientset {
	return t.client
}

// Exec executes command
func (t *NSMTesting) Exec(cmd string) {
	writer := logrus.StandardLogger().Writer()
	var errWriter strings.Builder
	err := exechelper.Run(cmd, exechelper.WithStderr(&errWriter), exechelper.WithStdout(writer))
	require.NoError(t.t, err, errWriter.String())
}
