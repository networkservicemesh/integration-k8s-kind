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

// package test contains k8s integration tests
package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/edwarnicke/exechelper"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s_yaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta2"

	"github.com/networkservicemesh/sdk/pkg/tools/prefixpool"

	"github.com/networkservicemesh/integration-k8s-kind/k8s"
)

const (
	nsmConfigDir = "var/lib/networkservicemesh/config"
	// kubeNamespace is KubeAdm ConfigMap namespace
	kubeNamespace = "kube-system"
	// kubeName is KubeAdm ConfigMap name
	kubeName            = "kubeadm-config"
	configMapPath       = "./files/userConfigMap.yaml"
	prefixesFileName    = "excluded_prefixes.yaml"
	bufferSize          = 4096
	appLabelKey         = "app"
	collectorNamespace  = "excluded-prefixes-collector"
	excludedPrefixesEnv = "EXCLUDE_PREFIXES_K8S_EXCLUDED_PREFIXES"
)

type ExcludedPrefixesSuite struct {
	suite.Suite
	options       []*exechelper.Option
	alpinePodName string
}

type prefixes struct {
	Prefixes []string
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

	et.Require().NoError(k8s.ApplyDeployment("../deployments/alpine.yaml", func(alpine *v1.Deployment) {
		alpine.Namespace = collectorNamespace
	}))
	et.waitForPodStart(collectorNamespace, "alpine")

	podInfo, err := k8s.DescribePod(collectorNamespace, "", map[string]string{
		"app": "alpine",
	})
	et.Require().NoError(err)
	et.alpinePodName = podInfo.Name
}

func (et *ExcludedPrefixesSuite) TearDownTest() {
	k8s.ShowLogs(collectorNamespace, et.options...)

	et.Require().NoError(exechelper.Run("kubectl delete daemonset -n excluded-prefixes-collector excluded-prefixes-collector"))
	et.Require().NoError(exechelper.Run("kubectl delete pods -n excluded-prefixes-collector -l app=excluded-prefixes-collector --now"))
	logrus.Info("Collector deleted")
}

func (et *ExcludedPrefixesSuite) TearDownSuite() {
	et.Require().NoError(exechelper.Run("kubectl delete -f ../deployments/prefixes-collector/collector-cluster-role.yaml", et.options...))
	et.Require().NoError(exechelper.Run("kubectl delete -f ../deployments/prefixes-collector/collector-namespace.yaml --now", et.options...))
}

func (et *ExcludedPrefixesSuite) TestWithKubeAdmConfigPrefixes() {
	et.deployCollector()

	expectedPrefixes, err := kubeAdmPrefixes()
	et.Require().NoError(err)

	et.Eventually(et.checkPrefixes(expectedPrefixes), time.Second*15, time.Second)
}

func (et *ExcludedPrefixesSuite) TestWithUserConfigPrefixes() {
	et.deployCollector()

	userConfigMap, err := userConfigMap(configMapPath)
	et.Require().NoError(err)

	configMapsInterface, err := k8s.ConfigMapInterface(userConfigMap.Namespace)
	et.Require().NoError(err)

	userConfigMap, err = configMapsInterface.Create(userConfigMap)
	et.Require().NoError(err)

	kubeAdmPrefixes, err := kubeAdmPrefixes()
	et.Require().NoError(err)

	prefixPool, err := prefixpool.New(kubeAdmPrefixes...)
	et.Require().NoError(err)

	userPrefixes := prefixes{}
	err = k8s_yaml.NewYAMLOrJSONDecoder(
		strings.NewReader(userConfigMap.Data[prefixesFileName]), bufferSize,
	).Decode(&userPrefixes)
	et.Require().NoError(err)

	err = prefixPool.ReleaseExcludedPrefixes(userPrefixes.Prefixes)
	et.Require().NoError(err)

	et.Eventually(et.checkPrefixes(prefixPool.GetPrefixes()), time.Second*15, time.Second)

	userConfigMap.Data[prefixesFileName] = "Prefixes:\n- 128.0.0.0/1\n- 0.0.0.0/1"
	userConfigMap, err = configMapsInterface.Update(userConfigMap)
	et.Require().NoError(err)

	expectedPrefixes := []string{"0.0.0.0/0"}
	et.Eventually(et.checkPrefixes(expectedPrefixes), time.Second*15, time.Second)

	err = configMapsInterface.Delete(userConfigMap.Name, &metav1.DeleteOptions{})
	et.Require().NoError(err)

	et.Eventually(et.checkPrefixes(kubeAdmPrefixes), time.Second*15, time.Second)
}

func (et *ExcludedPrefixesSuite) TestWithAllPrefixes() {
	envPrefixes := []string{
		"127.0.0.0/8",
		"134.65.0.0/16",
	}

	et.deployCollectorWithEnvs(envPrefixes)

	prefixPool, err := prefixpool.New(envPrefixes...)
	et.Require().NoError(err)

	userConfigMap, err := userConfigMap(configMapPath)
	et.Require().NoError(err)

	configMapsInterface, err := k8s.ConfigMapInterface(userConfigMap.Namespace)
	et.Require().NoError(err)

	userConfigMap, err = configMapsInterface.Create(userConfigMap)
	et.Require().NoError(err)

	defer func() {
		_ = configMapsInterface.Delete(userConfigMap.Name, &metav1.DeleteOptions{})
	}()

	kubeAdmPrefixes, err := kubeAdmPrefixes()
	et.Require().NoError(err)

	err = prefixPool.ReleaseExcludedPrefixes(kubeAdmPrefixes)
	et.Require().NoError(err)

	userPrefixes := prefixes{}
	err = k8s_yaml.NewYAMLOrJSONDecoder(
		strings.NewReader(userConfigMap.Data[prefixesFileName]), bufferSize,
	).Decode(&userPrefixes)
	et.Require().NoError(err)

	err = prefixPool.ReleaseExcludedPrefixes(userPrefixes.Prefixes)
	et.Require().NoError(err)

	et.Eventually(et.checkPrefixes(prefixPool.GetPrefixes()), time.Second*15, time.Second)
}

func (et *ExcludedPrefixesSuite) TestWithCorrectEnvPrefixes() {
	envPrefixes := []string{
		"127.0.0.0/8",
		"134.65.0.0/16",
	}

	prefixPool, _ := prefixpool.New(envPrefixes...)
	et.deployCollectorWithEnvs(envPrefixes)

	kubeAdmPrefixes, err := kubeAdmPrefixes()
	et.Require().NoError(err)

	err = prefixPool.ReleaseExcludedPrefixes(kubeAdmPrefixes)
	et.Require().NoError(err)

	et.Eventually(et.checkPrefixes(prefixPool.GetPrefixes()), time.Second*15, time.Second)
}

func (et *ExcludedPrefixesSuite) TestWithIncorrectEnvPrefixes() {
	envPrefixes := []string{
		"256.256.256.0",
	}
	et.deployCollectorWithEnvs(envPrefixes)

	et.Eventually(func() bool {
		podInfo, err := k8s.DescribePod(prefixesFileName, "", map[string]string{
			"app": "excluded-prefixes-collector",
		})
		et.Require().NoError(err)
		return podInfo == nil
	}, time.Second*15, time.Second)
}

func TestExcludedPrefixesSuite(t *testing.T) {
	suite.Run(t, &ExcludedPrefixesSuite{})
}

func (et *ExcludedPrefixesSuite) deployCollector() {
	et.Require().NoError(exechelper.Run("kubectl apply -f ../deployments/prefixes-collector/collector.yaml", et.options...))
	et.waitForPodStart(collectorNamespace, "excluded-prefixes-collector")
	logrus.Info("Collector deployed")
}

func (et *ExcludedPrefixesSuite) deployCollectorWithEnvs(envPrefixes []string) {
	et.Require().NoError(k8s.ApplyDaemonSet("../deployments/prefixes-collector/collector.yaml", func(collector *v1.DaemonSet) {
		collector.Spec.Template.Spec.Containers[0].Env = append(collector.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
			Name:  excludedPrefixesEnv,
			Value: strings.Join(envPrefixes, ","),
		})
	}))
	et.waitForPodStart(collectorNamespace, "excluded-prefixes-collector")
	logrus.Info("Collector deployed")
}

func (et *ExcludedPrefixesSuite) checkPrefixes(expectedPrefixes []string) func() bool {
	dirName := os.TempDir()
	kubeCpCmd := fmt.Sprintf("kubectl cp -n %v %v:%v %v", collectorNamespace, et.alpinePodName, nsmConfigDir, dirName)
	localPrefixesPath := filepath.Join(dirName, prefixesFileName)
	return func() bool {
		et.Require().NoError(exechelper.Run(kubeCpCmd))
		actualPrefixes, err := prefixesFromFile(localPrefixesPath)
		logrus.Infof("Actual: %v Expected: %v", actualPrefixes, expectedPrefixes)
		return err == nil && reflect.DeepEqual(expectedPrefixes, actualPrefixes)
	}
}

func (et *ExcludedPrefixesSuite) waitForPodStart(namespace, appLabelValue string) {
	client, err := k8s.Client()
	et.Require().NoError(err)

	podInterface := client.CoreV1().Pods(namespace)
	watcher, err := podInterface.Watch(metav1.ListOptions{})
	et.Require().NoError(err)

	for {
		select {
		case <-time.After(time.Second * 15):
			et.T().Fatalf("%v pod watch timeout", appLabelValue)
		case event := <-watcher.ResultChan():
			pod := event.Object.(*corev1.Pod)
			if pod.Labels[appLabelKey] == appLabelValue && pod.Status.Phase == corev1.PodRunning {
				return
			}
		}
	}
}

func prefixesFromFile(filename string) ([]string, error) {
	source, err := ioutil.ReadFile(filepath.Clean(filename))
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	destination := prefixes{}
	err = yaml.Unmarshal(source, &destination)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	return destination.Prefixes, nil
}

func kubeAdmPrefixes() ([]string, error) {
	configMapInterface, err := k8s.ConfigMapInterface(kubeNamespace)
	if err != nil {
		return nil, err
	}

	configMap, err := configMapInterface.Get(kubeName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	clusterConfiguration := &v1beta2.ClusterConfiguration{}
	err = k8s_yaml.NewYAMLOrJSONDecoder(
		strings.NewReader(configMap.Data["ClusterConfiguration"]), bufferSize,
	).Decode(clusterConfiguration)
	if err != nil {
		return nil, err
	}

	return []string{clusterConfiguration.Networking.PodSubnet, clusterConfiguration.Networking.ServiceSubnet}, nil
}

func userConfigMap(filePath string) (*corev1.ConfigMap, error) {
	destination := corev1.ConfigMap{}
	bytes, err := ioutil.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return nil, errors.Wrap(err, "Error reading user config map")
	}
	if err = yaml.Unmarshal(bytes, &destination); err != nil {
		return nil, errors.Wrap(err, "Error decoding user config map")
	}

	return &destination, nil
}
