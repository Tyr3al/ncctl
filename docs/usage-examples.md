# Command Reference

Full reference for all `ncctl` commands with examples.

---

## servers

### `servers list`

```sh
ncctl servers list
ncctl servers list --limit 50
```

Flags: `--limit N` (default 100)

### `servers get`

```sh
ncctl servers get 12345
ncctl servers get v2202508149564377314
```

Returns full server details including live state (power, CPU, RAM).

### `servers update`

Update one or more server attributes in a single call.

```sh
ncctl servers update v2202508149564377314 --nickname main-leviathan
ncctl servers update v2202508149564377314 --hostname web01.example.com
ncctl servers update v2202508149564377314 --set-autostart --autostart
ncctl servers update v2202508149564377314 --set-uefi --uefi=false
```

Flags:

| Flag | Description |
|------|-------------|
| `--hostname` | New hostname |
| `--nickname` | New nickname (label shown in the web UI) |
| `--set-autostart` | Apply the `--autostart` value |
| `--autostart` | Autostart value (`true` / `false`) |
| `--set-uefi` | Apply the `--uefi` value |
| `--uefi` | UEFI value (`true` / `false`) |

`--set-autostart` and `--set-uefi` are required to distinguish "set to false" from "leave unchanged".

### `servers power`

```sh
ncctl servers power v2202508149564377314 ON
ncctl servers power v2202508149564377314 OFF
ncctl servers power v2202508149564377314 SUSPENDED
```

Flags: `--state-option` — optional SCP state option passed to the API.

This is a risky operation and will ask for confirmation. Use `--yes` to skip.

---

## interfaces

### `interfaces list`

```sh
ncctl interfaces list v2202508149564377314
```

### `interfaces get`

```sh
ncctl interfaces get v2202508149564377314 aa:bb:cc:dd:ee:ff
```

### `interfaces create-vlan`

```sh
ncctl interfaces create-vlan v2202508149564377314 --vlan-id 100
ncctl interfaces create-vlan v2202508149564377314 --vlan-id 100 --driver VIRTIO
```

Flags: `--vlan-id` (required), `--driver` (default `VIRTIO`)

### `interfaces update`

```sh
ncctl interfaces update v2202508149564377314 aa:bb:cc:dd:ee:ff --driver E1000
```

Flags: `--driver` (required)

### `interfaces delete`

```sh
ncctl --yes interfaces delete v2202508149564377314 aa:bb:cc:dd:ee:ff
```

Risky — requires confirmation.

---

## failover

### `failover list`

```sh
ncctl failover list
ncctl failover list --family v4
ncctl failover list --family v6
ncctl failover list --server-id v2202508149564377314
ncctl failover list --ip 203.0.113.10
```

Flags: `--family` (`all` / `v4` / `v6`, default `all`), `--ip`, `--server-id`

### `failover route`

Route one or more failover IPs to a server. The IP family is inferred automatically from the address when `--ip` is used.

Route by IP address:

```sh
ncctl failover route --server-id v2202508149564377314 --ip 203.0.113.10
ncctl failover route --server-id v2202508149564377314 --ip 2001:db8::/64
```

Route multiple IPs in one command:

```sh
ncctl --timeout 2m failover route \
  --server-id v2202508149564377314 \
  --ip 203.0.113.10 \
  --ip 2001:db8::/64 \
  --wait
```

Route by failover IP ID (requires explicit `--family`):

```sh
ncctl failover route --server-id 12345 --id 99 --family v4 --wait
```

Flags:

| Flag | Description |
|------|-------------|
| `--server-id` | Target server ID or name (required) |
| `--ip` | Failover IP or IPv6 prefix; repeat for multiple |
| `--id` | Failover IP ID (alternative to `--ip`) |
| `--family` | `v4` or `v6`; inferred from `--ip` when omitted |
| `--wait` | Poll until the routing task finishes |

`--ip` and `--id` cannot be combined. When `--id` is used, `--family` is required.

The command exits non-zero if any lookup, route request, or waited task fails.

---

## tasks

### `tasks list`

```sh
ncctl tasks list
ncctl tasks list --limit 20
```

### `tasks get`

```sh
ncctl tasks get <uuid>
```

### `tasks wait`

Blocks until the task reaches a terminal state (`FINISHED`, `FAILED`, `CANCELED`):

```sh
ncctl tasks wait <uuid>
```

### `tasks cancel`

```sh
ncctl --yes tasks cancel <uuid>
```

Risky — requires confirmation.

---

## rdns

### `rdns get`

```sh
ncctl rdns get 203.0.113.10
ncctl rdns get 2001:db8::1
```

For IPv6, returns all rDNS entries in the prefix.

### `rdns set`

```sh
ncctl rdns set 203.0.113.10 host.example.com
ncctl rdns set 2001:db8::1 host.example.com
```

### `rdns delete`

```sh
ncctl rdns delete 203.0.113.10
ncctl rdns delete 2001:db8::1
```

---

## snapshots

### `snapshots list`

```sh
ncctl snapshots list v2202508149564377314
```

### `snapshots get`

```sh
ncctl snapshots get v2202508149564377314 before-upgrade
```

### `snapshots create`

```sh
ncctl snapshots create v2202508149564377314 before-upgrade
ncctl snapshots create v2202508149564377314 before-upgrade --online
```

Flags: `--online` — create a snapshot while the server is running.

### `snapshots dryrun`

Check whether creating a snapshot is possible without actually creating one:

```sh
ncctl snapshots dryrun v2202508149564377314
```

Flags: `--body`, `--body-file`

### `snapshots revert`

```sh
ncctl --yes snapshots revert v2202508149564377314 before-upgrade
```

Risky — requires confirmation.

### `snapshots export`

```sh
ncctl snapshots export v2202508149564377314 before-upgrade
```

### `snapshots delete`

```sh
ncctl --yes snapshots delete v2202508149564377314 before-upgrade
```

Risky — requires confirmation.

---

## rescue

### `rescue status`

```sh
ncctl rescue status v2202508149564377314
```

### `rescue enable`

```sh
ncctl --yes rescue enable v2202508149564377314
```

Risky — requires confirmation.

### `rescue disable`

```sh
ncctl --yes rescue disable v2202508149564377314
```

Risky — requires confirmation.

---

## disks

### `disks list`

```sh
ncctl disks list v2202508149564377314
```

### `disks get`

```sh
ncctl disks get v2202508149564377314 vda
```

### `disks supported-drivers`

```sh
ncctl disks supported-drivers v2202508149564377314
```

### `disks set-driver`

```sh
ncctl --yes disks set-driver v2202508149564377314 --body '{"disks":[{"name":"vda","driver":"VIRTIO"}]}'
ncctl --yes disks set-driver v2202508149564377314 --body-file driver-config.json
```

Flags: `--body`, `--body-file`

Risky — requires confirmation.

### `disks format`

Destroys all data on the disk.

```sh
ncctl --yes disks format v2202508149564377314 vda
```

Risky — requires confirmation.

---

## iso

### `iso attached`

Show the currently attached ISO:

```sh
ncctl iso attached v2202508149564377314
```

### `iso list`

List ISO images available for this server:

```sh
ncctl iso list v2202508149564377314
```

### `iso attach`

```sh
ncctl iso attach v2202508149564377314 --iso-id 42
ncctl iso attach v2202508149564377314 --user-iso my-image.iso
ncctl iso attach v2202508149564377314 --iso-id 42 --boot-cdrom
```

Flags: `--iso-id`, `--user-iso`, `--boot-cdrom`

### `iso detach`

```sh
ncctl --yes iso detach v2202508149564377314
```

Risky — requires confirmation.

---

## firewall

### `firewall policies list`

```sh
ncctl firewall policies list
```

### `firewall policies get`

```sh
ncctl firewall policies get 7
```

### `firewall policies create`

```sh
ncctl firewall policies create --body '{"name":"web","inboundRules":[]}'
ncctl firewall policies create --body-file policy.json
```

Flags: `--body`, `--body-file`

### `firewall policies update`

```sh
ncctl --yes firewall policies update 7 --body '{"name":"web-updated"}'
ncctl --yes firewall policies update 7 --body-file policy.json
```

Flags: `--body`, `--body-file`. Risky — requires confirmation.

### `firewall policies delete`

```sh
ncctl --yes firewall policies delete 7
```

Risky — requires confirmation.

### `firewall interface get`

```sh
ncctl firewall interface get v2202508149564377314 aa:bb:cc:dd:ee:ff
```

### `firewall interface save`

Configure the firewall for a specific interface:

```sh
ncctl --yes firewall interface save v2202508149564377314 aa:bb:cc:dd:ee:ff \
  --body '{"inboundRules":[],"outboundRules":[]}'
ncctl --yes firewall interface save v2202508149564377314 aa:bb:cc:dd:ee:ff \
  --body-file interface-firewall.json
```

Flags: `--body`, `--body-file`. Risky — requires confirmation.

### `firewall interface reapply`

```sh
ncctl --yes firewall interface reapply v2202508149564377314 aa:bb:cc:dd:ee:ff
```

Risky — requires confirmation.

### `firewall interface restore-copied`

```sh
ncctl --yes firewall interface restore-copied v2202508149564377314 aa:bb:cc:dd:ee:ff
```

Risky — requires confirmation.

---

## system

### `system ping`

```sh
ncctl system ping
```

### `system maintenance`

```sh
ncctl system maintenance
```

### `system openapi`

Fetch the full SCP OpenAPI document:

```sh
ncctl system openapi
```

### `system openapi-mcp`

```sh
ncctl system openapi-mcp --body '{}'
ncctl system openapi-mcp --body-file mcp-request.json
```

Flags: `--body`, `--body-file`

---

## server

Additional per-server operations that don't fit under `servers`.

### `server logs`

```sh
ncctl server logs v2202508149564377314
ncctl server logs v2202508149564377314 --limit 50 --offset 100
```

Flags: `--limit` (default 100), `--offset` (default 0)

### `server metrics`

Available metric types: `cpu`, `disk`, `network`, `network-packet`

```sh
ncctl server metrics cpu v2202508149564377314
ncctl server metrics cpu v2202508149564377314 --hours 6
ncctl server metrics disk v2202508149564377314
ncctl server metrics network v2202508149564377314
ncctl server metrics network-packet v2202508149564377314
```

Flags: `--hours` (default 24)

### `server guest-agent`

```sh
ncctl server guest-agent v2202508149564377314
```

### `server guest-agent-status`

```sh
ncctl server guest-agent-status v2202508149564377314
```

### `server gpu-driver`

```sh
ncctl server gpu-driver v2202508149564377314
```

### `server image flavours`

List available image flavours for a server:

```sh
ncctl server image flavours v2202508149564377314
```

### `server image setup`

Install an OS image. Destroys the current disk contents.

```sh
ncctl --yes server image setup v2202508149564377314 --body-file image-setup.json
```

Flags: `--body`, `--body-file`. Risky — requires confirmation.

### `server image setup-user`

Install a user-provided image:

```sh
ncctl --yes server image setup-user v2202508149564377314 --body-file user-image-setup.json
```

Flags: `--body`, `--body-file`. Risky — requires confirmation.

### `server storage-optimization`

```sh
ncctl --yes server storage-optimization v2202508149564377314
ncctl --yes server storage-optimization v2202508149564377314 --disk vda --disk vdb
ncctl --yes server storage-optimization v2202508149564377314 --start-after
```

Flags: `--disk` (repeat for multiple disks), `--start-after` (start server after optimization)

Risky — requires confirmation.

---

## user

### `user get`

```sh
ncctl user get
```

### `user update`

```sh
ncctl --yes user update --body '{"language":"en"}'
ncctl --yes user update --body-file user-update.json
```

Flags: `--body`, `--body-file`. Risky — requires confirmation.

### `user logs`

```sh
ncctl user logs
ncctl user logs --limit 50 --offset 0
```

Flags: `--limit` (default 100), `--offset` (default 0)

### `user ssh-keys list`

```sh
ncctl user ssh-keys list
```

### `user ssh-keys create`

```sh
ncctl user ssh-keys create --body '{"name":"laptop","publicKey":"ssh-ed25519 AAAA..."}'
ncctl user ssh-keys create --body-file key.json
```

Flags: `--body`, `--body-file`

### `user ssh-keys delete`

```sh
ncctl --yes user ssh-keys delete 42
```

Risky — requires confirmation.

### `user vlans list`

```sh
ncctl user vlans list
ncctl user vlans list --server-id v2202508149564377314
```

Flags: `--server-id`

### `user vlans get`

```sh
ncctl user vlans get 10
```

### `user vlans global-get`

Get a VLAN by the global (non-user-scoped) endpoint:

```sh
ncctl user vlans global-get 10
```

### `user vlans update`

```sh
ncctl --yes user vlans update 10 --body '{"name":"mgmt"}'
ncctl --yes user vlans update 10 --body-file vlan.json
```

Flags: `--body`, `--body-file`. Risky — requires confirmation.

### `user images` / `user isos`

Both subgroups share the same set of commands. Replace `images` with `isos` for ISO operations.

```sh
ncctl user images list
ncctl user images get my-image.qcow2
ncctl user images prepare-upload my-image.qcow2
ncctl user images prepare-upload my-image.qcow2 --multipart
ncctl user images part-url my-image.qcow2 <upload-id> --part 1
ncctl user images complete-upload my-image.qcow2 <upload-id> \
  --parts '[{"partNumber":1,"etag":"abc123"}]'
ncctl --yes user images delete my-image.qcow2
```

| Command | Description |
|---------|-------------|
| `list` | List all uploaded images / ISOs |
| `get <key>` | Get details for a specific image / ISO |
| `prepare-upload <key>` | Initiate a single or multipart upload |
| `part-url <key> <upload-id>` | Get the pre-signed URL for a multipart part |
| `complete-upload <key> <upload-id>` | Finalise a multipart upload |
| `delete <key>` | Delete an image / ISO (risky) |

Flags for `prepare-upload`: `--multipart`
Flags for `part-url`: `--part` (part number, default 1)
Flags for `complete-upload`: `--parts` (JSON array of `{"partNumber":N,"etag":"..."}`)

---

## Automation

### JSON output

Use `--json` for machine-readable output on any command:

```sh
ncctl --json servers list
ncctl --json tasks get <uuid>
ncctl --json failover list --family v4
```

### Non-interactive mode

Use `--yes` to suppress confirmation prompts for scripts where the surrounding context is already safe:

```sh
ncctl --yes snapshots delete v2202508149564377314 before-upgrade
ncctl --yes disks format v2202508149564377314 vda
```

### keepalived integration

A minimal notify script that routes failover IPs when the node becomes primary:

```sh
#!/bin/sh
set -eu

NCCTL=/usr/local/bin/ncctl
CONFIG=/etc/ncctl/config.json
SERVER=v2202508149564377314

exec "$NCCTL" \
  --config "$CONFIG" \
  --timeout 2m \
  --json \
  failover route \
  --server-id "$SERVER" \
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

    notify_master "/usr/local/sbin/ncctl-failover-master.sh"
}
```

An exit code of `0` means all requested failover IPs were routed successfully.
