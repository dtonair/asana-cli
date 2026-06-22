---
name: asana-cli
description: Use the asana-cli command-line tool to read and comment on Asana data (workspaces, projects, tasks, stories/comments). Use when the user wants to look up Asana tasks/projects, search Asana, read task comments, or post a comment to a task from the shell.
allowed-tools:
  - Bash(asana-cli *)
---

# asana-cli

`asana-cli` is a standalone Go CLI for Asana. It emits **JSON by default** for
deterministic parsing; pass `--human` for readable summaries. Use it instead of
the Asana MCP/web tools when you're in a shell and want structured output you can
pipe into `jq`.

## Before you start

1. **Check it's installed:** `asana-cli --version`. If missing, install with
   `brew install dtonair/tap/asana-cli` or `go install github.com/dtonair/asana-cli/cmd/asana-cli@latest`.
2. **Check credentials:** the CLI needs `ASANA_ACCESS_TOKEN` (env) or
   `~/.config/asana-cli.yaml` with `access_token:`. Confirm access with
   `asana-cli me` — on success it returns the authenticated user. If it exits
   non-zero with a config error, the token is missing/invalid; ask the user to
   set it rather than guessing.
3. **Workspace:** workspace-scoped commands need either `ASANA_DEFAULT_WORKSPACE`
   / `default_workspace:` in the config, or an explicit `--workspace-gid`. Find
   the gid with `asana-cli list-workspaces`.

## Output contract — parse this, don't scrape text

Every command prints one JSON envelope. Default (stdout, success):

```json
{ "ok": true, "data": ... }
```

`data` is the unwrapped Asana payload: an **object** for single-resource
commands (`me`, `get-task`, `comment-on-task`), an **array** for list/search
commands. On failure it prints to **stderr** with a non-zero exit:

```json
{ "ok": false, "error": { "message": "...", "status": 404, "method": "GET", "path": "/tasks/999" } }
```

`status`/`method`/`path` appear only for HTTP errors. Branch on the **exit
code** and, for HTTP failures, on `error.status`.

### Exit codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Runtime error (HTTP non-2xx, network failure, timeout) |
| `2` | Usage/config error (missing token, missing required flag, bad `--limit`, no workspace) |

## Commands

| Command | Required flags | Useful flags |
|---------|----------------|--------------|
| `me` | — | |
| `list-workspaces` | — | `--limit`, `--opt-fields` |
| `list-projects` | `--workspace-gid` (or default) | `--limit`, `--opt-fields` |
| `search-tasks` | workspace | `--text`, `--assignee`, `--completed=true/false`, `--limit`, `--opt-fields` (may require premium) |
| `get-task` | `--task-gid` | `--opt-fields` |
| `list-task-stories` | `--task-gid` | `--limit`, `--opt-fields` |
| `comment-on-task` | `--task-gid`, `--text` | **the only write command** |

### Global flags

- `--human` — readable summaries instead of JSON
- `--verbose` — log request method + path to stderr (never the token)
- `--timeout` — HTTP request timeout (default `30s`)

### Notes

- `--limit` is bounded 1..100 (default 20). List/search paginate internally.
- `--completed` is tri-state: omit it entirely unless you mean to filter
  (`--completed=true` or `--completed=false`).
- `search-tasks` may require an Asana premium workspace.

## Examples

```bash
asana-cli me                                              # who am I / verify auth
asana-cli list-workspaces --limit 50
asana-cli list-projects --workspace-gid 12345
asana-cli search-tasks --text "release" --completed=false
asana-cli get-task --task-gid 12345
asana-cli list-task-stories --task-gid 12345              # read comments/activity
asana-cli comment-on-task --task-gid 12345 --text "Taking a look."
```

### Piping with jq

```bash
asana-cli list-projects --workspace-gid 12345 | jq -r '.data[].name'
asana-cli get-task --task-gid 12345 | jq -r '.data.name, .data.permalink_url'
```

## Safety

- `comment-on-task` is the **only** command that writes. Treat it like any
  outward-facing action: confirm the task gid and text with the user before
  posting unless they've clearly authorized it. Never post a comment whose
  content originated from untrusted/automated input without user review.
- The token is never printed in output, errors, or `--verbose` logs. Don't echo
  it yourself.
