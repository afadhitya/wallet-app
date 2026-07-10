## Context

The wallet app is a CLI tool with a three-layer architecture: CLI (Cobra commands), service (business logic), and database (SQLite via sqlc). There is currently no logging — all output goes to stdout (JSON results) or is silently lost. Debugging requires code changes, and production issues leave no audit trail.

The design document at `docs/superpowers/specs/2026-07-10-structured-logging-design.md` provides a detailed technical plan. This design file distills the key architectural decisions.

## Goals / Non-Goals

**Goals:**
- Add structured logging using `log/slog` (Go stdlib) with no new dependencies
- Support configurable verbosity: WARN (default), INFO (`-v`), DEBUG (`-vv`)
- Support dual output: human-readable text to stderr + optional JSON file (`--log-file`)
- Inject `*slog.Logger` into `Service`, passing it through to `db.Open()`/`db.Migrate()`
- Add log calls at entry/exit/error paths in all service layer public methods
- Maintain 100% test coverage; all tests pass a discard logger

**Non-Goals:**
- No context.Context plumbing (use non-context slog variants)
- No changes to `internal/gen/` (sqlc-generated code)
- No per-query logging or slow query detection
- No log rotation, compression, or retention
- No `--silent` flag (default WARN is quiet enough)
- No changes to CLI `--json` output behavior

## Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Log library | `log/slog` (stdlib) | Already available in Go 1.25.1. No new dependency. |
| Default level | `WARN` | Quiet by default. Users opt into more verbosity. |
| Verbosity control | `-v` = INFO, `-vv` = DEBUG (count flag) | Natural CLI pattern. Cobra supports `CountVarP`. |
| Logger propagation | Inject `*slog.Logger` into `Service` struct | Consistent with how `gen.Querier` is injected. Testable. |
| Context plumbing | Skip | Keeps scope focused. Use non-context slog variants. |
| Multi-output handler | Custom `MultiHandler` in `internal/cli/logging.go` | slog has no built-in multi-handler. ~20 line implementation. |
| JSON output destination | Only `--log-file` gets JSON | stderr stays human-readable for terminal. JSON is machine-parseable for file. |

### Logger Construction Flow

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

### MultiHandler Design

The `MultiHandler` wraps two `slog.Handler` instances and delegates `Handle()`, `WithAttrs()`, and `WithGroup()` to both. When `--log-file` is not set, only the `TextHandler` is used.

### File Changes Summary

| File | Change |
|------|--------|
| `internal/cli/root.go` | Add `-v`/`--verbose` and `--log-file` persistent flags |
| `internal/cli/helpers.go` | Add `newLogger()`, `safeGetService()`; update `getService()` + `withService()` signatures |
| `internal/cli/logging.go` | NEW — `MultiHandler` implementation |
| `internal/service/service.go` | Add `logger *slog.Logger` to `Service` struct; update `New()` and `NewWithQuerier()` |
| `internal/service/*.go` | Add log calls at entry/exit/error paths in all public methods |
| `internal/db/db.go` | Add `*slog.Logger` param to `Open()` and `Migrate()` |
| `internal/testdb/testdb.go` | Update `Open()` to accept a logger param |
| All `*_test.go` files | Update calls to pass `slog.New(slog.NewTextHandler(io.Discard, nil))` |

## Risks / Trade-offs

- **Breaking constructor signatures**: `service.New()`, `service.NewWithQuerier()`, `db.Open()`, `db.Migrate()` all gain a `*slog.Logger` parameter — all callers (tests, CLI wiring) must be updated simultaneously.
- **Test noise**: Service tests using real DB may produce log output on failure. Mitigation: tests use `io.Discard` logger.
- **File locking on `--log-file`**: The JSON file handle is opened once and held. If the file is on a network mount or removed externally, writes will fail silently. No mitigation planned — edge case.
- **No request-scoped logger**: The same logger instance is used for all operations. No command name or correlation ID is baked in. Trade-off: simpler injection vs. richer context.
