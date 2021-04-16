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

### Dump-sizes

The dump sizes are calculated by letting the container specified by the
OCI-bundle running for an amount of time and then iterating over all the dump
directories and summing the sizes of all of the dumps' files. The size is found
using the [stat](https://linux.die.net/man/2/stat) command. Note that the bytes
size is calculated and not the block size.

As in the case of downtime/total migration time the script reads its input
parameters from the file specified by the first argument, i.e.:

```
sh scripts/benchmarking/storage/full.sh <input_file>
```

The input file follows the following format:

```text
# Lines starting with # are treated as comments.
# This is the format of a line:
# <RUNNING_TIME> <CHAIN_LENGTH> <DUMP_INTERVAL> <POST_RUN_HOOK>
10 5 2 redis-cli -h 172.17.0.2 --eval redis_populate.lua 1000
5 3 2 redis-cli -h 172.17.0.2 --eval redis_populate.lua 1000000
```

The script places the resulting log files in `.benchmark/storage/<date>/`.
