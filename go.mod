module github.com/networkservicemesh/integration-k8s-kind

go 1.16

require (
	github.com/googleapis/gnostic v0.5.1 // indirect
	github.com/networkservicemesh/integration-tests v0.0.0-20211014105710-40c3f2ba16b0
	github.com/stretchr/testify v1.7.0
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace (
	github.com/networkservicemesh/gotestmd => github.com/Mixaster995/gotestmd v0.0.0-20211018061647-1f37ae80de01
	github.com/networkservicemesh/integration-tests => github.com/Mixaster995/integration-tests v0.0.0-20211018061752-82f963696869
)
