# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - Unreleased

### Added
- `ncserver renew`: forces a token refresh and persists the updated refresh token to disk; prevents offline token expiry from inactivity
- systemd units in `contrib/systemd/`: `ncserver-token-renew.service` and `ncserver-token-renew.timer` for weekly automated token renewal

### Fixed
- Rotated refresh tokens are now persisted to the config file after every API call, preventing authentication failures when the SCP auth server issues a new refresh token

## [0.1.0] - 2026-06-01

### Added
- `ncserver` binary: a server-local CLI with a reduced command set scoped to the server it runs on, intended for automation scripts such as keepalived notify hooks
- `ncserver identify`: detects the server by matching local IP addresses against the SCP API; result is cached as `server_id` in the config file; accepts `--server-id` for manual override
- `ncserver` commands: `login`, `logout`, `whoami`, `identify`, `status`, `failover list`, `failover route`, `rescue status/enable/disable`, `snapshots list/create`, `rdns get/set/delete`, `tasks wait`
- `login` / `logout` / `whoami` commands using OAuth2 device authorization flow with persistent token storage
- Server commands: `list`, `get`, `update` (hostname, nickname, autostart, UEFI), `power`
- Snapshot commands: `list`, `get`, `create`, `delete`, `revert`, `export`, `dryrun`
- Disk commands: `list`, `get`, `format`, `set-driver`, `supported-drivers`
- Interface commands: `create-vlan`, `update`, `delete`
- ISO commands: `attached`, `list`, `attach`, `detach`
- Rescue system commands: `status`, `enable`, `disable`
- Firewall commands: policy `list`/`get`/`create`/`update`/`delete` and interface `get`/`save`/`reapply`/`restore-copied`
- Failover IP routing with `route` command; supports routing multiple IPs to a single destination in one invocation
- rDNS management: `set` and `delete` for IPv4 and IPv6 addresses
- Server name resolution: all commands accept either a numeric SCP ID or the server name shown in the web UI (e.g. `v2202508149564377314`)
- `--json` flag for machine-readable output on all commands
- `--yes` / `-y` flag to skip confirmation prompts on destructive operations
- `--timeout` flag for configuring the overall operation deadline
- `--config` flag for specifying a custom config file path
- Apache 2.0 license

[Unreleased]: https://github.com/Tyr3al/ncctl/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/Tyr3al/ncctl/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/Tyr3al/ncctl/releases/tag/v0.1.0
