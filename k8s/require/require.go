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

// Package require provides Kubernetes assertion functions that stop test on failure to fulfill a condition
package require

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/networkservicemesh/integration-k8s-kind/k8s"
)

// NoRestarts check that current pods have not restarts
func NoRestarts(t *testing.T) {
	c, err := k8s.Client()
	require.NoError(t, err)
	list, err := c.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})
	require.NoError(t, err)
	for i := 0; i < len(list.Items); i++ {
		pod := &list.Items[i]
		for j := 0; j < len(pod.Status.ContainerStatuses); j++ {
			status := &pod.Status.ContainerStatuses[j]
			reason := ""
			if status.LastTerminationState.Terminated != nil {
				reason = status.LastTerminationState.Terminated.Reason
			}

			require.Zero(
				t,
				status.RestartCount,
				fmt.Sprintf(
					"Container %v of Pod %v has restart count more then zero. Reason: %v",
					status.Name,
					pod.Name,
					reason,
				),
			)
		}
	}
}
