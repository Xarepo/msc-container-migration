#!/bin/sh

INPUT_FILE=$1
[ ! -e "$INPUT_FILE" ] && echo "Input file \"$INPUT_FILE\" does not exist" && exit 1

BASE_PATH=.benchmark/ping/$(date -u +"%Y-%m-%dT%H:%M:%S%Z")
index=0
while read -r running_time iterations chain_length dump_interval post_run_hook; do
	OUTPUT_PATH=$BASE_PATH/$index
	sh scripts/benchmarking/ping/start.sh $running_time $iterations $OUTPUT_PATH $chain_length $dump_interval "$post_run_hook"
	index=$((index+1))
done < <(sed '/^[[:space:]]*#/d' $INPUT_FILE)
