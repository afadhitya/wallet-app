## Purpose

Provide structured logging capabilities to the wallet CLI — controlling log verbosity, output destinations, and log format conventions, with logger injection throughout the application layers.

## Requirements

### Requirement: Verbose logging flag
The system SHALL support a `-v`/`--verbose` count flag on the root Cobra command. Zero occurrences (default) SHALL set the log level to WARN. One occurrence (`-v`) SHALL set it to INFO. Two occurrences (`-vv`) SHALL set it to DEBUG.

#### Scenario: Default log level is WARN
- **WHEN** the CLI is invoked without `-v` or `--verbose`
- **THEN** only WARN and ERROR log messages are emitted

#### Scenario: Single verbose flag sets INFO
- **WHEN** the CLI is invoked with `-v` or `--verbose`
- **THEN** INFO, WARN, and ERROR log messages are emitted

#### Scenario: Double verbose flag sets DEBUG
- **WHEN** the CLI is invoked with `-vv` or `-v -v`
- **THEN** DEBUG, INFO, WARN, and ERROR log messages are emitted

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

### Requirement: Logger injection into Service
The `Service` struct SHALL hold a `*slog.Logger` field. The `New()` and `NewWithQuerier()` constructors SHALL accept a `*slog.Logger` parameter and store it.

#### Scenario: Service receives logger
- **WHEN** `service.New()` or `service.NewWithQuerier()` is called with a valid `*slog.Logger`
- **THEN** the returned Service contains the logger and can use it for logging

### Requirement: Logger injection into database layer
The `db.Open()` and `db.Migrate()` functions SHALL accept a `*slog.Logger` parameter and SHALL emit structured log events for database lifecycle operations.

#### Scenario: Database opened with logger
- **WHEN** `db.Open(path, logger)` is called
- **THEN** a DEBUG log is emitted for opening, and an INFO log with `journal_mode` is emitted on success

#### Scenario: Migration applied with logger
- **WHEN** `db.Migrate(database, logger)` applies a new migration
- **THEN** an INFO log with the migration `file` and `version` is emitted

### Requirement: Service method entry and exit logging
All public methods in the service layer SHALL emit structured logs at entry (INFO level with key input parameters) and exit (INFO level with result identifiers on success, WARN level on business error, ERROR level on unexpected DB failure).

#### Scenario: Successful method invocation
- **WHEN** a public service method completes successfully
- **THEN** an INFO log at entry with key input params and an INFO log at exit with result identifiers are emitted

#### Scenario: Business rule violation
- **WHEN** a public service method encounters a validation or business rule error
- **THEN** a WARN log with the method name and reason is emitted

#### Scenario: Unexpected database failure
- **WHEN** a public service method encounters an unexpected database error
- **THEN** an ERROR log with the method name and error details is emitted

### Requirement: Log format and conventions
All log messages SHALL use lowercase with no trailing punctuation. Structured attributes SHALL use `slog.Int64`, `slog.String`, `slog.Bool`, `slog.Duration`, `slog.Group` — never `fmt.Sprintf` in log messages. Secrets, tokens, and passwords SHALL NOT be logged.

#### Scenario: Structured attribute usage
- **WHEN** any log message is emitted
- **THEN** all dynamic values are passed as typed slog attributes, not interpolated into the message string

#### Scenario: Non-interference with JSON output
- **WHEN** the `--json` flag is used for CLI output
- **THEN** stdout JSON output format is unchanged; logging goes exclusively to the log file and never appears on stdout or stderr
