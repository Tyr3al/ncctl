# Usage Examples

This page collects practical `ncctl` examples for interactive administration and automation.

## Login

Authenticate once on the machine that will run `ncctl`:

```sh
ncctl login
```

By default, credentials are stored in:

```text
~/.config/ncctl/config.json
```

Use a dedicated config file for automation:

```sh
ncctl --config /etc/ncctl/config.json login
```

## Servers

List servers:

```sh
ncctl servers list
```

Show a server by numeric SCP ID:

```sh
ncctl servers get 12345
```

Show a server by the name visible in the web UI:

```sh
ncctl servers get v220000000000000000
```

Commands that take a server accept either the numeric SCP ID or the server name.

## Failover IPs

List all failover IPs:

```sh
ncctl failover list
```

List only IPv4 failover IPs:

```sh
ncctl failover list --family v4
```

List failover IPs assigned to a server by name:

```sh
ncctl failover list --server-id v220000000000000000
```

Route a single failover IP to a server:

```sh
ncctl failover route \
  --server-id v220000000000000000 \
  --ip 203.0.113.10 \
  --wait
```

Route IPv4 and IPv6 failover IPs in one command:

```sh
ncctl --timeout 2m --json failover route \
  --server-id v220000000000000000 \
  --ip 203.0.113.10 \
  --ip 2001:db8:1234::/64 \
  --wait
```

The command exits non-zero if a lookup, route request, or waited task fails.

## keepalived

A minimal keepalived notify script can call one `ncctl` command when the node becomes primary:

```sh
#!/bin/sh
set -eu

NCCTL=/usr/local/bin/ncctl
CONFIG=/etc/ncctl/config.json
SERVER=v220000000000000000

exec "$NCCTL" \
  --config "$CONFIG" \
  --timeout 2m \
  --json \
  failover route \
  --server-id "$SERVER" \
  --ip 203.0.113.10 \
  --ip 2001:db8:1234::/64 \
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

Run the script manually after changing it:

```sh
/usr/local/sbin/ncctl-failover-master.sh
echo $?
```

An exit code of `0` means all requested failover IPs were routed successfully.

## JSON Output

Use `--json` for automation:

```sh
ncctl --json servers list
ncctl --json tasks get task-uuid
```

## Risky Operations

Risky write operations ask for confirmation. Use `--yes` only when the surrounding automation is already safe:

```sh
ncctl --yes snapshots delete v220000000000000000 before-upgrade
ncctl --yes disks format v220000000000000000 vda
```

## rDNS

Set rDNS:

```sh
ncctl rdns set 203.0.113.10 host.example.com
```

Delete rDNS:

```sh
ncctl rdns delete 203.0.113.10
```
