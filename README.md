## Environment variables

#### LOG_LEVEL

_required: No, default: info_

The log level to use. Any message below the specified level will not be logged.
See [zerolog documentation](https://github.com/rs/zerolog#leveled-logging) for
available levels.

#### RPC_IP

_required: No, default: localhost_

The IP address to listen for RPCs on.

#### RPC_PORT

_required: No, default: 1234_

The port to listen for RPCs on.

#### DUMP_PATH

_required: yes_

The path to the folder where dumps will be stored, either by direct dumps or
from file transfers from other hosts.

#### FILE_TRANSFER_PORT

_required: no, default: 22_

The port from which to receive file transfers.

## Running

### Docker

A node can be run in docker in order to simulate different hosts.
Building:

```shell
docker build -t msc -f docker/Dockerfile .
```

Docker sets the the cgroup filesystems in `/sys` to read-only. In order for runc
to mount these in the container, the `SYS_CAP_ADMIN` capability needs to be set,
which can be done via the `--privileged` flag to `docker run`.
Also the OCI-bundle's `rootfs` and `config.json` need to be passed to the
container as volumes using the `v` flag.

Running

```shell
docker run --name msc --privileged \
	-v $(pwd)/rootfs:/app/rootfs \
	-v $(pwd)/config.json:/app/config.json \
	msc <options>
```
