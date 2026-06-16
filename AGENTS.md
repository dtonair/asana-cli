# AGENTS.md — asana-cli

Cached repo memory for agents working in this project. Keep this current.

## What this is

`asana-cli` is a standalone Go CLI that exposes Asana operations for consumption
by other agents and scripts. It was ported 1:1 from the `pi-extensions/asana`
TypeScript Pi extension (7 operations: 6 read + 1 write). Output is JSON by
default; `--human` gives text summaries.

## Layout

```
cmd/asana-cli/main.go          # entrypoint: cli.Execute() -> os.Exit
internal/config/               # env-only credential loading + workspace resolution
  config.go                    #   Load(), LoadFrom(getenv), Config.ResolveWorkspace
internal/asana/                # HTTP client (ported from src/asana-client.ts)
  client.go                    #   Client.Request, Client.Paginate, HTTPError, EncodePathSegment
internal/cli/                  # Cobra command tree (one file per subcommand)
  root.go                      #   root cmd, persistent flags, exitCodeFor, usageError type
  run.go                       #   buildClient, withTimeout, validateLimit, requireFlag, query helpers, requestData
  output.go                    #   {ok,data} / {ok,error} envelopes, summarizers, humanList
  me.go list_workspaces.go list_projects.go search_tasks.go
  get_task.go list_task_stories.go comment_on_task.go
```

## Conventions

- **Error typing:** return a `*usageError` (via `usageErrorf` or wrapping) for
  usage/config problems → exit code 2. Any other error → exit code 1. `main`
  renders the error envelope and calls `exitCodeFor`.
- **Output:** commands call `writeSuccess(cmd.OutOrStdout(), data, opts.human, humanText)`.
  Single-resource commands pass the unwrapped object; list commands pass
  `[]json.RawMessage` and build `humanText` via `humanList(...)`.
- **Data unwrapping:** `requestData` strips Asana's top-level `{"data": ...}`;
  `Paginate` returns the accumulated `data` array elements.
- **Pagination:** page size 50, max 10 pages, capped at `--limit` (1..100).
  Constants `pageSize` / `maxPages` in `run.go`.
- **Tri-state flags:** detect explicit set with `cmd.Flags().Changed(name)`
  (see `--completed` in `search_tasks.go`).
- **Persistent flags** live on the root and populate the package-level `opts`
  (`--human`, `--verbose`, `--timeout`).

## Testing

- `go test ./...` — no network/token needed.
- Command tests use `runWithServer` (in `commands_test.go`): spins up `httptest`,
  sets `ASANA_ACCESS_TOKEN=tok` and `ASANA_API_BASE=<server>`, runs the root
  command, returns stdout + error. Assert exit semantics with `exitCodeFor(err)`.
- `ASANA_API_BASE` is the test-only base-URL seam (read in `buildClient`); it is
  intentionally undocumented for users.
- Tests assert the token never leaks into errors.

## How an agent invokes it

Prefer default JSON; parse the `{ok, data|error}` envelope. Branch on the process
exit code (0/1/2) and on `error.status` for HTTP failures. Pass `--workspace-gid`
explicitly or rely on `ASANA_DEFAULT_WORKSPACE`.

## Provenance / parity

Source of truth for behavior is `~/code/pi-extensions/asana/src/*`. If you change
endpoints, query params, error messages, or pagination, keep them aligned with
that extension (and its tests) unless intentionally diverging.
