module github.com/networkservicemesh/integration-k8s-kind

go 1.16

require (
	github.com/googleapis/gnostic v0.5.1 // indirect
	github.com/networkservicemesh/integration-tests v0.0.0-20211022065259-bd4a1d291346
	github.com/stretchr/testify v1.7.0
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

//github.com/networkservicemesh/gotestmd => github.com/Mixaster995/gotestmd v0.0.0-20211018061647-1f37ae80de01
replace github.com/networkservicemesh/integration-tests => github.com/Mixaster995/integration-tests v0.0.0-20211022074355-2a4ea7f71afb
