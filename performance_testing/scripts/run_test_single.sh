#!/bin/bash

parent_path=$( cd "$(dirname "$0")" ; pwd -P ) || exit

function k1() { kubectl --kubeconfig $KUBECONFIG1 "$@" ; }
function k2() { kubectl --kubeconfig $KUBECONFIG2 "$@" ; }

if [ -z "$1" ]; then echo 1st arg 'test_name' is missing; exit 1; fi
if [ -z "$2" ]; then echo 2nd arg 'result_folder' is missing; exit 1; fi
if [ -z "$3" ]; then echo 3rd arg 'test_iterations' is missing; exit 1; fi
if [ -z "$4" ]; then echo 4th arg 'test_url' is missing; exit 1; fi
if [ -z "$5" ]; then echo 5th arg 'test_qps' is missing; exit 1; fi
if [ -z "$6" ]; then echo 6th arg 'test_connections' is missing; exit 1; fi
if [ -z "$7" ]; then echo 7th arg 'test_duration' is missing; exit 1; fi
if [ -z "$8" ]; then echo 8th arg 'deploy_script' is missing; exit 1; fi
if [ -z "$9" ]; then echo 9th arg 'clear_script' is missing; exit 1; fi
if [ -z "${10}" ]; then echo 10th arg 'nsm_version' is missing; exit 1; fi
if [ -z "${11}" ]; then echo 11th arg 'nsm_deploy_folder' is missing; exit 1; fi

test_name=$1
result_folder=$2
test_iterations=$3
test_url=$4
test_qps=$5
test_connections=$6
test_duration=$7
deploy_script=$8
clear_script=$9
nsm_version=${10}
nsm_deploy_folder=${11}

echo "test_name: $test_name"
echo "result_folder: $result_folder"
echo "test_iterations: $test_iterations"
echo "test_url: $test_url"
echo "test_qps: $test_qps"
echo "test_connections: $test_connections"
echo "test_duration: $test_duration"
echo "deploy_script: $deploy_script"
echo "clear_script: $clear_script"
echo nsm_version: $nsm_version
echo nsm_deploy_folder: $nsm_deploy_folder

echo ------

function makeConfig() {
    url=$1
    qps=$2
    resolution=$3
    connections=$4
    duration=$5
    sed \
        -e "s^<url>^$url^g" \
        -e "s/<qps>/$qps/g" \
        -e "s/<resolution>/$resolution/g" \
        -e "s/<connections>/$connections/g" \
        -e "s/<duration>/$duration/g" \
        "$parent_path/fortio-config-template.json"
}

function captureState() {
    result_folder=$1
    k1 get pod -A -o wide > "$result_folder/pods-k1.log"
    k1 get svc -A -o wide > "$result_folder/svc-k1.log"
    k2 get pod -A -o wide > "$result_folder/pods-k2.log"
    k2 get svc -A -o wide > "$result_folder/svc-k2.log"
}

function runTest() {
    iterations=${1:-3}
    url=$2
    qps=$3
    connections=$4
    duration=$5
    deploy_script=$6
    clear_script=$7
    nsm_version=$8

    config=$(makeConfig $url $qps 0.00005 $connections $duration) || exit
    config_name="q$qps-c$connections-d$duration"

    warmup_results=$result_folder/warmup
    mkdir -p $warmup_results

    deploy_logs=$result_folder/deploy
    mkdir -p $deploy_logs

    echo "config name: $config_name"
    
    echo "measure for $iterations iterations"
    for i in $(seq -w 1 1 "$iterations")
    do
        echo "round $i"
        test_full_name=$test_name-$config_name-$i
        echo deploying nsm...
        "$nsm_deploy_folder/nsm_setup_nsm.sh" > "$deploy_logs/$test_full_name-deploy-nsm.log" "$nsm_version" 2>&1 || exit
        echo deploying apps...
        "$deploy_script" > "$deploy_logs/$test_full_name-deploy-apps.log" "$nsm_version" 2>&1 || exit
        echo doing warmup run...
        curl -s -d "$config" "localhost:8080/fortio/rest/run" > "$warmup_results/$test_full_name-warmup.json"
        echo doing main run...
        curl -s -d "$config" "localhost:8080/fortio/rest/run" > "$result_folder/$test_full_name.json"
        result_code=$?
        echo saving pod layout
        k1 get pod -A -o wide > "$deploy_logs/$test_full_name-k1-pods.log"
        k2 get pod -A -o wide > "$deploy_logs/$test_full_name-k2-pods.log"
        echo clearing apps...
        "$clear_script" > "$deploy_logs/$test_full_name-clear-apps.log" 2>&1
        echo clearing nsm...
        "$nsm_deploy_folder/nsm_clear_nsm.sh" > "$deploy_logs/$test_full_name-clear-nsm.log" 2>&1
        $(exit $result_code) || exit
    done
}

runTest "$test_iterations" "$test_url" "$test_qps" "$test_connections" "$test_duration" "$deploy_script" "$clear_script" "$nsm_version"

