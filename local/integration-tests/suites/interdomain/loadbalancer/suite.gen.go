// Code generated by gotestmd DO NOT EDIT.
package loadbalancer

import (
	"github.com/stretchr/testify/suite"

	"github.com/networkservicemesh/integration-tests/extensions/base"
)

type Suite struct {
	base.Suite
}

func (s *Suite) SetupSuite() {
	parents := []interface{}{&s.Suite}
	for _, p := range parents {
		if v, ok := p.(suite.TestingSuite); ok {
			v.SetT(s.T())
		}
		if v, ok := p.(suite.SetupAllSuite); ok {
			v.SetupSuite()
		}
	}
	r := s.Runner("../deployments-k8s/examples/interdomain/loadbalancer")
	s.T().Cleanup(func() {
		r.Run(`export KUBECONFIG=$KUBECONFIG1 && kubectl delete ns metallb-system`)
		r.Run(`export KUBECONFIG=$KUBECONFIG2 && kubectl delete ns metallb-system`)
		r.Run(`export KUBECONFIG=$KUBECONFIG3 && kubectl delete ns metallb-system`)
	})
	r.Run(`export KUBECONFIG=$KUBECONFIG1`)
	r.Run(`kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/master/manifests/namespace.yaml` + "\n" + `kubectl create secret generic -n metallb-system memberlist --from-literal=secretkey="$(openssl rand -base64 128)" ` + "\n" + `kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/master/manifests/metallb.yaml`)
	r.Run(`cat > metallb-config.yaml <<EOF` + "\n" + `apiVersion: v1` + "\n" + `kind: ConfigMap` + "\n" + `metadata:` + "\n" + `  namespace: metallb-system` + "\n" + `  name: config` + "\n" + `data:` + "\n" + `  config: |` + "\n" + `    address-pools:` + "\n" + `    - name: default` + "\n" + `      protocol: layer2` + "\n" + `      addresses:` + "\n" + `      - $CLUSTER_CIDR1` + "\n" + `EOF`)
	r.Run(`kubectl apply -f metallb-config.yaml`)
	r.Run(`kubectl wait --for=condition=ready --timeout=5m pod -l app=metallb -n metallb-system`)
	r.Run(`export KUBECONFIG=$KUBECONFIG2`)
	r.Run(`kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/master/manifests/namespace.yaml` + "\n" + `kubectl create secret generic -n metallb-system memberlist --from-literal=secretkey="$(openssl rand -base64 128)" ` + "\n" + `kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/master/manifests/metallb.yaml`)
	r.Run(`cat > metallb-config.yaml <<EOF` + "\n" + `apiVersion: v1` + "\n" + `kind: ConfigMap` + "\n" + `metadata:` + "\n" + `  namespace: metallb-system` + "\n" + `  name: config` + "\n" + `data:` + "\n" + `  config: |` + "\n" + `    address-pools:` + "\n" + `    - name: default` + "\n" + `      protocol: layer2` + "\n" + `      addresses:` + "\n" + `      - $CLUSTER_CIDR2` + "\n" + `EOF`)
	r.Run(`kubectl apply -f metallb-config.yaml`)
	r.Run(`kubectl wait --for=condition=ready --timeout=5m pod -l app=metallb -n metallb-system`)
	r.Run(`export KUBECONFIG=$KUBECONFIG3`)
	r.Run(`kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/master/manifests/namespace.yaml` + "\n" + `kubectl create secret generic -n metallb-system memberlist --from-literal=secretkey="$(openssl rand -base64 128)" ` + "\n" + `kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/master/manifests/metallb.yaml`)
	r.Run(`cat > metallb-config.yaml <<EOF` + "\n" + `apiVersion: v1` + "\n" + `kind: ConfigMap` + "\n" + `metadata:` + "\n" + `  namespace: metallb-system` + "\n" + `  name: config` + "\n" + `data:` + "\n" + `  config: |` + "\n" + `    address-pools:` + "\n" + `    - name: default` + "\n" + `      protocol: layer2` + "\n" + `      addresses:` + "\n" + `      - $CLUSTER_CIDR3` + "\n" + `EOF`)
	r.Run(`kubectl apply -f metallb-config.yaml`)
	r.Run(`kubectl wait --for=condition=ready --timeout=5m pod -l app=metallb -n metallb-system`)
}
func (s *Suite) Test() {}
