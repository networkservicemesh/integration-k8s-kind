# integration-k8s-kind

How to run integration tests locally?

## Single cluster tests

1. Create kind cluster:
```bash
kind create cluster --config cluster-config.yaml --wait 120s
```

2. Run tests
```bash
go test -count 1 -timeout 1h -race -v -run Single
```

## Calico single cluster tests

1. Create kind cluster:
```bash
kind create cluster --config cluster-config-calico.yaml
```

2. Apply calico:
```bash
kubectl apply -f https://projectcalico.docs.tigera.io/archive/v3.23/manifests/tigera-operator.yaml
kubectl apply -f https://raw.githubusercontent.com/projectcalico/vpp-dataplane/v3.23.0/yaml/calico/installation-default.yaml
kubectl apply -k calico
```

3. Wait for a calico-vpp rollout:
```bash
kubectl rollout status -n calico-vpp-dataplane ds/calico-vpp-node --timeout=5m
```

4. Run tests:
```bash
go test -count 1 -timeout 1h -race -v -run Calico
```

## Multiple cluster scenario(interdomain tests)
1. Create 3 kind clusters:
```bash
kind create cluster --name kind-1 --config cluster-config-interdomain.yaml --wait 120s
kind create cluster --name kind-2 --config cluster-config-interdomain.yaml --wait 120s
kind create cluster --name kind-3 --config cluster-config-interdomain.yaml --wait 120s
```

2. Save kubeconfig of each cluster(you may choose appropriate location)
```bash
kind get kubeconfig --name kind-1 > /tmp/config1
kind get kubeconfig --name kind-2 > /tmp/config2
kind get kubeconfig --name kind-3 > /tmp/config3
```

3. Run interdomain tests with necessary environment variables set
```bash
export KUBECONFIG1=/tmp/config1
export KUBECONFIG2=/tmp/config2 
export KUBECONFIG3=/tmp/config3 
export CLUSTER1_CIDR="172.18.1.128/25" 
export CLUSTER2_CIDR="172.18.2.128/25"
export CLUSTER3_CIDR="172.18.3.128/25"
go test -count 1 -timeout 1h -race -v -run Interdomain
```
