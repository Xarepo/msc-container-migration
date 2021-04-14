## Environment variables

#### LOG_LEVEL

_required: No, default: `info`_

The log level to use. Any message below the specified level will not be logged.
See [zerolog documentation](https://github.com/rs/zerolog#leveled-logging) for
available levels.

#### RPC_IP

_required: No, default: `localhost`_

The IP address to listen for RPCs on.

#### RPC_PORT

_required: No, default: `1234`_

The port to listen for RPCs on.

#### DUMP_PATH

_required: no, default: `/dumps`_

The path to the folder where dumps will be stored, either by direct dumps or
from file transfers from other hosts.

#### FILE_TRANSFER_PORT

_required: no, default: `22`_

The port from which to receive file transfers.

#### SSH_USER

_required: yes_

The name of the user to use when authenticating with ssh during file transfer.

#### SSH_USER

_required: yes_

The password of the user to use when authenticating with ssh during file
transfer.

#### CRIU_TCP_ESTABLISHED

_required: no, default: `false`_

Whether or not to pass the `AllowOpenTCP` option to go-runc which in turn
passes it as the `--tcp-established` option to CRIU while
checkpointing. Needed for checkpointing TCP connections, see the
[CRIU documentation](https://criu.org/TCP_connection) for more information.
Parsed as a boolean, see
[strconv.ParseBool()](https://golang.org/pkg/strconv/#ParseBool) for valid
formats.

#### ENABLE_CONTINOUS_DUMPING

_required: no, default: `true`_

If set to false no dumps will be made during the runtime before a migration is
initialized. Migrations will work as normal (but will effectively become
regular pre-copy migrations), but failover will not work.
Only intended to use for demonstration purposes.

#### DUMP_INTERVAL

_required: no, default: `5`_

The length, in seconds, of the intervals between performing dumps.

#### CHAIN_LENGTH

_required: no, default: `3`_

The length of each dump chain. The last dump of each chain will a full dump
from which the process can be recovered. If `DUMP_INTERVAL` is set to `m`
seconds and `CHAIN_LENGTH` is set to `n` then every `m*n`th second a full dump
will be made.

#### PING_INTERVAL

_required: no, default: `1`_

The length, in seconds, of the intervals between sending pings to nodes in the
cluster.

#### PING_TIMEOUT

_required: no, default: `5`_

The length, in seconds, of how long to wait for PING RPCs before considering
the source to be down.

#### PING_TIMEOUT_SOURCE

_required: no, default: `3`_

The length, in seconds, of how long to the source waits for the reply for any
PING RPC that it sends, before considering the target to be down.

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
[alpine/socat](https://hub.docker.com/r/alpine/socat) can be used as a image.

```shell
echo PID: $$

echo sleeping...
sleep 4
echo checkpointing
printf "CHECKPOINT" | socat - UNIX-SENDTO:/tmp/msc.sock
sleep 4
echo done sleeping!
```

## Benchmarking

Benchmarking scripts are provided in the directory `scripts/benchmarking/`.
There are two different aspects of the system to benchmark: disk usage (i.e.
size of the dumps) and the migration/down-time.

### Migration/down-time

This is how migration/down-time is measured and calculated.

1.  The benchmarking scripts work by starting two docker containers, simulating
    remote hosts. The first of these containers starts the system and the
    second one joins the cluster at the first container. The benchmarking
    OCI-bundle runs a script that not only starts a redis database, but also
    pings the docker network interface (typically located at `172.17.0.1`)
    every 10th millisecond.
2.  The benchmarking script then starts an instance of
    [tpcdump](https://www.tcpdump.org/) that listens for ICMP (i.e. ping)
    messages on that interface.
3.  After letting the containers/system run for a bit, it tells the first
    container to migrate and saves the current time, `t(m)` (`m` for migration).
4.  The scripts waits to let the migration finish and searches the `tcpdump`
    log file for two values. The first of which, `t(p1_last)`, is the time of
    the last ping sent from the first host. The second value, `t(p2_first)`, is
    the time of the first ping send from the second host.
5.  The total migration time, `t(tot)` is calculated as
    `t(tot)=t(p2_first)-t(m)` and the downtime, `t(down)`, is calculated as
    `t(down)=t(p2_first)-t(p1_last)`.

The system runs the current OCI-bundle specified by `rootfs` and `config.json`.
Scripts to initialize the benchmarking OCI-bundle can be found in
`scripts/benchmarking/ping`.

#### Running

Dependencies:

- [Docker](https://www.docker.com/)
- [tcpdump](https://www.tcpdump.org/)
- [bc](https://linux.die.net/man/1/bc)

The script used to run the bechmark is `scripts/benchmarking/ping/full.sh`.
It reads its input parameters from the file specified by the first argument,
i.e:

```shell
sh scripts/benchmarking/ping/full.sh <input_file>
```

Each line of the input file specifies the input parameters for a run of the
benchmark. This means that the system can be benchmarked with different
parameters by specifying different parameters on different lines. Each line
contains whitespace separated values. Please refer to
[start.sh](scripts/benchmarking/ping/start.sh) for a description of the values.

Example:

```text
# Lines starting with # are treated as comments.
# This is the format of a line:
# <RUNNING_TIME> <ITERATIONS> <CHAIN_LENGTH> <DUMP_INTERVAL> <POST_RUN_HOOK>
10 5 5 2 redis-cli -h 172.17.0.2 --eval redis_populate.lua 1000
5 5 3 2 redis-cli -h 172.17.0.2 --eval redis_populate.lua 1000000
```

The script places the resulting log files in `.benchmark/ping/<date>/`.
