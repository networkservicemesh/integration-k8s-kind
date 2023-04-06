#!/bin/bash

parent_path=$( cd "$(dirname "$0")" ; pwd -P ) || exit

function k1() { kubectl --kubeconfig "$KUBECONFIG1" "$@" ; }
function k2() { kubectl --kubeconfig "$KUBECONFIG2" "$@" ; }

echo running $0

if [ -z "$1" ]; then echo 1st arg 'nsm_version' is missing; exit 1; fi

nsm_version=$1

echo nsm_version is "$nsm_version"

#########################

# Specify NSM version for NSE
cat <<EOF > "$parent_path/cluster1/kustomization.yaml"
---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: perf-test-wg

bases:
- https://github.com/networkservicemesh/deployments-k8s/apps/nse-kernel?ref=$nsm_version

patchesStrategicMerge:
- patch-nse.yaml
EOF

# Deploy nginx
k1 create ns perf-test-wg
k1 apply -k "$parent_path/cluster1" || exit

# we need to wait a bit to make sure that pods are created, so that wait commands don't fail immediately
sleep 1

# Deploy fortio
k2 create ns perf-test-wg
k2 apply -n perf-test-wg -f "$parent_path/cluster2/fortio.yaml" || exit

# we need to wait a bit to make sure that pods are created, so that wait commands don't fail immediately
sleep 5
k1 -n perf-test-wg wait --for=condition=ready --timeout=1m pod -l app=nse-kernel || exit
k2 -n perf-test-wg wait --for=condition=ready --timeout=5m pod -l app=fortio || exit

# open access to the test-load service on local machine
k2 -n perf-test-wg port-forward svc/fortio-service 8080:8080 &
# it can take some time for the background job to start listening to local port
sleep 5
