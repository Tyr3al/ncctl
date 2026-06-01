# Agents.md

Context for AI coding agents working on this repository.

## Project

`ncctl` is a Go CLI and library for the netcup Server Control Panel (SCP) REST API. It is an **unofficial** tool. The module path is `github.com/tyr3al/ncctl`.

- Go 1.23, single external dependency: `github.com/spf13/cobra`
- Apache-2.0 license
- OpenAPI spec for the SCP API: `assets/netcup-openapi-260601.json`

## Repository Layout

```
cmd/ncctl/          Entry point — calls cli.NewRootCommand().Execute()
cmd/ncserver/       Entry point — calls cli.NewServerRootCommand().Execute()
internal/
  cli/              All Cobra command definitions and CLI helpers
    root.go         Global flags, command tree wiring
    readonly_commands.go   servers, interfaces, failover, tasks, rdns
    broad_commands.go      servers update/power, snapshots, disks, ISO,
                           rescue, firewall, interfaces write, rDNS write
    api_commands.go        system, server extras, user, vlans, SSH keys,
                           image/ISO upload
    auth_commands.go       login, logout, whoami
    failover_route.go      failover route (multi-IP logic)
    ncserver.go            NewServerRootCommand + all ncserver commands
    server_ref.go          resolveServerID — numeric ID or name lookup
    app.go                 newApp, apiClient, authClient, HTTPS validation
    context.go             commandOptions helper
    confirm.go             confirmRisky — interactive --yes guard
    format.go              writeTable, writeJSON, stringPtrValue
    json_flags.go          parseJSONObject, parseJSONObjectFile, etc.
  config/
    config.go       Load/Save/Remove config.json (refresh token, URLs, ServerID)
pkg/netcup/
  client.go         HTTP client, DoJSON, DoMergePatch, token injection
  auth.go           AuthClient — device flow, token refresh, UserInfo
  api.go            Read-only API methods (ListServers, GetServer, …)
  api_writes.go     Mutating API methods (PatchServer, RouteFailover, …)
  api_missing.go    Remaining API surface (system, metrics, user assets, …)
  models.go         All API types
assets/
  netcup-openapi-260601.json   Official SCP OpenAPI spec (reference)
docs/
  usage-examples.md   Full command reference
```

## Build & Test

```sh
go build ./...
go test ./...
go install ./cmd/ncctl/
go install ./cmd/ncserver/
```

No code generation, no Makefile. All tests are pure Go unit/integration tests against in-process HTTP stubs.

## Architecture

### Two-layer design

1. **`pkg/netcup`** — a standalone Go library. No CLI concerns. Can be imported independently. All methods accept `context.Context` and return typed values or errors.
2. **`internal/cli`** — thin Cobra wrappers that call the library, format output, and handle global flags. Shared by both `ncctl` and `ncserver`.

### Two binaries

- **`ncctl`** — full admin CLI. Exposes the complete SCP API surface. Intended for operators.
- **`ncserver`** — server-local CLI. Reduced command set scoped to the server it runs on. Identifies itself via IP address lookup (`identifyServerByIP`). The resolved server ID is stored as `server_id` in the config file. Intended to run on the server itself (e.g. keepalived scripts).

### HTTP client (`pkg/netcup/client.go`)

- `DoJSON` — sends `Content-Type: application/json`
- `DoMergePatch` — sends `Content-Type: application/merge-patch+json` (required for PATCH endpoints)
- Bearer token is injected by a `TokenSource` interface; the CLI wires in `RefreshTokenSource`
- Base URLs must use `https` (enforced in `internal/cli/app.go:newApp`)

### Authentication (`pkg/netcup/auth.go`)

OAuth2 Device Authorization Grant (RFC 8628), OAuth client ID `"scp"`. The CLI stores only the **refresh token** in the config file. On each command, `RefreshTokenSource.Token()` exchanges it for a short-lived access token using a mutex-protected cache.

### Config file (`internal/config/config.go`)

Written with `os.WriteFile(..., 0o600)` and parent directory `0o700`. Fields:

```json
{
  "api_base_url": "...",
  "auth_base_url": "...",
  "user_id": 12345,
  "refresh_token": "...",
  "server_id": 67890
}
```

`server_id` is written by `ncserver identify` and read by all ncserver commands. It is omitted from configs that have never been identified.

Default path via `os.UserConfigDir()`: `~/.config/ncctl/config.json` (Linux), `~/Library/Application Support/ncctl/config.json` (macOS), `%AppData%\ncctl\config.json` (Windows).

### Server reference resolution (`internal/cli/server_ref.go`)

`resolveServerID(ctx, client, ref)` accepts either a numeric string (`"12345"`) or a server name as shown in the netcup UI (`"v2202508149564377314"`). Name lookup calls `ListServers` with a name filter and matches exactly. Returns an error if ambiguous or not found.

### Async tasks

Many mutating API calls return a `*TaskInfo` with a UUID. If the API responds with HTTP 200 and an empty body (synchronous completion), `PatchServer` returns `nil, nil`. `writeTask` treats a nil task as `OK` and prints that string. Callers that need the task UUID for `--wait` should check for nil before polling.

### Risky operations

`confirmRisky(cmd, opts, description)` in `internal/cli/confirm.go` prompts the user unless `opts.Yes` is true (set by `--yes` / `-y`). Call it at the top of any command that is destructive or hard to reverse.

### Body flags pattern

Commands that accept arbitrary JSON request bodies expose two flags:

```go
cmd.Flags().StringVar(&jsonBody, "body", "", "JSON request body")
cmd.Flags().StringVar(&jsonFile, "body-file", "", "file containing JSON request body")
```

Parsed with `parseBodyFlags(raw, path)` from `internal/cli/json_flags.go`.

## Adding a New CLI Command

1. Pick the right file under `internal/cli/` (read-only → `readonly_commands.go`, mutating → `broad_commands.go`, extras → `api_commands.go`).
2. Write the command function returning `*cobra.Command`.
3. If it takes a server argument, use `resolveServerID` — never parse the server ID directly with `strconv.Atoi`.
4. If it is destructive, call `confirmRisky` before doing any work.
5. Wire the command into its parent in `root.go` or the relevant `newXxxCommand()` function.

## Adding a New API Method

1. Add the method to the appropriate file in `pkg/netcup/`:
   - Read-only → `api.go`
   - Mutating → `api_writes.go`
   - Everything else → `api_missing.go`
2. Use `c.DoJSON` for standard calls and `c.DoMergePatch` for PATCH endpoints (check the OpenAPI spec in `assets/` for the expected Content-Type).
3. Add the corresponding model types to `models.go` if needed.
4. Write a test in `pkg/netcup/client_test.go` using `roundTripFunc` to stub the HTTP transport.

## Key Constraints

- **HTTPS only**: both `APIBaseURL` and `AuthBaseURL` are validated in `newApp`; non-https URLs are rejected.
- **No shell execution**: the codebase makes no `exec.Command` calls. Do not introduce any.
- **No third-party libraries beyond cobra**: keep the dependency footprint minimal.
- **Confirm before destructive actions**: any command that deletes data, changes power state, or formats a disk must call `confirmRisky`.
- **`userID` is required for user-scoped endpoints**: stored as `a.cfg.UserID` after login. If it is 0, the user is not logged in or login stored a bad value.

## Common Pitfalls

- PATCH requests to the SCP API require `Content-Type: application/merge-patch+json`, not `application/json`. Always use `DoMergePatch` for PATCH calls.
- `PatchServer` returns `nil, nil` on HTTP 200 (synchronous completion). Callers must handle a nil `*TaskInfo`.
- `strconv.Atoi` overflows on 32-bit systems for large server IDs. The codebase uses `int`; this is only safe on 64-bit targets.
- `go install @latest` resolves to the latest **tagged** release. Without a tag, the proxy may serve a stale pseudo-version. Use `@main` or a specific tag when testing unreleased changes on a remote machine.
