# ncctl

License: Apache-2.0

`ncctl` is a Go CLI and library for administering netcup Server Control Panel (SCP) resources.

> This project is an experiment and is not production ready yet. Review commands carefully before using it against important infrastructure.
>
> This is an unofficial tool. I am not affiliated with netcup GmbH. netcup is a registered trademark of netcup GmbH.

## Install

```sh
go install github.com/tyr3al/ncctl/cmd/ncctl@latest   # admin CLI
go install github.com/tyr3al/ncctl/cmd/ncserver@latest # server-local CLI
```

## Authentication

`ncctl` uses the **OAuth2 Device Authorization Grant** (RFC 8628). No password is ever stored.

```sh
ncctl login
```

The login command:

1. Requests a device code from the SCP authorization server.
2. Prints a verification URL and a short user code.
3. You open the URL in a browser and approve the request.
4. `ncctl` polls until approval is granted, then stores a refresh token locally.

The refresh token is written to:

| Platform | Default path |
|----------|-------------|
| Linux    | `~/.config/ncctl/config.json` |
| macOS    | `~/Library/Application Support/ncctl/config.json` |
| Windows  | `%AppData%\ncctl\config.json` |

The file is created with mode `0600` (owner read/write only). The parent directory is created with mode `0700`.

On every subsequent command, `ncctl` silently exchanges the refresh token for a short-lived access token. No re-login is needed until the refresh token expires or is revoked.

Use a custom config path:

```sh
ncctl --config /etc/ncctl/config.json login
```

Check who is currently logged in:

```sh
ncctl whoami
```

Remove stored credentials:

```sh
ncctl logout
```

## Global Flags

These flags are available on every command:

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | OS default | Path to the config file |
| `--api-base-url` | `https://servercontrolpanel.de/scp-core` | SCP API base URL |
| `--auth-base-url` | `https://servercontrolpanel.de` | SCP auth base URL |
| `--timeout` | `0` (no limit) | Overall operation timeout; individual HTTP requests always time out after 30s |
| `--json` | `false` | Write JSON output instead of a table |
| `--yes` / `-y` | `false` | Skip confirmation prompts on destructive operations |

Both base URLs must use `https`. HTTP URLs are rejected to prevent token exposure.

## Server References

Every command that takes a `<server>` argument accepts either:

- The **numeric SCP ID** (e.g. `12345`)
- The **server name** as shown in the web UI (e.g. `v2202501234567890123`)

The same applies to `--server-id` flags.

## Command Reference

A full command reference with examples is in [docs/usage-examples.md](docs/usage-examples.md).

| Command group | What it manages |
|---------------|----------------|
| `servers`     | List, inspect, update, and control server power |
| `interfaces`  | List and manage network interfaces |
| `failover`    | List and route failover IPs |
| `tasks`       | List, inspect, wait for, and cancel async tasks |
| `rdns`        | Get, set, and delete reverse DNS entries |
| `snapshots`   | Create, list, revert, export, and delete snapshots |
| `rescue`      | Enable and disable the rescue system |
| `disks`       | List, inspect, format, and reconfigure disks |
| `iso`         | Attach and detach ISO images |
| `firewall`    | Manage firewall policies and interface rules |
| `system`      | Ping the API, read maintenance info and the OpenAPI document |
| `server`      | Metrics, logs, image setup, guest agent, GPU driver, storage optimization |
| `user`        | Inspect and update user, manage images, ISOs, SSH keys, and VLANs |

## ncserver

`ncserver` is a companion binary for installation on netcup servers. It exposes a reduced command set scoped to the server it runs on and identifies itself automatically via the SCP API.

**Setup** (run once, typically during provisioning):

```sh
ncserver login          # authenticate and store refresh token
ncserver identify       # detect this server by IP, cache its ID
```

**Keep the token alive:**

The SCP offline token expires after ~30 days without use. Install the provided systemd units to renew it weekly:

```sh
cp contrib/systemd/ncserver-token-renew.{service,timer} /etc/systemd/system/
systemctl daemon-reload
systemctl enable --now ncserver-token-renew.timer
```

If auto-detection fails (e.g. the server's IP is not yet registered):

```sh
ncserver identify --server-id v2202501234567890123
```

**Available commands:** `login`, `logout`, `whoami`, `identify`, `status`, `failover list`, `failover route`, `rescue status/enable/disable`, `snapshots list/create`, `rdns get/set/delete`, `tasks wait`

**keepalived example:**

```sh
#!/bin/sh
set -eu
exec ncserver --config /etc/ncserver/config.json \
  failover route \
  --ip 203.0.113.10 \
  --ip 2001:db8::/64 \
  --wait
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
