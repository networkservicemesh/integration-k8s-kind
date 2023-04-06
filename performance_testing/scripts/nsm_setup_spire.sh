#!/bin/bash

kubectl "--kubeconfig=$KUBECONFIG1" apply -k https://github.com/networkservicemesh/deployments-k8s/examples/spire/cluster1?ref=v1.8.0 || exit
kubectl "--kubeconfig=$KUBECONFIG2" apply -k https://github.com/networkservicemesh/deployments-k8s/examples/spire/cluster2?ref=v1.8.0 || exit

sleep 1

kubectl "--kubeconfig=$KUBECONFIG1" wait -n spire --timeout=1m --for=condition=ready pod -l app=spire-server || exit
kubectl "--kubeconfig=$KUBECONFIG2" wait -n spire --timeout=1m --for=condition=ready pod -l app=spire-server || exit

kubectl "--kubeconfig=$KUBECONFIG1" wait -n spire --timeout=1m --for=condition=ready pod -l app=spire-agent || exit
kubectl "--kubeconfig=$KUBECONFIG2" wait -n spire --timeout=1m --for=condition=ready pod -l app=spire-agent || exit

bundle1=$(kubectl "--kubeconfig=$KUBECONFIG1" exec spire-server-0 -n spire -- bin/spire-server bundle show -format spiffe) || exit
bundle2=$(kubectl "--kubeconfig=$KUBECONFIG2" exec spire-server-0 -n spire -- bin/spire-server bundle show -format spiffe) || exit

echo "$bundle2" | kubectl "--kubeconfig=$KUBECONFIG1" exec -i spire-server-0 -n spire -- bin/spire-server bundle set -format spiffe -id "spiffe://nsm.cluster2" || exit
echo "$bundle1" | kubectl "--kubeconfig=$KUBECONFIG2" exec -i spire-server-0 -n spire -- bin/spire-server bundle set -format spiffe -id "spiffe://nsm.cluster1" || exit
