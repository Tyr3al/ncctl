# ncserver Command Reference

`ncserver` is a companion binary intended to run on a netcup server. It exposes a reduced command set scoped to the server it runs on, identifies itself automatically via the SCP API, and is designed for use in automation scripts such as keepalived notify hooks.

For full administrative access across all servers, use [`ncctl`](ncctl.md).

---

## Global Flags

These flags are available on every `ncserver` command:

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | OS default | Path to the config file |
| `--timeout` | `0` (no limit) | Overall operation timeout; individual HTTP requests always time out after 30s |
| `--json` | `false` | Write JSON output instead of a table |
| `--yes` / `-y` | `false` | Skip confirmation prompts on destructive operations |

It is recommended to use a dedicated config file for `ncserver` (e.g. `--config /etc/ncserver/config.json`) to keep its credentials separate from `ncctl`.

---

## Setup

Run these once, typically during server provisioning:

```sh
ncserver --config /etc/ncserver/config.json login      # authenticate and store refresh token
ncserver --config /etc/ncserver/config.json identify   # detect this server by IP, cache its ID
```

---

## version

Print the semantic version, commit hash, and build date:

```sh
ncserver version
```

Example output:

```
ncserver v0.3.0
commit: 2311e32
built:  2026-06-03T10:00:00Z
```

---

## login / logout / whoami

Authenticate using the OAuth2 Device Authorization Grant. No password is ever stored — only a refresh token.

```sh
ncserver login
```

`login` prints a verification URL and a short user code. Open the URL in a browser, approve the request, and `ncserver` will store the refresh token locally.

```sh
ncserver whoami
```

Print the currently authenticated user's ID, username, and email.

```sh
ncserver logout
```

Revokes the refresh token on the authorization server and removes the local credentials. Revocation invalidates the token server-side so it cannot be used from other devices. If the revocation request fails (e.g. no network), a warning is printed but the local credentials are still removed.

---

## renew

Force a token refresh and persist the updated refresh token to disk. Prevents the offline token from expiring due to inactivity (tokens typically expire after ~30 days without use).

```sh
ncserver renew
```

Run this weekly via the provided systemd timer:

```sh
cp contrib/systemd/ncserver-token-renew.{service,timer} /etc/systemd/system/
systemctl daemon-reload
systemctl enable --now ncserver-token-renew.timer
```

---

## identify

Detect this server by matching local IP addresses against the SCP API and cache its ID in the config file. All subsequent commands use the cached ID automatically.

```sh
ncserver identify
```

If auto-detection fails (e.g. the server's IP is not yet registered), provide the ID manually:

```sh
ncserver identify --server-id v2202501234567890123
```

Flags: `--server-id` — server ID or name to use instead of auto-detection.

---

## status

Show this server's current state:

```sh
ncserver status
```

Returns ID, name, hostname, nickname, and power state.

---

## failover

### `failover list`

List failover IPs currently routed to this server:

```sh
ncserver failover list
```

### `failover route`

Route one or more failover IPs to this server. The IP family is inferred automatically from the address.

```sh
ncserver failover route --ip 203.0.113.10
ncserver failover route --ip 2001:db8::/64
ncserver failover route --ip 203.0.113.10 --ip 2001:db8::/64 --wait
```

Route by failover IP ID (requires explicit `--family`):

```sh
ncserver failover route --id 99 --family v4 --wait
```

Flags:

| Flag | Description |
|------|-------------|
| `--ip` | Failover IP or IPv6 prefix; repeat for multiple |
| `--id` | Failover IP ID (alternative to `--ip`) |
| `--family` | `v4` or `v6`; inferred from `--ip` when omitted |
| `--wait` | Poll until the routing task finishes |

`--ip` and `--id` cannot be combined. When `--id` is used, `--family` is required.

---

## rescue

### `rescue status`

```sh
ncserver rescue status
```

### `rescue enable`

```sh
ncserver --yes rescue enable
```

Risky — requires confirmation.

### `rescue disable`

```sh
ncserver --yes rescue disable
```

Risky — requires confirmation.

---

## snapshots

### `snapshots list`

```sh
ncserver snapshots list
```

### `snapshots create`

```sh
ncserver snapshots create before-upgrade
ncserver snapshots create before-upgrade --online
```

Flags: `--online` — create a snapshot while the server is running.

---

## rdns

### `rdns get`

```sh
ncserver rdns get 203.0.113.10
ncserver rdns get 2001:db8::1
```

For IPv6, returns all rDNS entries in the prefix.

### `rdns set`

```sh
ncserver rdns set 203.0.113.10 host.example.com
ncserver rdns set 2001:db8::1 host.example.com
```

### `rdns delete`

```sh
ncserver rdns delete 203.0.113.10
ncserver rdns delete 2001:db8::1
```

---

## tasks wait

Block until an async task reaches a terminal state (`FINISHED`, `FAILED`, `CANCELED`):

```sh
ncserver tasks wait <uuid>
```

---

## Automation

### JSON output

Use `--json` for machine-readable output on any command:

```sh
ncserver --json status
ncserver --json failover list
```

### keepalived integration

A minimal notify script that routes failover IPs to this server when it becomes the VRRP primary:

```sh
#!/bin/sh
set -eu
exec ncserver \
  --config /etc/ncserver/config.json \
  --timeout 2m \
  failover route \
  --ip 203.0.113.10 \
  --ip 2001:db8::/64 \
  --wait
```

Example keepalived snippet:

```conf
vrrp_instance VI_1 {
    state BACKUP
    interface eth0
    virtual_router_id 51
    priority 100
    advert_int 1

    authentication {
        auth_type PASS
        auth_pass replace-me
    }

    notify_master "/usr/local/sbin/ncserver-failover-master.sh"
}
```

An exit code of `0` means all requested failover IPs were routed successfully.
