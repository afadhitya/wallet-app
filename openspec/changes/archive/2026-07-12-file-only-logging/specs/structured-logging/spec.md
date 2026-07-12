## MODIFIED Requirements

### Requirement: Log file output flag
The system SHALL write all log output exclusively to a file in JSON format. When `--log-file` is not specified, logs SHALL be written to `<dataDir>/wallet.log`. When `--log-file` is specified, logs SHALL be written to the given path. Logs SHALL NOT be written to stderr in either case. If the log file cannot be opened, logging SHALL silently fall back to `io.Discard`.

#### Scenario: No log file specified
- **WHEN** the CLI is invoked without `--log-file`
- **THEN** logs are written to `<dataDir>/wallet.log` in JSON format and nothing is written to stderr

#### Scenario: Log file specified
- **WHEN** the CLI is invoked with `--log-file /path/to/log.json`
- **THEN** logs are written to the specified file in JSON format and nothing is written to stderr

#### Scenario: Log file cannot be opened
- **WHEN** the CLI is invoked and the log file cannot be opened (e.g., permissions, read-only filesystem)
- **THEN** logging silently falls back to `io.Discard` and no error is surfaced to the user

### Requirement: Log format and conventions
All log messages SHALL use lowercase with no trailing punctuation. Structured attributes SHALL use `slog.Int64`, `slog.String`, `slog.Bool`, `slog.Duration`, `slog.Group` — never `fmt.Sprintf` in log messages. Secrets, tokens, and passwords SHALL NOT be logged.

#### Scenario: Structured attribute usage
- **WHEN** any log message is emitted
- **THEN** all dynamic values are passed as typed slog attributes, not interpolated into the message string

#### Scenario: Non-interference with JSON output
- **WHEN** the `--json` flag is used for CLI output
- **THEN** stdout JSON output format is unchanged; logging goes exclusively to the log file and never appears on stdout or stderr
