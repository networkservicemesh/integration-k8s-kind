# integration-k8s-kind

How to run integration tests locally?
```bash
kind create cluster --name nsm --config cluster-config.yaml --wait 120s
go test ./...
```
