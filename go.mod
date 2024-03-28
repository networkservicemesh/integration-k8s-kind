module github.com/networkservicemesh/integration-k8s-kind

go 1.20

require (
	github.com/networkservicemesh/integration-tests v0.0.0-20240402121540-e554b3a9f4b8
	github.com/stretchr/testify v1.8.4
)

replace github.com/networkservicemesh/integration-tests => github.com/glazychev-art/integration-tests v0.0.0-20240404065437-d8f4278dd981

//replace github.com/networkservicemesh/integration-tests => ../integration-tests

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/networkservicemesh/gotestmd v0.0.0-20220628095933-eabbdc09e0dc // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	golang.org/x/sys v0.15.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
