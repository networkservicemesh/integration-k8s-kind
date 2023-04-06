#!/bin/bash

echo CLUSTER1_CIDR is "'$CLUSTER1_CIDR'"
echo CLUSTER2_CIDR is "'$CLUSTER2_CIDR'"

if [[ ! -z $CLUSTER1_CIDR ]]; then
    kubectl "--kubeconfig=$KUBECONFIG1" apply -f https://raw.githubusercontent.com/metallb/metallb/v0.12.1/manifests/namespace.yaml
    kubectl "--kubeconfig=$KUBECONFIG1" apply -f https://raw.githubusercontent.com/metallb/metallb/v0.12.1/manifests/metallb.yaml
    cat > metallb-config.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: metallb-system
  name: config
data:
  config: |
    address-pools:
    - name: default
      protocol: layer2
      addresses:
      - $CLUSTER1_CIDR
EOF
    kubectl "--kubeconfig=$KUBECONFIG1" apply -f metallb-config.yaml
    kubectl "--kubeconfig=$KUBECONFIG1" wait --for=condition=ready --timeout=5m pod -l app=metallb -n metallb-system
fi

if [[ ! -z $CLUSTER2_CIDR ]]; then
    kubectl "--kubeconfig=$KUBECONFIG2" apply -f https://raw.githubusercontent.com/metallb/metallb/v0.12.1/manifests/namespace.yaml
    kubectl "--kubeconfig=$KUBECONFIG2" apply -f https://raw.githubusercontent.com/metallb/metallb/v0.12.1/manifests/metallb.yaml
    cat > metallb-config.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: metallb-system
  name: config
data:
  config: |
    address-pools:
    - name: default
      protocol: layer2
      addresses:
      - $CLUSTER2_CIDR
EOF
    kubectl "--kubeconfig=$KUBECONFIG2" apply -f metallb-config.yaml
    kubectl "--kubeconfig=$KUBECONFIG2" wait --for=condition=ready --timeout=5m pod -l app=metallb -n metallb-system
fi
