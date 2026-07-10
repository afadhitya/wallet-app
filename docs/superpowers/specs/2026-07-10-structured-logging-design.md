# Structured Logging Design

## Overview

Add structured logging using `log/slog` (Go stdlib) across all three layers of the wallet app: CLI, service, and database. No new dependencies.

## Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Log library | `log/slog` (stdlib) | Already available in Go 1.25.1. No new dependency. |
| Default level | `WARN` | Quiet by default. Users opt into more verbosity. |
| Verbosity control | `-v` = INFO, `-vv` = DEBUG (count flag) | Natural CLI pattern. Cobra supports `CountVarP`. |
| Logger propagation | Inject `*slog.Logger` into `Service` struct | Consistent with how `gen.Querier` is injected. Testable. |
| Context plumbing | Skip | Keeps scope focused. Use non-context slog variants. |
| Silent mode | None | Default is WARN (quiet enough). Users can redirect stderr. |
| `gen/` layer | No changes | sqlc-generated code is never touched. |

## Flags

Two new global persistent flags on the root Cobra command:

| Flag | Type | Description |
|------|------|-------------|
| `-v`, `--verbose` | Count (`CountVarP`) | `-v` = INFO, `-vv` = DEBUG. Default (no flag) = WARN. |
| `--log-file` | String | Path to write JSON-formatted logs. Empty (default) = no file output. |

## Logger Construction

A `newLogger(cmd *cobra.Command) *slog.Logger` function in `internal/cli/helpers.go`:

```
log level = WARN
             +1 per -v

no --log-file:
  slog.NewTextHandler(os.Stderr, level)  →  human-readable to stderr only

--log-file set:
  slog.NewTextHandler(os.Stderr, level)   +   slog.NewJSONHandler(file, level)
                                   ↓
                          custom MultiHandler  →  dual output
```

The `MultiHandler` is a ~20-line helper in a new file (`internal/cli/logging.go`) that writes to two handlers simultaneously. slog has no built-in multi-handler.

## Architecture

### Logger Injection Path

```
root.go (flags defined)
    │
    ▼
helpers.go: withService()  ──calls──►  newLogger(cmd)
    │                                        │
    │  passes logger to:                     ▼
    │                                       
    ├──► db.Open(path, logger)             TextHandler(stderr)
    ├──► db.Migrate(database, logger)      [+ JSONHandler(file)]
    └──► service.New(database, logger)
             │
             ▼
        Service.logger
```

### File Changes

| File | Change |
|------|--------|
| `internal/cli/root.go` | Add `-v`/`--verbose` and `--log-file` persistent flags |
| `internal/cli/helpers.go` | Add `newLogger()`, `safeGetService()`; update `getService()` + `withService()` signatures |
| `internal/cli/logging.go` | NEW — `MultiHandler` implementation |
| `internal/service/service.go` | Add `logger *slog.Logger` to `Service` struct; update `New()` and `NewWithQuerier()` |
| `internal/service/*.go` | Add log calls at entry/exit/error paths in all public methods |
| `internal/db/db.go` | Add `*slog.Logger` param to `Open()` and `Migrate()` |
| `internal/testdb/testdb.go` | Update `Open()` to accept a logger param (or supply discard logger) |
| All `*_test.go` files | Update calls to `service.New()`/`NewWithQuerier()` and `db.Open()` to pass a logger |

## Log Call Sites

### DB Layer (`internal/db/db.go`)

| Level | Event | Attributes |
|-------|-------|------------|
| `DEBUG` | Opening database | `path` |
| `INFO` | Database opened | `journal_mode` |
| `INFO` | Migration applied | `file`, `version` |
| `DEBUG` | Migration skipped (already applied) | `file`, `version` |
| `INFO` | Migrations complete | — |

### CLI Layer (`internal/cli/helpers.go`)

| Level | Event | Attributes |
|-------|-------|------------|
| `INFO` | Command started | `command`, `arg_count` |
| `INFO` | Command completed | `command`, `latency` |
| `WARN` | Command failed | `command`, `latency`, `error` |
| `ERROR` | Service initialization failed | `error` |

### Service Layer (`internal/service/*.go`)

Each public method follows this pattern:

| Level | Event | Attributes |
|-------|-------|------------|
| `INFO` | Entry | method name, key input params |
| `DEBUG` | Intermediate step | resolved references, computed values |
| `WARN` | Business rule violation | method name, reason, relevant params |
| `ERROR` | Unexpected DB failure | method name, error |
| `INFO` | Exit (success) | method name, result identifiers |

## Patterns

### Service Method Template

```go
func (s *Service) AddExpense(params CreateExpenseParams) (*TransactionResult, error) {
    s.logger.Info("AddExpense called",
        slog.Int64("amount", params.Amount),
        slog.String("account", params.AccountName),
        slog.String("category", params.CategoryName),
    )

    // ... validation, DB calls ...

    if err != nil {
        s.logger.Warn("AddExpense failed",
            slog.String("reason", err.Error()),
        )
        return nil, err
    }

    s.logger.Info("AddExpense completed",
        slog.Int64("txID", result.TransactionID),
        slog.Int64("balance", result.NewBalance),
    )
    return result, nil
}
```

### Tests

All tests pass a silent discard logger:

```go
logger := slog.New(slog.NewTextHandler(io.Discard, nil))
svc := service.NewWithQuerier(db, querier, logger)
```

### Rules

- Use `slog.Int64`, `slog.String`, `slog.Bool`, `slog.Duration`, `slog.Group` — never `fmt.Sprintf` in log messages.
- Never log: file paths containing usernames, config contents beyond structure, passwords, tokens, API keys.
- Log messages are lowercase, no trailing punctuation.
- Log output goes to stderr only (the human-readable stream). JSON only goes to `--log-file`.
- The `--json` flag for CLI output (stdout) is independent of logging and unchanged.

## Non-Goals

- No automatic per-query logging or slow query detection (requires `gen/` layer changes)
- No context.Context plumbing through service methods
- No log rotation, compression, or retention
- No request-scoped logger with command name baked in
- No `--silent` flag
