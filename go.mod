module github.com/networkservicemesh/integration-k8s-kind

go 1.15

require (
	github.com/networkservicemesh/integration-tests v0.0.0-20210126135600-ad4f1062b671
	github.com/stretchr/testify v1.6.1
)

replace github.com/networkservicemesh/integration-tests => ../integration-tests
