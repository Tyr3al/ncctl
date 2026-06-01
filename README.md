# netcupctl

`netcupctl` is a Go CLI and library for administering netcup Server Control Panel resources.

The CLI authenticates with the SCP device-code flow, stores an offline refresh token locally, and provides workflow-oriented commands for servers, failover IPs, tasks, rDNS, and other administration tasks.

## Early usage

```sh
netcupctl login
netcupctl servers list
netcupctl failover list
netcupctl failover route --ip 192.0.2.10 --server-id 12345 --wait
```

Use `--json` on commands to get machine-readable output.
