module github.com/networkservicemesh/integration-k8s-kind

go 1.16

require (
	github.com/googleapis/gnostic v0.5.1 // indirect
	github.com/networkservicemesh/integration-tests v0.0.0-20220831100231-423de12a55e2
	github.com/stretchr/testify v1.7.0
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/networkservicemesh/integration-tests => github.com/NikitaSkrynnik/integration-tests v0.0.0-20220908040245-0ee858a4949a

// replace github.com/networkservicemesh/integration-tests => ../integration-tests
