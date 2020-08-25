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
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/edwarnicke/exechelper"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

var once sync.Once
var client *kubernetes.Clientset
var clientErr error

const namespace = "default"

// Client returns k8s client
func Client() (*kubernetes.Clientset, error) {
	once.Do(func() {
		path := os.Getenv("KUBECONFIG")
		if path == "" {
			path = filepath.Join(os.Getenv("HOME"), ".kube", "config")
		}
		config, err := clientcmd.BuildConfigFromFlags("", path)
		if err != nil {
			clientErr = err
			return
		}
		client, clientErr = kubernetes.NewForConfig(config)
	})
	return client, clientErr
}

// ApplyDeployment is analogy of 'kubeclt apply -f path' but with mutating deployment before apply
func ApplyDeployment(path string, mutators ...func(deployment *v1.Deployment)) error {
	client, err := Client()
	if err != nil {
		return err
	}
	b, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return err
	}
	var d v1.Deployment
	if parseErr := yaml.Unmarshal(b, &d); parseErr != nil {
		return parseErr
	}
	for _, m := range mutators {
		m(&d)
	}
	_, err = client.AppsV1().Deployments(namespace).Create(&d)
	return err
}

// ShowLogs prints logs into console all containers of pods
func ShowLogs(options ...*exechelper.Option) {
	client, err := Client()

	if err != nil {
		logrus.Errorf("Cannot get k8s client: %v", err.Error())
		return
	}

	pods, err := client.CoreV1().Pods(namespace).List(metav1.ListOptions{})

	if err != nil {
		logrus.Errorf("Cannot get pods%v", err.Error())
		return
	}

	for i := 0; i < len(pods.Items); i++ {
		pod := &pods.Items[i]
		for j := 0; j < len(pod.Spec.Containers); j++ {
			container := &pod.Spec.Containers[j]
			_ = exechelper.Run(fmt.Sprintf("kubeclt logs %v -c %v ", pod.Name, container.Name), options...)
		}
	}
}
