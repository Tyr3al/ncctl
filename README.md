# netcupctl

`netcupctl` is a Go CLI and library for administering netcup Server Control Panel resources.

> This project is an experiment and is not production ready yet. Review commands carefully before using it against important infrastructure.

The CLI authenticates with the SCP device-code flow, stores an offline refresh token locally, and provides workflow-oriented commands for servers, failover IPs, tasks, rDNS, and other administration tasks.

## Install

```sh
go install github.com/tyr3al/netcup-api/cmd/netcupctl@latest
```

## Authenticate

```sh
netcupctl login
```

The login command starts the SCP device-code flow, prints the verification URL and user code, and stores the offline refresh token in:

```text
~/.config/netcupctl/config.json
```

Use a different config path with `--config`.

## Common Tasks

List servers:

```sh
netcupctl servers list
```

Inspect a server with live state:

```sh
netcupctl servers get 12345
```

List failover IPs:

```sh
netcupctl failover list
netcupctl failover list --family v4
```

Route a failover IP to another server and wait for the async task:

```sh
netcupctl failover route --ip 192.0.2.10 --server-id 12345 --wait
```

Set and delete rDNS:

```sh
netcupctl rdns set 192.0.2.10 host.example.com
netcupctl rdns delete 192.0.2.10
```

Use JSON output for automation:

```sh
netcupctl --json servers list
```

Risky write operations ask for confirmation. Use `--yes` for non-interactive automation:

```sh
netcupctl --yes disks format 12345 vda
netcupctl --yes snapshots delete 12345 before-upgrade
```

## Library

```go
auth, err := netcup.NewAuthClient(netcup.DefaultAuthBaseURL)
if err != nil {
    // handle error
}
source := netcup.NewRefreshTokenSource(auth, refreshToken)
client, err := netcup.NewClient(netcup.DefaultAPIBaseURL, netcup.WithTokenSource(source))
if err != nil {
    // handle error
}
servers, err := client.ListServers(ctx, netcup.ListServersOptions{Limit: 100})
```
