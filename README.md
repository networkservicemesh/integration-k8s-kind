# integration-k8s-kind

How to run integration tests locally?

## Single cluster tests

1. Create kind cluster:
```bash
kind create cluster --config cluster-config.yaml --wait 120s
```

2. Run tests
```bash
go test -count 1 -timeout 1h -race -v -run "[a-zA-Z]*Single[a-zA-Z]*"
```

## Multiple cluster scenario(interdomain tests)
1. Create 3 kind clusters:
```bash
kind create cluster --name interdomain-kind-1 --config cluster-config-interdomain.yaml --wait 120s
kind create cluster --name interdomain-kind-2 --config cluster-config-interdomain.yaml --wait 120s
kind create cluster --name interdomain-kind-3 --config cluster-config-interdomain.yaml --wait 120s
```

2. Save kubeconfig of each cluster(you may choose appropriate location)
```bash
kind get kubeconfig --name interdomain-kind-1 > /tmp/config1
kind get kubeconfig --name interdomain-kind-2 > /tmp/config2
kind get kubeconfig --name interdomain-kind-3 > /tmp/config3
```

3. Run interdomain tests with necessary environment variables set
```bash
KUBECONFIG1==/tmp/config1 \
KUBECONFIG2==/tmp/config2 \
KUBECONFIG3==/tmp/config3 \
CLUSTER1_CIDR="172.18.1.128/25" \
CLUSTER2_CIDR="172.18.2.128/25"\
CLUSTER3_CIDR="172.18.3.128/25"\
bash -c 'go test -count 1 -timeout 1h -race -v -run "[a-zA-Z]*Interdomain[a-zA-Z]*"'
```
