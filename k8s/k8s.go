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

// Package k8s provides kubernetes helper functions
package k8s

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/edwarnicke/exechelper"
	errors "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	apiv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

const (
	// DefaultNamespace - a default kubernetes namespace to use
	DefaultNamespace = "default"
	// CIImagePrefix - a prefix to tag images with
	CIImagePrefix = "networkservicemeshci"
	// CIImageTag - a version for tagged images.
	CIImageTag = "master"
)

var _client *kubernetes.Clientset
var _config *rest.Config

var deployments map[string]bool = map[string]bool{}
var builds map[string]bool = map[string]bool{}
var writer *io.PipeWriter = logrus.StandardLogger().Writer()
var options []*exechelper.Option = []*exechelper.Option{
	exechelper.WithStderr(writer),
	exechelper.WithStdout(writer),
}
var lock sync.Mutex
var once sync.Once

// client returns kubernetes client
func client() *kubernetes.Clientset {
	once.Do(func() {
		doInit()
	})
	return _client
}

func doInit() {
	path := os.Getenv("KUBECONFIG")
	if path == "" {
		path = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}

	var err error
	_config, err = clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		logrus.Fatalf("failed to connect kubernetes: %v", err)
	}
	_client, err = kubernetes.NewForConfig(_config)
	if err != nil {
		logrus.Fatalf("failed to connect kubernetes: %v", err)
	}
}

// ApplyDeployment is analogy of 'kubeclt apply -f path' but with mutating deployment before apply
func ApplyDeployment(path string, mutators ...func(deployment *apiv1.Deployment)) error {
	_, err := ApplyGetDeployment(path, mutators...)
	return err
}

// ApplyGetDeployment is analogy of 'kubeclt apply -f path' but with mutating deployment before apply, and it return deployment
func ApplyGetDeployment(path string, mutators ...func(deployment *apiv1.Deployment)) (*apiv1.Deployment, error) {
	b, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	var d apiv1.Deployment
	if parseErr := yaml.Unmarshal(b, &d); parseErr != nil {
		return nil, parseErr
	}
	for _, m := range mutators {
		m(&d)
	}
	var dd *apiv1.Deployment
	dd, err = client().AppsV1().Deployments(DefaultNamespace).Get(d.Name, metav1.GetOptions{})
	if err == nil && dd != nil {
		return client().AppsV1().Deployments(DefaultNamespace).Update(&d)
	}
	return client().AppsV1().Deployments(DefaultNamespace).Create(&d)
}

// UpdateDeployment - rerieve a fresh informaton for deployment
func UpdateDeployment(deployment *apiv1.Deployment) (*apiv1.Deployment, error) {
	return client().AppsV1().Deployments(DefaultNamespace).Get(deployment.Name, metav1.GetOptions{})
}

// WaitDeploymentReady - wait for required number of replas to be ready
func WaitDeploymentReady(deployment *apiv1.Deployment, replicas int, timeout time.Duration) (err error) {
	t := time.Now()
	for deployment.Status.ReadyReplicas == 0 {
		deployment, err = UpdateDeployment(deployment)
		if err != nil {
			return err
		}
		time.Sleep(time.Millisecond * 50)
		if time.Since(t) > timeout {
			return errors.Errorf("timeout waiting for pod replicas %v", deployment.Status.ReadyReplicas)
		}
	}
	return nil
}

// ShowLogs prints logs into console all containers of pods
func ShowLogs(namespace string, options ...*exechelper.Option) {
	pods, err := client().CoreV1().Pods(namespace).List(metav1.ListOptions{})

	if err != nil {
		logrus.Errorf("Cannot get pods: %v", err.Error())
		return
	}

	for i := 0; i < len(pods.Items); i++ {
		pod := &pods.Items[i]
		for j := 0; j < len(pod.Spec.Containers); j++ {
			container := &pod.Spec.Containers[j]
			ns := ""
			if namespace != DefaultNamespace {
				ns = fmt.Sprintf("--namespace=%v", namespace)
			}
			_ = exechelper.Run(fmt.Sprintf("kubectl logs %v -c %v %v ", pod.Name, container.Name, ns), options...)
		}
	}
}

// NamespaceExists - check if passed DefaultNamespace are exists
func NamespaceExists(name string) bool {
	_, err := client().CoreV1().Namespaces().Get(name, metav1.GetOptions{})
	return err == nil
}

// DescribePod - Find a pod by name or by matching all labels passed.
func DescribePod(namespace, name string, labels map[string]string) (*corev1.Pod, error) {
	var pods *corev1.PodList
	var err error
	pods, err = client().CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(pods.Items); i++ {
		pod := pods.Items[i]
		// If name matches
		if pod.Name == name {
			return &pod, nil
		}
		if labels != nil && pod.Labels != nil && pod.Status.Phase == corev1.PodRunning {
			// Check if all labels are in pod labels,
			matches := len(labels)
			for k, v := range labels {
				if pod.Labels[k] == v {
					matches--
				}
			}
			if matches == 0 {
				return &pod, nil
			}
		}
	}

	return nil, nil
}

// ListPods -  List all pods by matching all labels passed.
// namespace - a namespace to check in
// nameExpr - name matching expression.
// labels - a set of labels pod to contain
func ListPods(namespace, nameExpr string, labels map[string]string) ([]*corev1.Pod, error) {
	var pods *corev1.PodList
	var err error
	var result []*corev1.Pod
	pods, err = client().CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var expr *regexp.Regexp
	if nameExpr != "" {
		expr, err = regexp.Compile(nameExpr)
		if err != nil {
			return nil, err
		}
	}
	for i := 0; i < len(pods.Items); i++ {
		pod := pods.Items[i]

		if expr != nil {
			if !expr.MatchString(pod.Name) {
				// Skip not matched pods.
				continue
			}
		}
		if labels != nil && pod.Labels != nil && pod.Status.Phase == corev1.PodRunning {
			// Check if all labels are in pod labels,
			matches := len(labels)
			for k, v := range labels {
				if pod.Labels[k] == v {
					matches--
				}
			}
			if matches == 0 {
				result = append(result, &pod)
			}
		}
	}

	return result, nil
}

// NoRestarts check that current pods have not restarts
func NoRestarts(t *testing.T) {
	list, err := client().CoreV1().Pods("default").List(metav1.ListOptions{})
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

// CleanNamespace - remove all serviceaccounts/services/deployments and delete all pods from namespace.
func CleanNamespace(namespace string, options ...*exechelper.Option) (err error) {
	ShowLogs(namespace, options...)

	var wg sync.WaitGroup
	wg.Add(4)
	go func() {
		defer func() { wg.Done() }()
		e := exechelper.Run("kubectl delete serviceaccounts --all")
		if e != nil {
			err = e
		}
	}()
	go func() {
		defer func() { wg.Done() }()
		e := exechelper.Run("kubectl delete services --all")
		if e != nil {
			err = e
		}
	}()
	go func() {
		defer func() { wg.Done() }()
		e := exechelper.Run("kubectl delete deployment --all")
		if e != nil {
			err = e
		}
	}()
	go func() {
		defer func() { wg.Done() }()
		e := exechelper.Run("kubectl delete pods --all --grace-period=0 --force --wait")
		if e != nil {
			err = e
		}
	}()
	wg.Wait()
	return err
}

// DockerBuild - build a particular app using docker from cmd.
func DockerBuild(name string) error {
	lock.Lock()
	defer lock.Unlock()
	if _, ok := builds[name]; ok {
		// We already did this
		return nil
	}
	builds[name] = true
	cmd := fmt.Sprintf("docker build cmd/%s -t %s/%s:%s ", name, CIImagePrefix, name, CIImageTag)
	logrus.Infof("Running: %v", cmd)
	return exechelper.Run(cmd, options...)
}

// KindDeployApp - deploy container name to kind
func KindDeployApp(name string) error {
	lock.Lock()
	defer lock.Unlock()
	if _, ok := deployments[name]; ok {
		// We already did this
		return nil
	}
	deployments[name] = true
	cmd := fmt.Sprintf("kind load docker-image %s/%s:%s ", CIImagePrefix, name, CIImageTag)
	logrus.Infof("Running: %v", cmd)
	return exechelper.Run(cmd, options...)
}

// KindRequire - require for particilar container to be build and deployed/
func KindRequire(containers ...string) error {
	for _, c := range containers {
		if err := DockerBuild(c); err != nil {
			return err
		}
		if err := KindDeployApp(c); err != nil {
			return err
		}
	}
	return nil
}

// GetFreePort - find unused local port
func GetFreePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}

	port := listener.Addr().(*net.TCPAddr).Port
	err = listener.Close()
	if err != nil {
		return 0, err
	}

	return port, nil
}
