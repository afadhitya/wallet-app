## Why

The repository needs a consistent Go project foundation before application development can proceed reliably. Adding standard Go tooling, linting, unit test coverage checks, and GitHub Actions CI will make quality gates repeatable for local development and pull requests.

## What Changes

- Initialize the repository as a Go module with a conventional project layout suitable for the wallet application.
- Add local quality commands for formatting, linting, tests, and coverage checks.
- Add a GitHub Actions workflow that runs Go tests and enforces unit test coverage.
- Add a code linter configuration so CI and local development use the same rules.
- Document how contributors run the Go quality checks locally.

## Capabilities

### New Capabilities
- `go-project-quality`: Defines the repository's Go project structure and required quality automation for linting, testing, and coverage enforcement.

### Modified Capabilities

## Impact

- Affects repository root project configuration, Go module files, CI workflow files, lint configuration, and developer documentation.
- Introduces Go toolchain expectations for local development and GitHub Actions.
- Adds CI checks that can fail pull requests when linting, tests, or coverage thresholds are not met.
