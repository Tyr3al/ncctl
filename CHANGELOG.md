# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `ncserver` binary: a server-local CLI with a reduced command set scoped to the server it runs on
- `ncserver identify`: detects the server by querying the SCP API for each local IP address; result is cached in the config file; accepts `--server-id` for manual override
- `ncserver` commands: `status`, `failover list`, `failover route`, `rescue status/enable/disable`, `snapshots list/create`, `rdns get/set/delete`, `tasks wait`
- `server_id` field in the config file, written by `ncserver identify`

### Fixed

- `version` command now reports the installed module version via `runtime/debug.ReadBuildInfo()`; prints `dev` in local builds
- Nil task guard in `writeTasks` to match existing `writeTask` behaviour
- Server resolution now matches by nickname in addition to name
- `--timeout` now controls the overall operation deadline only; individual HTTP requests always time out after 30s, so `--wait` no longer requires a manual `--timeout` override

### Added

- GitHub Actions CI workflow running `go vet` and `go test` on push and pull requests

## [0.1.0] - 2026-06-01

### Added

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
- Server name resolution: all commands accept either a numeric server ID or the server name as shown in the netcup UI (e.g. `v2202508149564377314`)
- `--json` flag for machine-readable output on all commands
- `--yes` / `-y` flag to skip confirmation prompts on destructive operations
- `--timeout` flag for configuring API request timeout
- `--config` flag for specifying a custom config file path
- Apache 2.0 license


[Unreleased]: https://github.com/Tyr3al/ncctl/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/Tyr3al/ncctl/releases/tag/v0.1.0
