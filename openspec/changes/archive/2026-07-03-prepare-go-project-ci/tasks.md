## 1. Go Module Foundation

- [x] 1.1 Create root `go.mod` using module path `github.com/afadhitya/wallet-app` and an appropriate Go version.
- [x] 1.2 Add a minimal `cmd/wallet/main.go` entry point that compiles without implementing wallet domain behavior.
- [x] 1.3 Add unit test coverage for the initial Go code so `go test ./...` succeeds and reports 100% coverage.

## 2. Local Developer Tooling

- [x] 2.1 Add a `Makefile` with targets for `build`, `test`, `test-cover`, `coverage-check`, `lint`, `fmt`, `tidy`, and `clean`.
- [x] 2.2 Add repository ignore rules for Go build outputs and coverage artifacts such as `bin/`, `coverage.out`, and `coverage.html`.
- [x] 2.3 Add or update repository documentation with required Go version and local quality commands.

## 3. Linting

- [x] 3.1 Add a `golangci-lint` configuration file for the repository.
- [x] 3.2 Ensure the local lint command runs the configured linter across all Go packages and fails on lint violations.

## 4. GitHub Actions CI

- [x] 4.1 Add a GitHub Actions workflow that runs on pushes to `main` and `brainstorming` and pull requests to `main`.
- [x] 4.2 Configure the workflow to check out code, set up Go from `go.mod`, run linting, and run `go test -coverprofile=coverage.out -covermode=atomic ./...`.
- [x] 4.3 Add a CI coverage gate that fails when total coverage from `go tool cover -func=coverage.out` is below 100%.
- [x] 4.4 Configure CI artifact upload for `coverage.out` and generate/upload `coverage.html` when the coverage gate fails.

## 5. Verification

- [x] 5.1 Run `go test ./...` and confirm it passes.
- [x] 5.2 Run the local coverage check and confirm it enforces the 100% threshold.
- [x] 5.3 Run the local lint command and confirm it passes.
- [x] 5.4 Run `openspec status --change "prepare-go-project-ci"` and confirm the change is apply-ready.

## 6. GitHub Actions Runtime Remediation

- [x] 6.1 Update the lint workflow step to use a maintained `golangci/golangci-lint-action` major version compatible with the default GitHub Actions runner Node runtime.
- [x] 6.2 Confirm the workflow does not set `ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION` or otherwise opt in to deprecated Node runtimes.
- [x] 6.3 Re-run the GitHub Actions workflow and confirm the lint step starts without the Node 20 deprecation failure.
