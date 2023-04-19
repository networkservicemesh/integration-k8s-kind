#!/bin/bash

result_folder=$1

echo showing results from "$result_folder"

function printForFile() {
    filepath=$1
    
    qps=$(< "$filepath" jq .ActualQPS)
    count=$(< "$filepath" jq .DurationHistogram.Count)
    min=$(< "$filepath" jq .DurationHistogram.Min)
    max=$(< "$filepath" jq .DurationHistogram.Max)
    avg=$(< "$filepath" jq .DurationHistogram.Avg)
    p50=$(< "$filepath" jq .DurationHistogram.Percentiles[0].Value)
    p99=$(< "$filepath" jq .DurationHistogram.Percentiles[4].Value)
    
    echo -----------------------
    echo From "$filepath"
    echo QPS: "$qps"
    echo Total queries: "$count"
    echo Min: "$(awk '{print $1 * 1000}' <<<"$min")" ms
    echo Max: "$(awk '{print $1 * 1000}' <<<"$max")" ms
    echo Avg: "$(awk '{print $1 * 1000}' <<<"$avg")" ms
    echo p50: "$(awk '{print $1 * 1000}' <<<"$p50")" ms
    echo p99: "$(awk '{print $1 * 1000}' <<<"$p99")" ms
    echo -----------------------
}

files="$result_folder/*"

for file in $files
do
    printForFile "$file"
done
