package spire

import (
	"os"
	"path/filepath"

	"github.com/networkservicemesh/gotestmd/pkg/suites/shell"
)

type Suite struct {
	shell.Suite
}

func (s *Suite) SetupSuite() {
	dir := filepath.Join(os.Getenv("GOPATH"), "src", "/github.com/networkservicemesh/deployments-k8s/examples/spire")
	r := s.Runner(dir)
	s.T().Cleanup(func() {
		r.Run(`kubectl delete ns spire`)
	})
	r.Run(`kubectl apply -k .`)
	r.Run(`kubectl wait -n spire --timeout=1m --for=condition=ready pod -l app=spire-agent`)
	r.Run(`kubectl wait -n spire --timeout=1m --for=condition=ready pod -l app=spire-server`)
	r.Run(`kubectl exec -n spire spire-server-0 -- \` + "\n" + `/opt/spire/bin/spire-server entry create \` + "\n" + `-spiffeID spiffe://example.org/ns/spire/sa/spire-agent \` + "\n" + `-selector k8s_sat:cluster:nsm-cluster \` + "\n" + `-selector k8s_sat:agent_ns:spire \` + "\n" + `-selector k8s_sat:agent_sa:spire-agent \` + "\n" + `-node`)
}
func (s *Suite) Test() {}
