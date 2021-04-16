# Example usages

### Communicating with the system from inside the container

The system listens for IPCs on a unix datagram socket, at path `/tmp/msc.sock`.
In order to expose these IPCs to the application running inside the container,
(for example to allow it to determine itself when to migrate), one can bind
mount the system's socket into the container. This socket will then be
available to the application running inside the container which can then send
IPCs to the socket as it would any other unix datagram sockets. The bind mount
should be specified by the OCI-bundle, which can be done by adding the
following object to the `mounts`-array in `config.json`:

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

#### Example

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
