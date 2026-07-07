## 1. Core update library

- [x] 1.1 Create `pkg/update/updater.go` with `Version` var (default `"dev"`), `CurrentVersion()`, and semver comparison helpers
- [x] 1.2 Implement `LatestRelease()` — fetch and parse `GET /repos/afadhitya/wallet-app/releases/latest` from GitHub API (filter prereleases)
- [x] 1.3 Implement platform asset matching — construct `wallet_{{ .Os }}_{{ .Arch }}.tar.gz` from runtime GOOS/GOARCH, match against release assets
- [x] 1.4 Implement `DownloadAndVerify()` — download matched archive and `checksums.txt`, verify SHA256 digest, extract `wallet` binary from tar.gz to temp file
- [x] 1.5 Implement `ReplaceBinary()` — discover current binary via `os.Executable()`, write new binary to `<binary>.new`, atomic `os.Rename`

## 2. Build system integration

- [x] 2.1 Add `VERSION` variable and ldflags to `build` and `install` targets in `Makefile`
- [x] 2.2 Append version ldflag (`-X github.com/afadhitya/wallet-app/pkg/update.Version={{ .Version }}`) to `.goreleaser.yml`

## 3. CLI commands

- [x] 3.1 Create `internal/cli/version.go` — `wallet version` command with `--check` flag, text and JSON output
- [x] 3.2 Create `internal/cli/update.go` — `wallet update` command with `--force` flag, flows: check version → download → verify → replace → report
- [x] 3.3 Register `newVersionCmd()` and `newUpdateCmd()` in `internal/cli/root.go`
- [x] 3.4 Add error classification for update error codes (`UPDATE_FAILED`, `UPDATE_CHECKSUM_MISMATCH`, `UPDATE_NETWORK_ERROR`, `UPDATE_PERMISSION_ERROR`, `UPDATE_ALREADY_LATEST`) in `internal/cli/helpers.go`

## 4. Unit tests — pkg/update

- [x] 4.1 Create `pkg/update/updater_test.go` with version comparison and asset matching tests
- [x] 4.2 Test `LatestRelease()` with `httptest.NewServer` mock: success, no assets, prerelease filtering, non-200, malformed JSON
- [x] 4.3 Test checksum verification against a known checksums file (valid and invalid digests)
- [x] 4.4 Test archive extraction with `pkg/update/testdata/wallet_darwin_amd64.tar.gz` fixture
- [x] 4.5 Create test fixtures: `pkg/update/testdata/wallet_darwin_amd64.tar.gz` and `pkg/update/testdata/checksums.txt`

## 5. CLI tests

- [x] 5.1 Create `internal/cli/version_test.go` — inject known version, verify text and JSON output, `--check` flag paths
- [x] 5.2 Create `internal/cli/update_test.go` — inject mock updater, verify success/error paths, JSON output structure

## 6. Verification and cleanup

- [x] 6.1 Add OS-level failure branches (rename permission errors, network timeouts) to `coverignore.txt`
- [x] 6.2 Run `make coverage-check` to verify 100% coverage
- [x] 6.3 Run `make lint` and fix any issues
- [x] 6.4 Run `make test` to verify all tests pass
