#!/bin/sh

INPUT_FILE=$1
[ ! -e "$INPUT_FILE" ] && echo "Input file \"$INPUT_FILE\" does not exist" && exit 1

BASE_PATH=.benchmark/storage/$(date -u +"%Y-%m-%dT%H:%M:%S%Z")
index=0
while read -r running_time chain_length dump_interval post_run_hook; do
	OUTPUT_PATH=$BASE_PATH/$index
	sh scripts/benchmarking/storage/start.sh $running_time $chain_length $dump_interval $OUTPUT_PATH "$post_run_hook"
	index=$((index+1))
done < <(sed '/^[[:space:]]*#/d' $INPUT_FILE)
