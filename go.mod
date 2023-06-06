module github.com/networkservicemesh/integration-k8s-kind

go 1.18

require (
	github.com/networkservicemesh/integration-tests v0.0.0-20230523103420-a83ce8480e01
	github.com/stretchr/testify v1.7.0
)

replace github.com/networkservicemesh/integration-tests => github.com/NikitaSkrynnik/integration-tests v0.0.0-20230606140720-a53bce483689

// replace github.com/networkservicemesh/integration-tests => ../integration-tests

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/networkservicemesh/gotestmd v0.0.0-20220628095933-eabbdc09e0dc // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	golang.org/x/sys v0.0.0-20211116061358-0a5406a5449c // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
)
