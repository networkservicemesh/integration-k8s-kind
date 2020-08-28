module github.com/networkservicemesh/integration-k8s-kind/cmd/kube-helper

go 1.14

require (
	github.com/antonfisher/nested-logrus-formatter v1.3.0
	github.com/edwarnicke/exechelper v1.0.2
	github.com/networkservicemesh/integration-k8s-kind v0.0.0-20201008181911-c1b9b66edc54
	github.com/networkservicemesh/sdk v0.0.0-20201021144352-abb45b1f2a5f
	github.com/sirupsen/logrus v1.7.0
)

replace github.com/networkservicemesh/integration-k8s-kind => ../../
