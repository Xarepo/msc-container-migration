This demo is based on CRIU's [example](https://criu.org/Simple_loop).
Starts a simple process that runs a counter that prints and increments once
every second.

# Instructions

## Non-shell

Start the process in a new session, with stdio redirected

```shell
setsid ./test.sh  < /dev/null &> count.log &
```

Dump the state and kill the process

```shell
sudo criu dump -t $(pgrep test.sh)
```

Restore the state

```shell
sudo criu restore -d
```

## With shell

Start the process

```shell
./count.sh
```

Dump the state and kill the process

```shell
sudo criu dump -t $(pgrep count.sh) --shell-job
```

Restore the state and bring

```shell
sudo criu restore --shell-job
```
