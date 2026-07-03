## MODIFIED Requirements

### Requirement: Local quality commands
The repository SHALL provide documented local commands for building, testing, generating coverage, linting, tidying dependencies, generating sqlc code, and cleaning generated artifacts.

#### Scenario: Developer runs quality checks locally
- **WHEN** a developer follows the documented local quality command sequence
- **THEN** formatting, linting, unit tests, coverage checks, sqlc generation, and dependency verification run without requiring GitHub Actions

## ADDED Requirements

### Requirement: Makefile build tooling
The repository SHALL provide Makefile targets for common wallet development workflows.

#### Scenario: Build wallet binary locally
- **WHEN** a developer runs `make build`
- **THEN** the wallet binary is built from `./cmd/wallet` into `bin/wallet`

#### Scenario: Run wallet tests locally
- **WHEN** a developer runs `make test`
- **THEN** Go tests run for all packages with `go test ./...`

#### Scenario: Generate coverage locally
- **WHEN** a developer runs `make test-cover`
- **THEN** Go tests produce `coverage.out` and an HTML coverage report at `coverage.html`

#### Scenario: Verify dependencies locally
- **WHEN** a developer runs the dependency target
- **THEN** Go module dependencies are tidied and verified

### Requirement: Repository ignore rules
The repository SHALL ignore generated local build, coverage, and SQLite runtime artifacts that should not be committed.

#### Scenario: Generated artifacts are ignored
- **WHEN** local development produces binaries, coverage files, or SQLite database files
- **THEN** `.gitignore` excludes `/bin/`, `/coverage.out`, `/coverage.html`, `*.db`, `*.db-journal`, and `*.db-wal`

### Requirement: Coverage artifacts
The GitHub Actions workflow SHALL publish coverage artifacts for debugging coverage results.

#### Scenario: Coverage workflow uploads profile
- **WHEN** the quality workflow runs tests with coverage
- **THEN** it uploads `coverage.out` as an artifact regardless of success or failure

#### Scenario: Coverage workflow uploads HTML report on failure
- **WHEN** the quality workflow fails coverage validation
- **THEN** it generates `coverage.html` and uploads the HTML report as an artifact
