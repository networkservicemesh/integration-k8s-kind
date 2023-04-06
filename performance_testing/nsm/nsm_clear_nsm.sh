#!/bin/bash

WH=$(kubectl "--kubeconfig=$KUBECONFIG1" get pods -l app=admission-webhook-k8s -n nsm-system --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}')
kubectl "--kubeconfig=$KUBECONFIG1" delete mutatingwebhookconfiguration "${WH}"
kubectl "--kubeconfig=$KUBECONFIG1" delete ns nsm-system

WH=$(kubectl "--kubeconfig=$KUBECONFIG2" get pods -l app=admission-webhook-k8s -n nsm-system --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}')
kubectl "--kubeconfig=$KUBECONFIG2" delete mutatingwebhookconfiguration "${WH}"
kubectl "--kubeconfig=$KUBECONFIG2" delete ns nsm-system

true
