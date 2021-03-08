#!/bin/sh

# Script parameters:
#
# RUNNING_TIME: For how long (in seconds) to let the containers run before
# starting the migration.
# ITERATIONS: How times to run the benchmark.
# OUTPUT_PATH: The path to the directory where to write the log files
# CHAIN_LENGTH: What value to pass to the CHAIN_LENGTH environment variable.
# DUMP_INTERVAL: What value to pass to the DUMP_INTERVAL environment variable.
# POST_RUN_HOOK: The remainder of values passed to the input file will be used
# as a command to run after the containers have been started. Can be used to
# initialize the redis database with the population script.
RUNNING_TIME=$1
ITERATIONS=$2
OUTPUT_PATH=$3
CHAIN_LENGTH=$4
DUMP_INTERVAL=$5
POST_RUN_HOOK=$6

container_name=msc-bench-ping
container_name_1=$container_name-1
container_name_2=$container_name-2

host_1_ip=172.17.0.2
host_2_ip=172.17.0.3

mkdir -p $OUTPUT_PATH
echo Benchmark logs will be written to $OUTPUT_PATH

stdout_log_full() { echo -e $1 | tee -a $OUTPUT_PATH/full.log ; }
print_step() { stdout_log_full "> $1" ; }

run_benchmark() {
	iteration=$1
	# Start containers
	# Sets the ping related options sufficiently high as too disable them, as ping
	# timeouts could potentially ruin benchmarks for large dumps.
	print_step "Starting containers..."
	docker run --rm --privileged \
			--name $container_name_1 \
			--env-file ./.env \
			-e LOG_LEVEL=debug \
			-e CHAIN_LENGTH=$CHAIN_LENGTH -e DUMP_INTERVAL=$DUMP_INTERVAL \
			-e PING_INTERVAL=600000 -e PING_TIMEOUT=600000 -e PING_TIMEOUT_SOURCE=600000 \
			-v $(pwd)/rootfs:/app/rootfs -v $(pwd)/config.json:/app/config.json msc \
			run $container_name > $OUTPUT_PATH/container1_i$iteration.log 2>&1 || \
			{ echo Container 1 already running; exit 1; } &
	sleep 1
	docker run --rm --privileged \
			--name $container_name_2 \
			--env-file ./.env \
			-e LOG_LEVEL=debug \
			-e CHAIN_LENGTH=$CHAIN_LENGTH -e DUMP_INTERVAL=$DUMP_INTERVAL \
			-e PING_INTERVAL=600000 -e PING_TIMEOUT=600000 -e PING_TIMEOUT_SOURCE=600000 \
			-v $(pwd)/rootfs:/app/rootfs -v $(pwd)/config.json:/app/config.json msc \
			join $host_1_ip:1234 > $OUTPUT_PATH/container2_i$iteration.log 2>&1 || \
			{ echo Container 1 already running; exit 1; } &
	[ -n "$POST_RUN_HOOK" ] && $POST_RUN_HOOK

	# -tt prints the time in unix timestamp with fractions
	TCPDUMP_LOG=$OUTPUT_PATH/"$iteration"_tcpdump.log
	tcpdump -i docker0 -tt -l ip proto \\icmp > $TCPDUMP_LOG 2> /dev/null &

	# Sleep for a bit so that containers have the time to start
	print_step "Running..."
	sleep $RUNNING_TIME

	print_step "Initiating migration"
	migration_time=$(date +"%s.%N")
	docker exec $container_name_1 msc migrate $container_name &> /dev/null

	# Wait a while before assuming migration is done
	sleep 5
	print_step "Assuming migration is done"
	print_step "Stopping containers..."
	docker stop $container_name_1 $container_name_2 &> /dev/null
}

TOTAL_MIGRATION_TIME_SUM=0
DOWNTIME_SUM=0

calculate_results() {
	LOG_FILE=$1

	host_1_last=0
	unset host_2_first

	print_step "Calculating results..."

	# Remove all lines that are ping replies (i.e. from the host to the
	# containers) or that was made before the migration request
	lines=$( sed '/ICMP echo reply/d' $TCPDUMP_LOG | \
		awk -v migration_time=$migration_time '{if($1 > migration_time){printf ("%s %s\n", $1, $3)}}')

	while read -r line; do
		timestamp=$(echo $line | awk '{print $1}')
		host=$(echo $line | awk '{print $2}')
		diff=$(echo "$timestamp - $migration_time" | bc -l)
		
		if [ $(echo "$diff > 0" | bc -l) -eq 1 ]; then
			if [ $host = $host_1_ip ]; then
				host_1_last=$timestamp
				continue
			fi

			if [ $host = $host_2_ip ]; then
				host_2_first=$timestamp
				total=$(echo "$host_2_first - $migration_time" | bc -l)
				downtime=$(echo "$host_2_first - $host_1_last" | bc -l)
				TOTAL_MIGRATION_TIME_SUM=$(echo "$TOTAL_MIGRATION_TIME_SUM + $total" | bc -l)
				DOWNTIME_SUM=$(echo "$DOWNTIME_SUM + $downtime" | bc -l)
				break
			fi
		fi
	done < <(echo "$lines")
}

for((i=0; i < $ITERATIONS; i++)); do
	stdout_log_full 
	stdout_log_full "ITERATION $i"
	stdout_log_full ==============================
	run_benchmark $i
	calculate_results $TCPDUMP_LOG

	# Result sanity check
	if [ $(echo "$host_1_last == 0" | bc -l) -eq 1 ]; then 
		stdout_log_full 
		stdout_log_full "Erraneous result, never received last ping from host 1" 
		exit 1
	fi
	if [ -z $host_2_first ]; then 
		stdout_log_full 
		stdout_log_full "Erraneous result, never received ping from host 2" 
		exit 1
	fi

	stdout_log_full "----- Results:"
	stdout_log_full "MIGRATION_TIMESTAMP: $migration_time"
	stdout_log_full "HOST_1_LAST: $host_1_last"
	stdout_log_full "HOST_2_FIRST: $host_2_first"
	stdout_log_full "DOWNTIME: $downtime"
	stdout_log_full "TOTAL_MIGRATION_TIME: $total"
	stdout_log_full ==============================
done

# Print results and parameters to stdout and log file
stdout_log() { echo $1 | tee -a $OUTPUT_PATH/results ; }
print_result() { stdout_log "$1: $2" ; }
AVG_MIGRATION_TIME=$(echo "$TOTAL_MIGRATION_TIME_SUM / $ITERATIONS" | bc -l)
AVG_DOWNTIME=$(echo "$DOWNTIME_SUM / $ITERATIONS" | bc -l)
stdout_log "------------------------RESULTS-----------------------"
stdout_log "RUNNING_TIME=$RUNNING_TIME ITERATIONS=$ITERATIONS CHAIN_LENGTH=$CHAIN_LENGTH DUMP_INTERVAL=$DUMP_INTERVAL POST_RUN_HOOK=\"$POST_RUN_HOOK\""
stdout_log ===============================
print_result "SUM_TOT_MIG_TIME" $TOTAL_MIGRATION_TIME_SUM
print_result "SUM_DOWNTIME" $DOWNTIME_SUM
print_result "AVG_TOT_MIG_TIME" $AVG_MIGRATION_TIME
print_result "AVG_DOWNTIME" $AVG_DOWNTIME
stdout_log ===============================
