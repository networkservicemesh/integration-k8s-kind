package require

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/networkservicemesh/integration-k8s-kind/k8s"
)

func NoRestarts(t *testing.T) {
	c, err := k8s.Client()
	require.NoError(t, err)
	list, err := c.CoreV1().Pods("default").List(metav1.ListOptions{})
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
