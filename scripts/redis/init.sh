rm -rf rootfs && mkdir -p rootfs && \
	docker export $(docker create redis) | tar xvfC - rootfs &&  \
	oci-runtime-tool generate \
		--args "redis-server" --args "--save ''" --args "--appendonly no" \
		--linux-namespace-remove network > config.json
