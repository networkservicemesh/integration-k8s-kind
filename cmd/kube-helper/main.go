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

// Package main define a kube-helper debug tool application
package main

import (
	"context"
	"fmt"
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/edwarnicke/exechelper"
	"github.com/networkservicemesh/integration-k8s-kind/k8s"
	"github.com/networkservicemesh/integration-k8s-kind/spire"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/networkservicemesh/sdk/pkg/tools/log"
	"github.com/networkservicemesh/sdk/pkg/tools/signalctx"
)

func main() {
	// ********************************************************************************
	// Configure signal handling context
	// ********************************************************************************
	ctx := signalctx.WithSignals(context.Background())
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	// ********************************************************************************
	// Setup logger
	// ********************************************************************************
	logrus.Info("Kube helper application. Please use it to proxy nsmgr on different ports.")
	logrus.SetFormatter(&nested.Formatter{})
	logrus.SetLevel(logrus.TraceLevel)

	args := os.Args[1:]

	if len(args) == 0 {
		logrus.Infof(
			`
Please pass one of following commands:
	start-spire - will start spire using ./spire/spire.go
	port-forward namespace prefix initialport pod-port - will forward a pods port to local one using kubectl port-forward with startd initialport with increment.
`)
		return
	}
	if args[0] == "start-spire" {
		logrus.Infof("start spire")
		_ = spire.Setup(exechelper.WithContext(ctx), exechelper.WithStdout(os.Stdout), exechelper.WithStderr(os.Stdout))
	}
	if args[0] == "port-forward" && len(args) == 5 {
		logrus.Infof("Proxy every nsmgr one by one")

		namespace := args[1]
		pattern := args[2]
		initialPort, err := strconv.Atoi(args[3])
		var podPort int
		podPort, err = strconv.Atoi(args[4])
		if err != nil {
			logrus.Fatal("failed to parse initial port %v", args[3])
		}

		for {
			pods, err := k8s.ListPods(namespace, pattern, map[string]string{})

			if err != nil {
				logrus.Fatalf("failed to find pods %v", err)
			}
			podMap := map[string]int{}
			iterationCtx, icancel := context.WithCancel(ctx)
			for i, p := range pods {
				// Tail all logs
				rctx := log.WithField(iterationCtx, "cmd", p.Name)

				// Forward port
				cmd := fmt.Sprintf("kubectl port-forward %v %v:%v --namespace %v", p.Name, initialPort+i, podPort, namespace)

				logrus.Infof("Staring %v", cmd)
				_ = exechelper.Start(cmd, exechelper.WithContext(rctx), exechelper.WithStdout(os.Stdout), exechelper.WithStderr(os.Stdout))
				logrus.Infof("Forward started for %v on port: %v", p.Name, initialPort+i)
				podMap[p.Name] = initialPort + i
			}

			for {
				pods2, err := k8s.ListPods(namespace, pattern, map[string]string{})
				if err != nil {
					logrus.Fatalf("failed to find pods %v", err)
				}
				changeDetected := false
				if len(pods2) != len(pods) {
					logrus.Infof("Pod change detected. Restarting forward")
					changeDetected = true
				}

				for _, pp := range pods2 {
					if _, ok := podMap[pp.Name]; !ok {
						changeDetected = true
						break
					}
				}
				if changeDetected {
					logrus.Infof("Pod change detected. Restarting forward")
					break
				}
				time.Sleep(1000)
			}
			icancel()
		}

	}
	if args[0] == "logs" && len(args) == 3 {
		for {
			var wg sync.WaitGroup

			logrus.Infof("Listen for %v", args)
			pods, err := k8s.ListPods(args[1], args[2], map[string]string{})
			if err != nil {
				logrus.Errorf("failed to find pods %v", err)
			}
			for _, p := range pods {
				pp := p
				rctx := log.WithField(ctx, "cmd", p.Name)
				wg.Add(1)
				// Tail all logs
				go func() {
					defer wg.Done()
					ec := exechelper.Start(fmt.Sprintf("kubectl logs -f %v --namespace %v", pp.Name, args[1]), exechelper.WithContext(rctx), exechelper.WithStdout(os.Stdout), exechelper.WithStderr(os.Stdout))
					<-ec
				}()
			}
			wg.Wait()
			if ctx.Err() != nil {
				break
			}
			time.Sleep(time.Second)
		}
	}

	// Wait for cancel event to terminate
	<-ctx.Done()
}
