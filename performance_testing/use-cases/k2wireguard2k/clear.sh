#!/bin/bash

parent_path=$( cd "$(dirname "$0")" ; pwd -P ) || exit

function k1() { kubectl --kubeconfig "$KUBECONFIG1" "$@" ; }
function k2() { kubectl --kubeconfig "$KUBECONFIG2" "$@" ; }

echo running "$0"

pkill -f "port-forward svc/fortio-service 8080:8080"

# delete without waiting, to delete in parallel
k1 delete ns perf-test-wg --wait=false
k2 delete ns perf-test-wg --wait=false

# wait for everything to be deleted
k1 delete ns perf-test-wg
k2 delete ns perf-test-wg

# previous command may have failed if the setup have failed and not all resources have been deployed
true
