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

#### MIGRATION_TARGET

_required: Yes_

The address of the host to which migration should be prepared, i.e. dumps sent.
Must be in the form of `<ip>:<port>`.

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
Also the OCI-bundle's `rootfs` needs to be passed to the container using the `-v`
flag.

Running

```shell
docker run --name msc --privileged -v $(pwd):/rootfs:/app/rootfs msc
```
