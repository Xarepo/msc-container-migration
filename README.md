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

#### CRIU_TCP_ESTABLISHED

_required: no, default: false_

Whether or not to pass the `AllowOpenTCP` option to go-runc which in turn
passes it as the `--tcp-established` option to CRIU while
checkpointing. Needed for checkpointing TCP connections, see the
[CRIU documentation](https://criu.org/TCP_connection) for more information.
Parsed as a boolean, see
[strconv.ParseBool()](https://golang.org/pkg/strconv/#ParseBool) for valid
formats.

#### DUMP_INTERVAL

_required: no, default: 5_

The length, in seconds, of the intervals between performing dumps.

#### PING_INTERVAL

_required: no, default: 1_

The length, in seconds, of the intervals between sending pings to nodes in the
cluster.

#### PING_TIMEOUT

_required: no, default: 5_

The length, in seconds, of how long to wait for PING RPCs before considering
the source to be down.

## Running

### Example OCI-bundles

While the system should work with any OCI-bundle, how to generate some examples
that may be interesting are provided here.

Dependencies:

- [oci-runtime-tool](https://github.com/opencontainers/runtime-tools)
- [Docker](https://www.docker.com/)

#### Counter

This a minimal example of an OCI-bundle that runs a simple script
that increments an integer every second.

To generate the `rootfs`, any simple filesystem containing a shell should work.
Here, [busybox](https://hub.docker.com/_/busybox/) is used.

```shell
mkdir -p rootfs && docker export $(docker create busybox) | tar xvfC - rootfs
```

Next we need to copy the actual script to execute:

```shell
cp demos/runc/count.sh rootfs/
```

Generating `config.json`:

```shell
oci-runtime-tool generate \
	--args "sh" --args "/count.sh" \
	--linux-namespace-remove network > config.json
```

#### Redis

This example creates an OCI-bundle for a redis database.

```shell
mkdir -p rootfs && docker export $(docker create redis) | tar xvfC - rootfs
```

Generate `config.json`:

```shell
oci-runtime-tool generate \
	--args "redis-server" --args "--save ''" --args "--appendonly no" \
	--linux-namespace-remove network > config.json
```

The `--save ''` and `--appendonly no` options passed to the `--args` options
disables writing the redis database to disk, keeping everything in-memory, as
the system is not able to deal with persistent storage.

The database can populate with junk data for testing using the script
`scripts/redis/redis_populate.lua`, which generates `n` pairs of the form `i=i`
for every `0<=i<n`, where `n` is the first argument passed to script. The script
can be evaluated as follows:

```shell
redis-cli -h <host> --eval scripts/redis/redis_populate.lua <n>
```

### Docker

After creating a OCI-bundle a node can be run in docker in order to simulate
different hosts.
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

## Communicating with the system from inside the container

The system listens for IPCs on a unix datagram socket, at path `/tmp/msc.sock`. In order
to expose these IPCs to the application running inside the container, (for
example to allow it to determine itself when to migrate), one can bind mount
the system's socket into the container. This socket will then be available to
the application running inside the container which can then send IPCs to the
socket as it would any other unix datagram sockets. The bind mount should be specified
by the OCI-bundle, which can be done by adding the following object to the
`mounts`-array in `config.json`:

```json
{
  "destination": "/tmp/msc.sock",
  "source": "/tmp/msc.sock",
  "options": ["bind"]
}
```

**Note:** As previously mentioned the socket is a unix _diagram_ socket, thus
the system needs to make sure it writes to it as such (and not as a stream
socket, etc.). This means that the communication is unidirectional from the
application running inside the container to the system. The reason for using a
_diagram_ socket is that CRIU will not be able to checkpoint both ends of the
socket (even with the `--external [...]` option) as the socket is bind-mounted
and both ends of the socket exist in different namespaces. See the
[CRIU documentation](https://criu.org/External_UNIX_socket) for more
information.

### Example

The following example runs a script inside a container that tells the system to
checkpoint said container. [socat](https://linux.die.net/man/1/socat) is used
to write to the unix socket.

```shell
echo PID: $$

echo sleeping...
sleep 4
echo checkpoiting
printf "CHECKPOINT" | socat - UNIX-SENDTO:/tmp/msc.sock
sleep 4
done sleeping!
```
