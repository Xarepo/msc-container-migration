#!/bin/sh

docker export $(docker create busybox) | tar xfC - rootfs
cp ./scripts/benchmarking/ping/ping.sh rootfs/

oci-runtime-tool generate \
	--args "sh" --args "ping.sh" \
	--linux-namespace-remove network > config.json
