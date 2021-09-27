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

// Package logs exports helper functions for storing logs from containers.
package logs

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"sync"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultQPS        = 5 // this is default value for QPS of kubeconfig. See at documentation.
	fromAllNamespaces = ""
	kubeconfigEnv     = "KUBECONFIG"
)

var (
	once       sync.Once
	config     Config
	jobsCh     chan func()
	ctx        context.Context
	kubeClient kubernetes.Interface
	matchRegex *regexp.Regexp
)

// Config is env config to setup log collecting.
type Config struct {
	KubeConfig        string        `default:"" desc:".kube config file path" envconfig:"KUBECONFIG"`
	ArtifactsDir      string        `default:"logs" desc:"Directory for storing container logs" envconfig:"ARTIFACTS_DIR"`
	Timeout           time.Duration `default:"5s" desc:"Context timeout for kubernetes queries" split_words:"true"`
	WorkerCount       int           `default:"8" desc:"Number of log collector workers" split_words:"true"`
	AllowedNamespaces string        `default:"(ns-.*)|(nsm-system)" desc:"Regex of allowed namespaces" split_words:"true"`
}

func savePodLogs(ctx context.Context, pod *corev1.Pod, opts *corev1.PodLogOptions, fromInitContainers bool, dir string) {
	containers := pod.Spec.Containers
	if fromInitContainers {
		containers = pod.Spec.InitContainers
	}
	for _, prev := range []bool{false, true} {
		opts.Previous = prev
		for i := 0; i < len(containers); i++ {
			opts.Container = containers[i].Name

			// Add container name to log filename in case of init-containers or multiple containers in the pod
			containerName := ""
			if fromInitContainers || len(containers) > 1 {
				containerName = "-" + containers[i].Name
			}

			// Retrieve logs
			data, err := kubeClient.CoreV1().
				Pods(pod.Namespace).
				GetLogs(pod.Name, opts).
				DoRaw(ctx)
			if err != nil {
				logrus.Errorf("%v: An error while retrieving logs: %v", pod.Name, err.Error())
				return
			}

			// Save logs
			suffix := ".logs"
			if opts.Previous {
				suffix = "-previous.logs"
			}
			err = ioutil.WriteFile(filepath.Join(dir, pod.Name+containerName+suffix), data, os.ModePerm)
			if err != nil {
				logrus.Errorf("An error during saving logs: %v", err.Error())
			}
		}
	}
}

func captureLogs(from time.Time, dir string) {
	operationCtx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()
	resp, err := kubeClient.CoreV1().Pods(fromAllNamespaces).List(operationCtx, metav1.ListOptions{})
	if err != nil {
		logrus.Errorf("An error while retrieving list of pods: %v", err.Error())
	}
	var wg sync.WaitGroup

	for i := 0; i < len(resp.Items); i++ {
		pod := &resp.Items[i]
		if !matchRegex.MatchString(pod.Namespace) {
			continue
		}
		wg.Add(1)
		captureLogsTask := func() {
			opts := &corev1.PodLogOptions{
				SinceTime: &metav1.Time{Time: from},
			}
			savePodLogs(operationCtx, pod, opts, false, dir)
			savePodLogs(operationCtx, pod, opts, true, dir)

			wg.Done()
		}
		select {
		case <-ctx.Done():
			return
		case jobsCh <- captureLogsTask:
			continue
		}
	}

	wg.Wait()
}

func initialize() {
	const prefix = "logs"
	if err := envconfig.Usage(prefix, &config); err != nil {
		logrus.Fatal(err.Error())
	}

	if err := envconfig.Process(prefix, &config); err != nil {
		logrus.Fatal(err.Error())
	}

	matchRegex = regexp.MustCompile(config.AllowedNamespaces)

	jobsCh = make(chan func(), config.WorkerCount)

	if config.KubeConfig == "" {
		config.KubeConfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}

	kubeconfig, err := clientcmd.BuildConfigFromFlags("", config.KubeConfig)
	if err != nil {
		logrus.Fatal(err.Error())
	}

	kubeconfig.QPS = float32(config.WorkerCount) * defaultQPS
	kubeconfig.Burst = int(kubeconfig.QPS) * 2

	kubeClient, err = kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		logrus.Fatal(err.Error())
	}

	var cancel context.CancelFunc
	ctx, cancel = signal.NotifyContext(context.Background(),
		os.Interrupt,
		os.Kill,
		syscall.SIGHUP,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	for i := 0; i < config.WorkerCount; i++ {
		go func() {
			for j := range jobsCh {
				j()
			}
		}()
	}

	go func() {
		defer cancel()
		<-ctx.Done()
		close(jobsCh)
	}()
}

func capture(name string) context.CancelFunc {
	once.Do(initialize)

	now := time.Now()

	dir := filepath.Join(config.ArtifactsDir, name)
	_ = os.MkdirAll(dir, os.ModePerm)

	return func() {
		captureLogs(now, dir)
	}
}

// Capture returns a function that saves logs since Capture function has been called.
func Capture(name string) context.CancelFunc {
	c := capture(name)

	return func() {
		kubeconfigValue := os.Getenv(kubeconfigEnv)
		c()
		for i := 0; ; i++ {
			val := os.Getenv(kubeconfigEnv + fmt.Sprint(i))

			if val == "" {
				break
			}

			_ = os.Setenv(kubeconfigEnv, val)
			c()
		}
		_ = os.Setenv(kubeconfigEnv, kubeconfigValue)
	}
}
