## Configuration

The system is configured via environment variables. These are the ones
available:

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
