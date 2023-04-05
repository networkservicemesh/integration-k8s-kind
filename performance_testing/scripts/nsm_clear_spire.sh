#!/bin/bash

kubectl --kubeconfig=$KUBECONFIG1 delete crd spiffeids.spiffeid.spiffe.io
kubectl --kubeconfig=$KUBECONFIG1 delete ns spire

kubectl --kubeconfig=$KUBECONFIG2 delete crd spiffeids.spiffeid.spiffe.io
kubectl --kubeconfig=$KUBECONFIG2 delete ns spire

true
