#!/bin/bash

parent_path=$( cd "$(dirname "$0")" ; pwd -P ) || exit

function k1() { kubectl --kubeconfig $KUBECONFIG1 "$@" ; }
function k2() { kubectl --kubeconfig $KUBECONFIG2 "$@" ; }

echo running $0

if [ -z "$1" ]; then echo 1st arg 'nsm_version' is missing; exit 1; fi

nsm_version=$1

echo nsm_version is $nsm_version

#########################

# Specify vl3 NSE version
cat <<EOF > $parent_path/vl3-dns/kustomization.yaml
---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: ns-dns-vl3

bases:
- https://github.com/networkservicemesh/deployments-k8s/apps/nse-vl3-vpp?ref=$nsm_version
- https://github.com/networkservicemesh/deployments-k8s/apps/vl3-ipam?ref=$nsm_version

resources:
- namespace.yaml
- vl3-netsvc.yaml

patchesStrategicMerge:
- nse-patch.yaml
EOF

# Start vl3 NSE
k1 apply -k $parent_path/vl3-dns || exit

# we need to wait a bit to make sure that pods are created, so that wait commands don't fail immediately
sleep 1
k1 -n ns-dns-vl3 wait --for=condition=ready --timeout=5m pod -l app=vl3-ipam || exit
k1 -n ns-dns-vl3 wait --for=condition=ready --timeout=5m pod -l app=nse-vl3-vpp || exit

# Deploy test apps:
k1 create ns perf-test-vl3
k1 apply -n perf-test-vl3 -f $parent_path/apps/nginx.yaml || exit

k2 create ns perf-test-vl3
k2 apply -n perf-test-vl3 -f $parent_path/apps/fortio.yaml || exit

# we need to wait a bit to make sure that pods are created, so that wait commands don't fail immediately
sleep 5
k1 -n perf-test-vl3 wait --for=condition=ready --timeout=5m pod -l app=nginx || exit
k2 -n perf-test-vl3 wait --for=condition=ready --timeout=5m pod -l app=fortio || exit

# open access to the test-load service on local machine
k2 -n perf-test-vl3 port-forward svc/fortio-service 8080:8080 &
# it can take some time for the background job to start listening to local port
sleep 5
