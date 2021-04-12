#!/bin/sh

# Script parameters:
#
# RUNNING_TIME: For how long (in seconds) to let the containers run before
# starting the migration.
# CHAIN_LENGTH: What value to pass to the CHAIN_LENGTH environment variable.
# DUMP_INTERVAL: What value to pass to the DUMP_INTERVAL environment variable.
# OUTPUT_PATH: The path to the directory where to write the log files
# POST_RUN_HOOK: The remainder of values passed to the input file will be used
# as a command to run after the containers have been started. Can be used to
# initialize the redis database with the population script.

RUNNING_TIME=$1
CHAIN_LENGTH=$2
DUMP_INTERVAL=$3
OUTPUT_PATH=$4
POST_RUN_HOOK=$5

stdout_log() { echo -e "$1" | tee -a $OUTPUT_PATH/result.log ; }

mkdir -p $OUTPUT_PATH
echo Benchmark logs will be written to $OUTPUT_PATH

CONTAINER_NAME=msc-bench-storage

docker run --rm --privileged \
	--name $CONTAINER_NAME \
	--env-file ./.env \
	-e LOG_LEVEL=debug \
	-e CHAIN_LENGTH=$CHAIN_LENGTH -e DUMP_INTERVAL=$DUMP_INTERVAL \
	-v $(pwd)/rootfs:/app/rootfs -v $(pwd)/config.json:/app/config.json msc \
	run msc > $OUTPUT_PATH/container.log 2>&1 || { echo Failed to run container; exit 1; } &
sleep 5 # Give some time for the container to start

[ -n "$POST_RUN_HOOK" ] && $POST_RUN_HOOK

sleep $RUNNING_TIME

output=$(docker exec -i $CONTAINER_NAME sh < scripts/benchmarking/storage/storage.sh)

# Print/log results and parameters
params=$(printf "RUNNING_TIME=%d CHAIN_LENGTH=%d DUMP_INTERVAL=%d POST_RUN_HOOK=\"%s\"" \
	$RUNNING_TIME $CHAIN_LENGTH $DUMP_INTERVAL "$POST_RUN_HOOK"
)
stdout_log ------------------------RESULTS-----------------------
stdout_log "$params"
stdout_log ==========================
stdout_log "$output"
stdout_log ==========================

docker stop $CONTAINER_NAME > /dev/null
# Give some time to stop the container.
# This is needed because the stop command actually finishes before the
# container is completely stopped.
sleep 5 
