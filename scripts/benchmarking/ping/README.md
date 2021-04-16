# Scripts

- [full.sh](./full.sh): Runs a full suite of benchmarks based on the input file.
- [init.sh](./init.sh): Initializes an OCI-bundle that simply pings the
  host machine's docker interface.
- [init_redis.sh](./init_redis.sh): Initializes an OCI-bundle that pings the
  host machine's docker interface and runs an instance of a redis database.
- [ping.sh](./ping.sh): The script that actually pings the host machines docker
  interface.
- [ping_redis.sh](./ping_redis.sh): Starts a redis instance and calls
  [ping.sh](./ping.sh).
- [start.sh](./start.sh): Runs a single benchmark case.
- [to_data.sh](./to_data.sh)/[to_csv.sh](./to_csv.sh): Converts a given
  directory of results into a data or csv format that can be used in figures or
  tables.
