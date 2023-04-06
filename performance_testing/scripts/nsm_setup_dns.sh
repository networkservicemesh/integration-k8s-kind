#!/bin/bash

kubectl "--kubeconfig=$KUBECONFIG1" expose service kube-dns -n kube-system --port=53 --target-port=53 --protocol=TCP --name=exposed-kube-dns --type=LoadBalancer
kubectl "--kubeconfig=$KUBECONFIG2" expose service kube-dns -n kube-system --port=53 --target-port=53 --protocol=TCP --name=exposed-kube-dns --type=LoadBalancer

kubectl "--kubeconfig=$KUBECONFIG1" get services exposed-kube-dns -n kube-system -o go-template='{{index (index (index (index .status "loadBalancer") "ingress") 0) "ip"}}' || sleep 10
kubectl "--kubeconfig=$KUBECONFIG1" get services exposed-kube-dns -n kube-system -o go-template='{{index (index (index (index .status "loadBalancer") "ingress") 0) "ip"}}' || exit
echo

kubectl "--kubeconfig=$KUBECONFIG2" get services exposed-kube-dns -n kube-system -o go-template='{{index (index (index (index .status "loadBalancer") "ingress") 0) "ip"}}' || sleep 10
kubectl "--kubeconfig=$KUBECONFIG2" get services exposed-kube-dns -n kube-system -o go-template='{{index (index (index (index .status "loadBalancer") "ingress") 0) "ip"}}' || exit
echo

ip1=$(kubectl "--kubeconfig=$KUBECONFIG1" get services exposed-kube-dns -n kube-system -o go-template='{{index (index (index (index .status "loadBalancer") "ingress") 0) "ip"}}') || exit
if [[ $ip1 == *"no value"* ]]; then 
    hostname1=$(kubectl "--kubeconfig=$KUBECONFIG1" get services exposed-kube-dns -n kube-system -o go-template='{{index (index (index (index .status "loadBalancer") "ingress") 0) "hostname"}}') || exit
    echo hostname1 is $hostname1
    ip1=$(dig +short $hostname1 | head -1) || exit
fi
# if IPv6
if [[ $ip1 =~ ":" ]]; then ip1=[$ip1]; fi

echo Selected externalIP: $ip1 for cluster1

if [[ -z "$ip1" ]]; then echo ip1 is empty; exit 1; fi

ip2=$(kubectl "--kubeconfig=$KUBECONFIG2" get services exposed-kube-dns -n kube-system -o go-template='{{index (index (index (index .status "loadBalancer") "ingress") 0) "ip"}}') || exit
if [[ $ip2 == *"no value"* ]]; then 
    hostname2=$(kubectl "--kubeconfig=$KUBECONFIG2" get services exposed-kube-dns -n kube-system -o go-template='{{index (index (index (index .status "loadBalancer") "ingress") 0) "hostname"}}') || exit
    echo hostname2 is $hostname2
    ip2=$(dig +short $hostname2 | head -1) || exit
fi
# if IPv6
if [[ $ip2 =~ ":" ]]; then ip2=[$ip2]; fi

echo Selected externalIP: $ip2 for cluster2

if [[ -z "$ip2" ]]; then echo ip2 is empty; exit 1; fi

cat > configmap.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: coredns
  namespace: kube-system
data:
  Corefile: |
    .:53 {
        errors
        health {
            lameduck 5s
        }
        ready
        kubernetes cluster.local in-addr.arpa ip6.arpa {
            pods insecure
            fallthrough in-addr.arpa ip6.arpa
            ttl 30
        }
        k8s_external my.cluster1
        prometheus :9153
        forward . /etc/resolv.conf {
            max_concurrent 1000
        }
        loop
        reload 5s
    }
    my.cluster2:53 {
      forward . ${ip2}:53 {
        force_tcp
      }
    }
EOF
kubectl "--kubeconfig=$KUBECONFIG1" apply -f configmap.yaml

cat > custom-configmap.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: coredns-custom
  namespace: kube-system
data:
  server.override: |
    k8s_external my.cluster2
  proxy1.server: |
    my.cluster2:53 {
      forward . ${ip2}:53 {
        force_tcp
      }
    }
EOF

kubectl "--kubeconfig=$KUBECONFIG1" apply -f custom-configmap.yaml


cat > configmap.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: coredns
  namespace: kube-system
data:
  Corefile: |
    .:53 {
        errors
        health {
            lameduck 5s
        }
        ready
        kubernetes cluster.local in-addr.arpa ip6.arpa {
            pods insecure
            fallthrough in-addr.arpa ip6.arpa
            ttl 30
        }
        k8s_external my.cluster2
        prometheus :9153
        forward . /etc/resolv.conf {
            max_concurrent 1000
        }
        loop
        reload 5s
    }
    my.cluster1:53 {
      forward . ${ip1}:53 {
        force_tcp
      }
    }
EOF
kubectl "--kubeconfig=$KUBECONFIG2" apply -f configmap.yaml
cat > custom-configmap.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: coredns-custom
  namespace: kube-system
data:
  server.override: |
    k8s_external my.cluster1
  proxy1.server: |
    my.cluster1:53 {
      forward . ${ip1}:53 {
        force_tcp
      }
    }
EOF
kubectl "--kubeconfig=$KUBECONFIG2" apply -f custom-configmap.yaml

