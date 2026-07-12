## Context

The current logger (`internal/cli/helpers.go`) uses a `MultiHandler` that writes to both `os.Stderr` (text) and optionally a log file (JSON). This means log output always appears on stderr, interfering with CLI output — especially when `--json` is used and log messages interleave with JSON responses on stderr.

The design doc at `docs/superpowers/specs/2026-07-12-file-only-logging-design.md` provides detailed motivation and design decisions.

## Goals / Non-Goals

**Goals:**
- Redirect all `slog` output exclusively to a log file (default or custom path)
- Remove stderr `TextHandler` so CLI output is never polluted by log messages
- `--log-file` flag overrides the default path instead of being an additional destination
- Silent fallback to `io.Discard` if the file cannot be opened

**Non-Goals:**
- No log rotation, compression, or retention
- No config file changes (no `[log]` section)
- No changes to log level or verbosity behavior
- No changes to CLI `--json` output

## Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Default log path | `<dataDir>/wallet.log` | Same directory as the DB (`~/.local/share/wallet/`). Directory already created by `getService()` before logger creation. |
| Log format | JSON only (`JSONHandler`) | Machine-readable. Matches current file output format. Consistent regardless of destination. |
| `--log-file` semantics | Override instead of addition | The flag now means "write here instead of default." No dual-output use case remains. |
| File open failure | `io.Discard` fallback | No output pollution. Don't break the CLI with a missing file. |
| `MultiHandler` | Removed from usage; type kept | No longer needed since only one handler writes. Type kept in `logging.go` for potential future use. |
| `-v` / `--verbose` | Unchanged | Still controls `slog.Level` identically via the same `countFlag` logic. |

### Before/After Architecture

```
Before:                         After:
newLogger(cmd)                  newLogger(cmd, dir)
  ├─ TextHandler(os.Stderr)       └─ JSONHandler(file)
  └─ JSONHandler(file, optional)       dir/wallet.log  (default)
    └─ MultiHandler                       or --log-file (override)
```

## Risks / Trade-offs

- **Users lose real-time stderr log visibility**: Previously, errors/warnings appeared immediately on stderr. Now users must tail the log file. Mitigation: document the default path; users can `tail -f ~/.local/share/wallet/wallet.log`.
- **Log file grows unbounded**: No rotation. Mitigation: non-goal for this change; addressed in future work.
