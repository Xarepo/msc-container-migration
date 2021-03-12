#!/bin/bash
running_time=$1

function run_benchmark(){
	container_name=$1
	dump_freq=$2
	running_time=$3
	output_path=$4

	echo
	header=$(printf "Benchmarking %s, dump_freq=%d, running_time=%d\n==========================" \
		$container_name $dump_freq $running_time
	)
	echo "$header"

	docker run -d --rm --privileged \
		--name $container_name \
		--env-file ./.env \
		-e FULLDUMP_FREQ=$dump_freq \
		-v $(pwd)/rootfs:/app/rootfs -v $(pwd)/config.json:/app/config.json msc \
		run msc > /dev/null

	sleep $running_time
	output=$(docker exec -i $container_name bash < scripts/benchmarking/storage/storage.sh)
	echo "$output"
	header_end="=========================="
	printf "%s\n%s\n%s\n" "$header" "$output" "$header_end" > $output_path/$container_name
	echo $header_end

	docker stop $container_name > /dev/null
}

output_path=./.benchmark/$(date -u +"%Y-%m-%dT%H:%M:%S%Z")
mkdir -p $output_path

run_benchmark msc-b1 2 $running_time $output_path
run_benchmark msc-b2 4 $running_time $output_path
run_benchmark msc-b3 8 $running_time $output_path
run_benchmark msc-b4 16 $running_time $output_path
run_benchmark msc-b5 32 $running_time $output_path
