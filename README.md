# asana-cli

A standalone Go CLI for Asana, ported from the `pi-extensions` Asana extension so
any agent (or human/script) can invoke Asana operations from the shell. Output is
**JSON by default** for deterministic machine parsing; pass `--human` for readable
summaries.

## Install

### Homebrew (recommended)

```bash
brew tap dtonair/tap
brew install asana-cli
```

To upgrade later: `brew upgrade asana-cli`.

### Build from source

```bash
go build -o asana-cli ./cmd/asana-cli
# or
go install ./cmd/asana-cli
```

Requires Go 1.22+. The only third-party dependency is `spf13/cobra`.

Check the installed version with `asana-cli --version`.

## Configuration

Credentials are read from environment variables only:

```bash
export ASANA_ACCESS_TOKEN="your-asana-personal-access-token"   # required
export ASANA_DEFAULT_WORKSPACE="workspace-gid"                 # optional
```

`ASANA_DEFAULT_WORKSPACE` is optional, but workspace-scoped commands require either
that variable or an explicit `--workspace-gid`.

Recommended token scopes: `users:read`, `workspaces:read`, `projects:read`,
`tasks:read`, `stories:read`, and `stories:write` for `comment-on-task`.

## Global flags

| Flag | Default | Description |
|------|---------|-------------|
| `--human` | off | Print human-readable summaries instead of JSON |
| `--verbose` | off | Log request method + path to stderr (never the token) |
| `--timeout` | `30s` | HTTP request timeout |

## Commands

| Command | Asana endpoint | Notes |
|---------|----------------|-------|
| `me` | `GET /users/me` | |
| `list-workspaces` | `GET /workspaces` | `--limit`, `--opt-fields` |
| `list-projects` | `GET /workspaces/{ws}/projects` | `--workspace-gid`, `--limit`, `--opt-fields` |
| `search-tasks` | `GET /workspaces/{ws}/tasks/search` | `--text`, `--assignee`, `--completed`, `--limit`, `--opt-fields` (may require premium) |
| `get-task` | `GET /tasks/{gid}` | `--task-gid` (required), `--opt-fields` |
| `list-task-stories` | `GET /tasks/{gid}/stories` | `--task-gid` (required), `--limit`, `--opt-fields` |
| `comment-on-task` | `POST /tasks/{gid}/stories` | `--task-gid`, `--text` (both required). The only write command. |

`--limit` is bounded to 1..100 (default 20). List/search commands paginate
internally (page size 50, up to 10 pages, capped at `--limit`).

`--completed` is tri-state: omitted entirely unless you pass it
(`--completed=true` or `--completed=false`).

## Output contract

**Success** (stdout):

```json
{
  "ok": true,
  "data": <Asana resource (object) or array of resources>
}
```

`data` is the unwrapped Asana payload: an object for single-resource commands
(`me`, `get-task`, `comment-on-task`), an array for list/search commands.

**Error** (stderr, non-zero exit):

```json
{
  "ok": false,
  "error": {
    "message": "Asana resource not found. ...",
    "status": 404,
    "method": "GET",
    "path": "/tasks/999"
  }
}
```

`status`/`method`/`path` are present only for HTTP errors. With `--human`, errors
print as a plain message line instead.

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Runtime error (HTTP non-2xx, network failure, timeout) |
| `2` | Usage/config error (missing token, missing required flag, bad `--limit`, no workspace) |

## Examples

```bash
asana-cli me
asana-cli list-workspaces --limit 50
asana-cli list-projects --workspace-gid 12345
asana-cli search-tasks --text "release" --completed=false
asana-cli get-task --task-gid 12345 --human
asana-cli list-task-stories --task-gid 12345
asana-cli comment-on-task --task-gid 12345 --text "Taking a look."
```

## Test

```bash
go test ./...
```

Tests run against an in-process `httptest` server; no network or real token is
required. The API base URL is overridable via the `ASANA_API_BASE` environment
variable (test-only seam; not a user-facing flag).

## Security

The token is read from the environment only, never written to disk, and never
included in any rendered output, error, or `--verbose` log line. All requests go
over HTTPS to `https://app.asana.com/api/1.0`.
