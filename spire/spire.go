package spire

import (
	"github.com/edwarnicke/exechelper"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/networkservicemesh/integration-k8s-kind/k8s"
)

func exist() bool {
	client, err := k8s.Client()
	if err != nil {
		return false
	}
	_, err = client.CoreV1().Namespaces().Get("spire", v1.GetOptions{})
	return err == nil
}

// Delete deletes spire namespace with all depended resources
func Delete(options ...*exechelper.Option) error {
	return exechelper.Run("kubectl delete -f ./deployments/spire/spire-namespace.yaml", options...)
}

// Setup setups spire in spire namespace
func Setup(options ...*exechelper.Option) error {
	if exist() {
		return nil
	}
	if err := exechelper.Run("kubectl apply -f ./deployments/spire/spire-namespace.yaml", options...); err != nil {
		return errors.Wrap(err, "cannot create spire namespace")
	}
	if err := exechelper.Run("kubectl apply -f ./deployments/spire", options...); err != nil {
		return errors.Wrap(err, "cannot deploy spire deployments")
	}
	if err := exechelper.Run("kubectl wait -n spire --timeout=60s --for=condition=ready pod -l app=spire-agent", options...); err != nil {
		return errors.Wrap(err, "spire-agent cannot start")
	}
	if err := exechelper.Run("kubectl wait -n spire --timeout=60s --for=condition=ready pod -l app=spire-server", options...); err != nil {
		return errors.Wrap(err, "spire-server cannot start")
	}
	if err := exechelper.Run(`kubectl exec -n spire spire-server-0 --
                                      /opt/spire/bin/spire-server entry create
                                      -spiffeID spiffe://example.org/ns/spire/sa/spire-agent
                                      -selector k8s_sat:cluster:nsm-cluster
                                      -selector k8s_sat:agent_ns:spire
                                      -selector k8s_sat:agent_sa:spire-agent
                                      -node`, options...); err != nil {
		return errors.Wrap(err, "cannot create spire-entry for spire-agent")
	}
	if err := exechelper.Run(`kubectl exec -n spire spire-server-0 --
                                      /opt/spire/bin/spire-server entry create
                                      -spiffeID spiffe://example.org/ns/default/sa/default
                                      -parentID spiffe://example.org/ns/spire/sa/spire-agent
                                      -selector k8s:ns:default \
                                      -selector k8s:sa:default`, options...); err != nil {
		return errors.Wrap(err, "cannot create spire-entry for default namespace")
	}
	return nil
}
