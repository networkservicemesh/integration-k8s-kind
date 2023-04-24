#!/bin/bash

parent_path=$( cd "$(dirname "$0")" ; pwd -P ) || exit

result_folder=$1

echo showing results from "$result_folder"

folders="$result_folder/*"

for folder in $folders
do
    "$parent_path/print_summary.sh" "$folder"/results
done
