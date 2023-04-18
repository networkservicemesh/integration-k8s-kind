#!/bin/bash

result_folder=$1

echo showing results from "$result_folder"

folders="$result_folder"/*

for folder in $folders
do
    print_summary.sh "$folder"/results
done
