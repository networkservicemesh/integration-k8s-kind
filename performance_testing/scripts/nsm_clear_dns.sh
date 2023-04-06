#!/bin/bash

kubectl "--kubeconfig=$KUBECONFIG1" delete service -n kube-system exposed-kube-dns
kubectl "--kubeconfig=$KUBECONFIG2" delete service -n kube-system exposed-kube-dns

true
