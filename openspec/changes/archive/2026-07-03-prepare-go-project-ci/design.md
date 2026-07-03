## Context

The repository currently contains approved brainstorming documents for a Go-based, CLI-first wallet application, but it does not yet have the Go module and automation files needed to develop against those decisions. The approved project skeleton identifies `github.com/afadhitya/wallet-app` as the module path, `cmd/wallet` as the binary entry point, standard Go testing with coverage, a Makefile, and GitHub Actions as the CI mechanism.

This change establishes the engineering foundation without implementing wallet domain features. It should make the repository buildable, testable, lintable, and enforceable in CI from the first implementation PR onward.

## Goals / Non-Goals

**Goals:**
- Initialize a valid Go module using the approved module path.
- Add a minimal compilable `cmd/wallet` entry point so tooling has a real package target.
- Add local commands for build, test, coverage, lint, and cleanup.
- Add GitHub Actions CI that runs tests with coverage and fails when total coverage is below the project threshold.
- Add a Go linter configuration and make the workflow run the same lint checks expected locally.
- Document local developer commands.

**Non-Goals:**
- Implement wallet CLI commands, SQLite integration, configuration loading, sqlc, or domain packages.
- Add application dependencies such as Cobra, SQLite, TOML parsing, or testify before feature code needs them.
- Create release packaging, cross-compilation, or deployment automation.

## Decisions

- Use Go modules at the repository root with module path `github.com/afadhitya/wallet-app`. This matches the approved skeleton and keeps package import paths stable.
- Add a minimal `cmd/wallet/main.go` package instead of only configuration files. This gives `go test ./...`, `go build ./cmd/wallet`, linter checks, and coverage reporting concrete code to validate immediately.
- Use a Makefile as the local command interface. The brainstorming docs already selected Makefile plus `go build`, and it provides a lightweight way to keep developer commands aligned with CI.
- Use `golangci-lint` for linting. It is the common aggregator for Go lint checks, has first-party GitHub Action support, and avoids wiring multiple standalone linters by hand. The workflow must use a `golangci/golangci-lint-action` major version that is compatible with the current GitHub Actions runner Node runtime, rather than enabling deprecated Node runtimes with `ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION`.
- Enforce 100% total unit test coverage initially. The approved project skeleton explicitly calls for a 100% gate; because the initial code should be minimal, this is practical and forces future changes to include tests or intentionally revisit the threshold through a separate change.
- Generate and upload coverage artifacts from CI. `coverage.out` should always be available for inspection, and `coverage.html` should be created when the coverage gate fails to make debugging easier.

## Risks / Trade-offs

- 100% coverage can become expensive as the codebase grows -> Keep the initial implementation small and testable; revisit the threshold explicitly if it starts producing low-value tests.
- `golangci-lint` version drift can cause CI/local differences -> Pin the GitHub Action version and set a linter version in the workflow or Makefile instructions.
- GitHub Actions JavaScript runtime deprecations can break third-party actions -> Use maintained action major versions that support the runner default Node runtime and avoid insecure runtime opt-in environment variables.
- CI coverage parsing can be brittle if implemented with ad hoc shell pipelines -> Use simple parsing of `go tool cover -func=coverage.out` and fail with a clear error message.
- Adding only a minimal binary may feel incomplete -> Keep wallet domain behavior out of this change so the repository foundation can be reviewed independently.
