rm -rf rootfs && mkdir -p rootfs && \
	docker export $(docker create busybox) | tar xvfC - rootfs && \
	cp demos/runc/count.sh rootfs/ && \
	oci-runtime-tool generate \
		--args "sh" --args "/count.sh" \
		--linux-namespace-remove network > config.json
