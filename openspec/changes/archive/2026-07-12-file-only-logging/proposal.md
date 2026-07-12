## Why

Log messages currently write to stderr unconditionally, interleaving with CLI output. This is particularly disruptive when using `--json` for machine-readable CLI output, as log text pollutes the JSON stream on stderr.

## What Changes

- **BREAKING**: Default log output moves from stderr to `<dataDir>/wallet.log` — no more log messages on stderr
- `--log-file` flag now means "write here instead of default" rather than "also write here"
- `newLogger` signature changes from `newLogger(cmd)` to `newLogger(cmd, dir)` to accept the data directory for default path resolution
- Remove `MultiHandler` usage; write only to a single `JSONHandler`-backed file
- Fall back to `io.Discard` silently if the log file cannot be opened

## Capabilities

### New Capabilities
<!-- None — this change modifies existing logging behavior, not adding new capabilities. -->

### Modified Capabilities
- `structured-logging`: Log file output flag and stderr output behavior changes. Default output moves from stderr to file. `--log-file` overrides the default path instead of adding a secondary destination.

## Impact

- `internal/cli/helpers.go` — `newLogger` signature and body, `getService` call site
- `internal/cli/init.go` — `runInit` call site
- `internal/cli/logging.go` — `MultiHandler` type remains but is no longer used by the main path
- User-visible: log output no longer appears on stderr; users must check `<dataDir>/wallet.log` or their custom `--log-file` path for logs
