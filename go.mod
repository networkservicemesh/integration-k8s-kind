module github.com/networkservicemesh/integration-k8s-kind

go 1.16

require (
	github.com/googleapis/gnostic v0.5.1 // indirect
	github.com/networkservicemesh/integration-tests v0.0.0-20220706100247-e55849fda82e
	github.com/stretchr/testify v1.7.0
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/networkservicemesh/integration-tests => github.com/MarinaShustowa/integration-tests v0.0.0-20220711122212-3dcffcf7049c
