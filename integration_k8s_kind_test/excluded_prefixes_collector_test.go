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

// package integration_k8s_kind_test contains k8s integration tests
package integration_k8s_kind_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/edwarnicke/exechelper"
	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/networkservicemesh/integration-k8s-kind/k8s"
)

const (
	prefixesFilePath    = "/var/lib/networkservicemesh/config/excluded_prefixes.yaml"
	collectorNamespace  = "excluded-prefixes-collector"
	excludedPrefixesEnv = "EXCLUDE_PREFIXES_K8S_EXCLUDED_PREFIXES"
	defaultTimeout      = time.Second * 30
	defaultTick         = time.Second * 1
)

type ExcludedPrefixesSuite struct {
	suite.Suite
	options      []*exechelper.Option
	nsmgrPodName string
}

type prefixes struct {
	Prefixes []string
}

var kubeAdmPrefixes = []string{
	"10.244.0.0/16",
	"10.96.0.0/16",
}

var userConfigPrefixes = []string{
	"134.8.0.0/16",
	"64.5.12.0/24",
}

func (et *ExcludedPrefixesSuite) SetupSuite() {
	writer := logrus.StandardLogger().Writer()

	et.options = []*exechelper.Option{
		exechelper.WithStderr(writer),
		exechelper.WithStdout(writer),
	}

	et.Require().NoError(exechelper.Run("kubectl apply -f ../deployments/prefixes-collector/collector-namespace.yaml", et.options...))
	et.Require().NoError(exechelper.Run("kubectl apply -f ../deployments/prefixes-collector/collector-account.yaml", et.options...))
	et.Require().NoError(exechelper.Run("kubectl apply -f ../deployments/prefixes-collector/collector-cluster-role.yaml", et.options...))

	et.setupNsmgr()
}

func (et *ExcludedPrefixesSuite) TearDownTest() {
	k8s.ShowLogs(collectorNamespace, et.options...)

	et.Require().NoError(exechelper.Run("kubectl delete daemonset -n excluded-prefixes-collector excluded-prefixes-collector", et.options...))
	et.Require().NoError(exechelper.Run("kubectl delete pods -n excluded-prefixes-collector -l app=excluded-prefixes-collector --now", et.options...))
	logrus.Info("Collector deleted")
}

func (et *ExcludedPrefixesSuite) TearDownSuite() {
	k8s.ShowLogs(collectorNamespace, et.options...)

	et.Require().NoError(exechelper.Run("kubectl delete -f ../deployments/prefixes-collector/collector-cluster-role.yaml", et.options...))
	et.Require().NoError(exechelper.Run("kubectl delete -f ../deployments/prefixes-collector/collector-namespace.yaml --now", et.options...))
}

func (et *ExcludedPrefixesSuite) TestWithKubeAdmConfigPrefixes() {
	et.Require().NoError(exechelper.Run("kubectl apply -f ../deployments/prefixes-collector/collector.yaml", et.options...))

	et.Eventually(et.checkPrefixes(kubeAdmPrefixes), defaultTimeout, defaultTick)
}

func (et *ExcludedPrefixesSuite) TestWithUserConfigPrefixes() {
	et.Require().NoError(exechelper.Run("kubectl apply -f ../deployments/prefixes-collector/collector.yaml", et.options...))
	et.Require().NoError(exechelper.Run("kubectl apply -f ./files/userConfigMap.yaml", et.options...))

	expectedPrefixes := append(kubeAdmPrefixes, userConfigPrefixes...)
	et.Eventually(et.checkPrefixes(expectedPrefixes), defaultTimeout, defaultTick)

	expectedPrefixes = []string{"0.0.0.0/0"}
	et.Require().NoError(exechelper.Run("kubectl replace -f ./files/updatedUserConfigMap.yaml", et.options...))
	et.Eventually(et.checkPrefixes(expectedPrefixes), defaultTimeout, defaultTick)

	et.Require().NoError(exechelper.Run("kubectl delete configmaps excluded-prefixes-config", et.options...))
	et.Eventually(et.checkPrefixes(kubeAdmPrefixes), defaultTimeout, defaultTick)
}

func (et *ExcludedPrefixesSuite) TestWithAllPrefixes() {
	envPrefixes := []string{
		"127.0.0.0/8",
		"134.65.0.0/16",
	}

	et.Require().NoError(k8s.ApplyDaemonSet("../deployments/prefixes-collector/collector.yaml", func(collector *v1.DaemonSet) {
		collector.Spec.Template.Spec.Containers[0].Env = append(collector.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
			Name:  excludedPrefixesEnv,
			Value: strings.Join(envPrefixes, ","),
		})
	}))

	et.Require().NoError(exechelper.Run("kubectl apply -f ./files/userConfigMap.yaml", et.options...))
	defer func() {
		et.Require().NoError(exechelper.Run("kubectl delete configmaps excluded-prefixes-config", et.options...))
	}()

	expectedPrefixes := append(append(kubeAdmPrefixes, envPrefixes...), userConfigPrefixes...)
	et.Eventually(et.checkPrefixes(expectedPrefixes), defaultTimeout, defaultTick)
}

func (et *ExcludedPrefixesSuite) TestWithCorrectEnvPrefixes() {
	envPrefixes := []string{
		"127.0.0.0/8",
		"134.65.0.0/16",
	}

	et.Require().NoError(k8s.ApplyDaemonSet("../deployments/prefixes-collector/collector.yaml", func(collector *v1.DaemonSet) {
		collector.Spec.Template.Spec.Containers[0].Env = append(collector.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
			Name:  excludedPrefixesEnv,
			Value: strings.Join(envPrefixes, ","),
		})
	}))

	expectedPrefixes := append(kubeAdmPrefixes, envPrefixes...)
	et.Eventually(et.checkPrefixes(expectedPrefixes), defaultTimeout, defaultTick)
}

func (et *ExcludedPrefixesSuite) TestWithIncorrectEnvPrefixes() {
	envPrefixes := []string{
		"306.306.306.0",
	}

	et.Require().NoError(k8s.ApplyDaemonSet("../deployments/prefixes-collector/collector.yaml", func(collector *v1.DaemonSet) {
		collector.Spec.Template.Spec.Containers[0].Env = append(collector.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
			Name:  excludedPrefixesEnv,
			Value: strings.Join(envPrefixes, ","),
		})
	}))

	client, err := k8s.Client()
	et.Require().NoError(err)
	inter := client.AppsV1().DaemonSets(collectorNamespace)
	et.Eventually(func() bool {
		daemonSet, err := inter.Get("excluded-prefixes-collector", metav1.GetOptions{})
		return err == nil && daemonSet.Status.NumberAvailable == 0
	}, defaultTimeout, defaultTick)
}

func TestExcludedPrefixesSuite(t *testing.T) {
	suite.Run(t, &ExcludedPrefixesSuite{})
}

func (et *ExcludedPrefixesSuite) setupNsmgr() {
	et.Require().NoError(k8s.ApplyDaemonSet("../deployments/nsmgr.yaml", func(nsmgr *v1.DaemonSet) {
		nsmgr.Namespace = collectorNamespace
		spec := nsmgr.Spec.Template.Spec
		// Remove spire sockets directory mount
		spec.Volumes = spec.Volumes[1:]
		spec.Containers[0].VolumeMounts = spec.Containers[0].VolumeMounts[1:]
	}))

	var podInfo *corev1.Pod
	labels := map[string]string{
		"app": "nsmgr",
	}

	et.Eventually(func() bool {
		var err error
		if podInfo == nil {
			podInfo, err = k8s.GetPod(collectorNamespace, "", labels)
			et.Require().NoError(err)
		}
		return podInfo != nil
	}, time.Second*60, time.Second*5)

	et.nsmgrPodName = podInfo.Name
}

func (et *ExcludedPrefixesSuite) checkPrefixes(expectedPrefixes []string) func() bool {
	expectedPrefixesYaml, err := yaml.Marshal(&prefixes{expectedPrefixes})
	et.Require().NoError(err)

	return func() bool {
		var sb strings.Builder
		cmd := fmt.Sprintf("kubectl exec -n excluded-prefixes-collector -ti %v -- cat %v", et.nsmgrPodName, prefixesFilePath)
		err := exechelper.Run(cmd, exechelper.WithStdout(&sb))
		logrus.Infof("Current: %v Expected: %v", sb.String(), string(expectedPrefixesYaml))
		return err == nil && sb.String() == string(expectedPrefixesYaml)
	}
}
