#!/bin/bash

parent_path=$( cd "$(dirname "$0")" ; pwd -P ) || exit

if [ -z "$1" ]; then echo 1st arg 'nsm_version' is missing; exit 1; fi

nsm_version=$1

echo nsm_version is $nsm_version

#########################

cat <<EOF >$parent_path/c1/kustomization.yaml
---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

bases:
- https://github.com/networkservicemesh/deployments-k8s/examples/interdomain/nsm/cluster1?ref=$nsm_version

patchesStrategicMerge:
- forwarder-patch.yaml
EOF

cat <<EOF >$parent_path/c2/kustomization.yaml
---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

bases:
- https://github.com/networkservicemesh/deployments-k8s/examples/interdomain/nsm/cluster2?ref=$nsm_version

patchesStrategicMerge:
- forwarder-patch.yaml
EOF

kubectl --kubeconfig=$KUBECONFIG1 apply -k $parent_path/c1 || (sleep 10 && kubectl --kubeconfig=$KUBECONFIG1 apply -k $parent_path/c1) || exit
kubectl --kubeconfig=$KUBECONFIG2 apply -k $parent_path/c2 || (sleep 10 && kubectl --kubeconfig=$KUBECONFIG2 apply -k $parent_path/c2) || exit

sleep 5

kubectl --kubeconfig=$KUBECONFIG1 wait --for=condition=ready --timeout=1m pod -n nsm-system --all || exit
kubectl --kubeconfig=$KUBECONFIG2 wait --for=condition=ready --timeout=1m pod -n nsm-system --all || exit
