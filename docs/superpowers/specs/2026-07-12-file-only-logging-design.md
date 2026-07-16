# File-Only Logging Design

## Overview

Redirect all `slog` log output exclusively to a file. Remove the stderr `TextHandler` so CLI output (stdout/stderr) is never polluted by log messages.

## Motivation

The current logger writes text output to `os.Stderr` unconditionally. This interferes with CLI output — especially when consuming JSON output via `--json` — because log messages and CLI responses interleave on stderr.

## Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Default log path | `$data_dir/wallet.log` | Same directory as the DB (`~/.local/share/wallet/`). Already created by `getService()`. |
| `--log-file` flag | Overrides default path | Preserves existing flag. Now means "write here instead of default" rather than "also write here". |
| Log format | JSON (`JSONHandler`) | Machine-readable. Matches current file output format. |
| All levels to file | Yes | WARN, ERROR, INFO, DEBUG — all go to file only. No stderr. |
| File open failure | Silently discard (`io.Discard`) | No output pollution. Don't break the CLI. |
| MultiHandler | No longer used | Single handler to file. Keep the type in `logging.go` for potential future use. |
| `-v` / `--verbose` | Unchanged | Still controls log level identically. |
| `--json` CLI flag | Unchanged | Independent of logging. Controls CLI success/error output format. |

## Logger Construction

```
newLogger(cmd, dataDir) → *slog.Logger

  logFile = cmd.Flag("--log-file")  OR  filepath.Join(dataDir, "wallet.log")
  level   = WARN + per -v

  open logFile with CREATE|WRONLY|APPEND
    success → slog.NewJSONHandler(file, level)
    failure → slog.NewTextHandler(io.Discard, level)   (silent fallback)
```

## Architecture

### Before

```
newLogger(cmd)
  TextHandler(os.Stderr)  ──always──┐
  JSONHandler(file)       ──if --log├──── MultiHandler → *slog.Logger
```

### After

```
newLogger(cmd, dataDir)
  JSONHandler(file)       ──default or --log-file──→ *slog.Logger
  (no stderr handler)
  (no MultiHandler)
```

### Call path

```
getService(cmd)
  ├── dir = filepath.Dir(dbPath)          // already computed
  ├── svcMkdirAll(dir, 0755)              // already done before logger
  ├── logger = newLogger(cmd, dir)        // now passes dir
  ├── db.Open(dbPath, logger)
  └── service.New(db, logger)

runInit(cmd)
  └── newLogger(cmd, dir)                 // passes DB dir
```

## File Changes

| File | Change |
|------|--------|
| `internal/cli/helpers.go` | `newLogger(cmd, dir)` — signature change; replace `TextHandler`+`MultiHandler` with single `JSONHandler`; default path logic |
| `internal/cli/init.go` | Update `newLogger()` call to pass `dir` |
| `internal/cli/logging.go` | `MultiHandler` unused by main path (keep, unchanged) |

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| No `--log-file` | Writes to `<dataDir>/wallet.log` |
| `--log-file /custom/path.log` | Writes to `/custom/path.log` |
| Dir doesn't exist | Already handled by `svcMkdirAll(dir, 0755)` in `getService()` before logger creation |
| File can't be opened (permissions, read-only FS) | Fall back to `io.Discard` — no log output anywhere |
| `:memory:` DB (tests) | Tests use `io.Discard` logger directly; `newLogger` not called |
| Existing log file | Append (`os.O_APPEND`) |

## Non-Goals

- No log rotation, compression, or retention
- No config file changes (no `[log]` section)
- No changes to log level or verbosity behavior
- No changes to CLI `--json` output
