#!/bin/sh

docker build -t redis-ping-bench -f ./scripts/benchmarking/ping/redis_ping.Dockerfile . > /dev/null

rm -rf rootfs && mkdir -p rootfs

# Generate OCI-bundle
docker export $(docker create redis-ping-bench) | tar xfC - rootfs
oci-runtime-tool generate \
	--args "sh" --args "ping_redis.sh" \
	--linux-namespace-remove network > config.json

# Copy scripts
cp ./scripts/benchmarking/ping/ping_redis.sh ./scripts/redis/redis_populate.lua ./scripts/benchmarking/ping/ping.sh rootfs
