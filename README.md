# ncctl

License: Apache-2.0

`ncctl` is a Go CLI and library for administering netcup Server Control Panel resources.

> This project is an experiment and is not production ready yet. Review commands carefully before using it against important infrastructure.
>
> This is an unofficial tool. I am not affiliated with netcup GmbH. netcup is a registered trademark of netcup GmbH.

The CLI authenticates with the SCP device-code flow, stores an offline refresh token locally, and provides workflow-oriented commands for servers, failover IPs, tasks, rDNS, and other administration tasks.

## Install

```sh
go install github.com/tyr3al/ncctl/cmd/ncctl@latest
```

## Authenticate

```sh
ncctl login
```

The login command starts the SCP device-code flow, prints the verification URL and user code, and stores the offline refresh token in:

```text
~/.config/ncctl/config.json
```

Use a different config path with `--config`.

## Common Tasks

List servers:

```sh
ncctl servers list
```

Inspect a server with live state:

```sh
ncctl servers get 12345
ncctl servers get v220000000000000000
```

Commands that take a server accept either the numeric SCP ID or the server name shown in the web UI. For flags that are already named `--server-id`, the value can also be a server name.

List failover IPs:

```sh
ncctl failover list
ncctl failover list --family v4
```

Route a failover IP to another server and wait for the async task:

```sh
ncctl failover route --ip 192.0.2.10 --server-id 12345 --wait
```

Route multiple failover IPs in one command, for example from a keepalived transition script:

```sh
ncctl --timeout 2m --json failover route \
  --server-id v220000000000000000 \
  --ip 192.0.2.10 \
  --ip 2001:db8:1234::/64 \
  --wait
```

The command routes each IP sequentially and exits non-zero if any lookup, route request, or waited task fails.

Set and delete rDNS:

```sh
ncctl rdns set 192.0.2.10 host.example.com
ncctl rdns delete 192.0.2.10
```

Use JSON output for automation:

```sh
ncctl --json servers list
```

Risky write operations ask for confirmation. Use `--yes` for non-interactive automation:

```sh
ncctl --yes disks format 12345 vda
ncctl --yes snapshots delete 12345 before-upgrade
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
