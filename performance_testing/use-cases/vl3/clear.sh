#!/bin/bash

parent_path=$( cd "$(dirname "$0")" ; pwd -P ) || exit

function k1() { kubectl --kubeconfig "$KUBECONFIG1" "$@" ; }
function k2() { kubectl --kubeconfig "$KUBECONFIG2" "$@" ; }

echo running "$0"

pkill -f "port-forward svc/fortio-service 8080:8080"

# delete without waiting, to delete in parallel
k1 delete -k "$parent_path/vl3-dns" --wait=false
k1 delete ns perf-test-vl3 --wait=false
k2 delete ns perf-test-vl3 --wait=false

# wait for everything to be deleted
k1 delete -k "$parent_path/vl3-dns"
k1 delete ns perf-test-vl3
k2 delete ns perf-test-vl3

# previous command may have failed if the setup have failed and not all resources have been deployed
true
