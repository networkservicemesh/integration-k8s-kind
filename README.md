# integration-k8s-kind

How to run integration tests locally?
```bash
kind create cluster --config cluster-config.yaml --wait 120s
go test ./...
```
