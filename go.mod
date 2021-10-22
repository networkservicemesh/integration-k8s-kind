module github.com/networkservicemesh/integration-k8s-kind

go 1.16

require (
	github.com/googleapis/gnostic v0.5.1 // indirect
	github.com/networkservicemesh/integration-tests v0.0.0-20211020150021-720af7b25bb9
	github.com/stretchr/testify v1.7.0
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace (
	github.com/networkservicemesh/gotestmd => github.com/Mixaster995/gotestmd v0.0.0-20211022102757-ca2cb4b9f76d
	github.com/networkservicemesh/integration-tests => github.com/Mixaster995/integration-tests v0.0.0-20211022103710-9b80da93825b
)
