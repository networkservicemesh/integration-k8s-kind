module github.com/networkservicemesh/integration-k8s-kind

go 1.16

require (
	github.com/googleapis/gnostic v0.5.1 // indirect
	github.com/networkservicemesh/integration-tests v0.0.0-20220607095554-76de215d9305
	github.com/stretchr/testify v1.7.0
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/networkservicemesh/integration-tests => github.com/anastasia-malysheva/integration-tests v0.0.0-20220614080538-3d25271ec50d
