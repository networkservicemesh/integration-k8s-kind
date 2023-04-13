#!/bin/bash

function k1() { kubectl --kubeconfig "$KUBECONFIG1" "$@" ; }
function k2() { kubectl --kubeconfig "$KUBECONFIG2" "$@" ; }

parent_path=$( cd "$(dirname "$0")" ; pwd -P ) || exit

if [ -z "$1" ]; then echo 1st arg 'name' is missing; exit 1; fi
if [ -z "$2" ]; then echo 2nd arg 'result_folder' is missing; exit 1; fi
if [ -z "$3" ]; then echo 3rd arg 'test_iterations' is missing; exit 1; fi
if [ -z "$4" ]; then echo 4th arg 'test_url' is missing; exit 1; fi
if [ -z "$5" ]; then echo 5th arg 'deploy_script' is missing; exit 1; fi
if [ -z "$6" ]; then echo 6th arg 'clear_script' is missing; exit 1; fi
if [ -z "$7" ]; then echo 7th arg 'nsm_version' is missing; exit 1; fi
if [ -z "$8" ]; then echo 8th arg 'nsm_deploy_folder' is missing; exit 1; fi
if [ -z "$9" ]; then echo 9th arg 'qps_list' is missing; exit 1; fi
if [ -z "${10}" ]; then echo 10th arg 'duration' is missing; exit 1; fi
if [ -z "${11}" ]; then echo 10th arg 'connections' is missing; exit 1; fi

test_name=test-$(TZ=UTC date +%F-T%H-%M-%S)-$1
result_folder=$2/$test_name
test_iterations=$3
test_url=$4
deploy_script=$5
clear_script=$6
nsm_version=$7
nsm_deploy_folder=$8
qps_list=$9
duration=${10}
connections=${11}

echo "test_name: $test_name"
echo "result_folder: $result_folder"
echo "test_iterations: $test_iterations"
echo "test_url: $test_url"
echo "deploy_script: $deploy_script"
echo "clear_script: $clear_script"
echo "nsm_version: $nsm_version"
echo "nsm_deploy_folder: $nsm_deploy_folder"
echo "qps_list: $qps_list"
echo "duration: $duration"
echo "connections: $connections"

mkdir -p "$result_folder" || exit

echo running tests for "$test_url"
# for current_qps in $qps3
for current_qps in $qps_list
do
    echo "testing qps $current_qps"
    "$parent_path/run_test_single.sh" \
        "$test_name" \
        "$result_folder" \
        "$test_iterations" \
        "$test_url" \
        "$current_qps" \
        "$connections" \
        "$duration" \
        "$deploy_script" \
        "$clear_script" \
        "$nsm_version" \
        "$nsm_deploy_folder" \
        || exit
done
