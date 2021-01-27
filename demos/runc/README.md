Starts a simple container that runs a counter that prints and increments once
every second.

# Dependencies

- [oci-runtime-tools](https://github.com/opencontainers/runtime-tools)
- [runc](https://github.com/opencontainers/runc)
- [docker](https://www.docker.com/)

# Instructions

Generate the rootfs (based on busybox) for the oci-bundle and copy the script
into the rootfs

```
mkdir -p rootfs && \
	docker export $(docker create busybox) | tar xvfC - rootfs && \
	cp count.sh rootfs/
```

Generate the `config.json` for the oci-bundle

```
oci-runtime-tool generate --args "sh" --args "/count.sh" > config.json
```

Start the container.
Note that the PID printed by the script is 1 (because it's running in a
container, i.e. its new PID-namespace)

```
sudo runc run counter
```

Checkpoint the container (from another terminal)

```
sudo runc checkpoint counter
```

Restore the container

```
sudo runc restore counter
```
