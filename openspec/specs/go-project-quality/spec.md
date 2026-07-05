## Purpose

TBD - Captures Go project quality requirements for the wallet application.

## Requirements

### Requirement: Go module initialization
The repository SHALL define a Go module at the repository root using module path `github.com/afadhitya/wallet-app` and include a minimal compilable wallet binary entry point.

#### Scenario: Module-aware Go commands run
- **WHEN** a developer runs `go test ./...` from the repository root
- **THEN** Go discovers the module and completes successfully without missing module metadata errors

#### Scenario: Wallet binary package builds
- **WHEN** a developer runs `go build ./cmd/wallet`
- **THEN** the wallet command package compiles successfully

### Requirement: Local quality commands
The repository SHALL provide documented local commands for building, testing, generating coverage, linting, tidying dependencies, generating sqlc code, and cleaning generated artifacts.

#### Scenario: Developer runs quality checks locally
- **WHEN** a developer follows the documented local quality command sequence
- **THEN** formatting, linting, unit tests, coverage checks, sqlc generation, and dependency verification run without requiring GitHub Actions

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

### Requirement: Code linting configuration
The repository SHALL configure a Go linter and provide a local command that runs the configured lint checks against all Go packages.

#### Scenario: Linter runs locally
- **WHEN** a developer runs the documented lint command
- **THEN** the configured Go linter checks all repository Go packages and reports failures with non-zero exit status

### Requirement: GitHub Actions quality workflow
The repository SHALL include a GitHub Actions workflow that runs on pushes to `main` and `brainstorming` and on pull requests to `main`.

#### Scenario: Workflow validates Go project quality
- **WHEN** the GitHub Actions workflow runs for a configured push or pull request event
- **THEN** it checks out the repository, sets up Go from the module configuration, runs the configured linter, and runs Go unit tests with coverage enabled

#### Scenario: Workflow uses maintained action runtimes
- **WHEN** the workflow invokes third-party GitHub Actions such as `golangci/golangci-lint-action`
- **THEN** each action uses a maintained version compatible with the default GitHub Actions runner Node runtime without setting `ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION`

### Requirement: Unit test coverage gate
The GitHub Actions workflow SHALL fail when total Go unit test coverage is below 100%.

#### Scenario: Coverage meets threshold
- **WHEN** total coverage reported by `go tool cover -func=coverage.out` is 100%
- **THEN** the workflow succeeds after publishing the coverage profile artifact

#### Scenario: Coverage below threshold
- **WHEN** total coverage reported by `go tool cover -func=coverage.out` is less than 100%
- **THEN** the workflow fails with a clear coverage error and publishes coverage artifacts for debugging

### Requirement: Coverage artifacts
The GitHub Actions workflow SHALL publish coverage artifacts for debugging coverage results.

#### Scenario: Coverage workflow uploads profile
- **WHEN** the quality workflow runs tests with coverage
- **THEN** it uploads `coverage.out` as an artifact regardless of success or failure

#### Scenario: Coverage workflow uploads HTML report on failure
- **WHEN** the quality workflow fails coverage validation
- **THEN** it generates `coverage.html` and uploads the HTML report as an artifact

### Requirement: Codecov upload restricted to main branch
The CI workflow SHALL upload coverage reports to Codecov only when running on the `main` branch, to prevent duplicate uploads from pull request branches.

#### Scenario: Coverage uploaded to Codecov on main push
- **WHEN** the quality workflow runs on a push to the `main` branch
- **THEN** the workflow uploads the coverage profile to Codecov using `codecov/codecov-action`

#### Scenario: Coverage not uploaded to Codecov on PR
- **WHEN** the quality workflow runs on a pull request to `main`
- **THEN** the workflow runs coverage checks locally but does NOT upload to Codecov

#### Scenario: Coverage not uploaded to Codecov on brainstorming push
- **WHEN** the quality workflow runs on a push to the `brainstorming` branch
- **THEN** the workflow runs coverage checks locally but does NOT upload to Codecov

### Requirement: Developer documentation
The repository SHALL document the Go project setup and quality commands needed by contributors.

#### Scenario: Contributor finds setup instructions
- **WHEN** a contributor reads the repository documentation
- **THEN** they can identify the required Go version and commands for build, test, coverage, lint, and dependency tidy operations
