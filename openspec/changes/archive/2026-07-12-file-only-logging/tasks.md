## 1. Update `newLogger` implementation

- [x] 1.1 Change `newLogger` signature from `newLogger(cmd *cobra.Command)` to `newLogger(cmd *cobra.Command, dir string)` in `internal/cli/helpers.go`
- [x] 1.2 Define default log path as `filepath.Join(dir, "wallet.log")` when `--log-file` is empty
- [x] 1.3 Replace `TextHandler(os.Stderr)` + `MultiHandler` pattern with single `JSONHandler` writing to the log file
- [x] 1.4 Fall back to `slog.NewTextHandler(io.Discard, ...)` if log file cannot be opened (instead of falling back to stderr)

## 2. Update callers of `newLogger`

- [x] 2.1 Update `getService()` in `internal/cli/helpers.go` to pass `dir` to `newLogger(cmd, dir)`
- [x] 2.2 Update `runInit()` in `internal/cli/init.go` to pass `dir` to `newLogger(cmd, dir)`

## 3. Update tests

- [x] 3.1 Update `TestNewLoggerDefaultWarn` — write to a temp file instead of capturing stderr; verify file content
- [x] 3.2 Update `TestNewLoggerVerboseInfo` — write to a temp file instead of capturing stderr; verify file content
- [x] 3.3 Update `TestNewLoggerVerboseDebug` — write to a temp file instead of capturing stderr; verify file content
- [x] 3.4 Update `TestNewLoggerVerboseBeyondMax` — write to a temp file instead of capturing stderr; verify file content
- [x] 3.5 Update `TestNewLoggerWithLogFile` — verify log file content only (no stderr check); verify stderr is empty
- [x] 3.6 Update `TestNewLoggerLogFileOpenError` — verify log output is discarded (no stderr output when file open fails)
- [x] 3.7 Update `TestNewLoggerVerboseFlagError` — write to a temp file instead of capturing stderr; verify file content

## 4. Verify

- [x] 4.1 Run `make coverage-check` to ensure all tests pass and coverage threshold is met
- [x] 4.2 Run `make lint` to ensure code quality
