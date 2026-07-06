# Self-Update Feature for Wallet CLI

Date: 2026-07-06

## Overview

Add `wallet version` and `wallet update` commands that enable the CLI to self-update by fetching the latest stable release from GitHub Releases, verifying integrity via checksums, and atomically replacing the running binary.

## Architecture

```
cmd/wallet/main.go
  ├── version.go   →  `wallet version` command
  └── update.go    →  `wallet update` command

pkg/update/
  └── updater.go   →  core update library (version, release fetching, download, verify, replace)
```

### Layer Responsibilities

- `cmd/wallet/main.go` — unchanged entry point
- `internal/cli/version.go` — Cobra command for `wallet version`, reads `pkg/update.Version`
- `internal/cli/update.go` — Cobra command for `wallet update`, orchestrates the update flow via `pkg/update`
- `pkg/update/updater.go` — stateless functions for version reading, GitHub API interaction, archive download/extraction, checksum verification, and atomic binary replacement

## Commands

### `wallet version`

```
wallet version [--check]
```

- **Default mode**: Prints current version (ldflags-injected `pkg/update.Version`).
- **`--check` flag**: Additionally queries GitHub Releases API for the latest stable version. If newer, prints an upgrade hint.
- **JSON output**: `{"success":true,"data":{"version":"v1.2.0","latest":"v1.2.1","update_available":true}}` (with `--check`), or without latest info when `--check` is omitted.

### `wallet update`

```
wallet update [--force]
```

**Flow:**
1. Read `CurrentVersion` — if already latest, print message and exit (skip if `--force`)
2. Fetch `https://api.github.com/repos/afadhitya/wallet-app/releases/latest`
3. Determine GOOS/GOARCH from runtime constants; match asset by pattern `wallet_{{ .Os }}_{{ .Arch }}.tar.gz`
4. Download the matched archive and `checksums.txt` to temporary files
5. Verify SHA256 of archive against the entry in `checksums.txt`
6. Extract the `wallet` binary from the tar.gz to a temp file
7. Replace the current binary atomically:
   - Discover current binary path via `os.Executable()`
   - Write new binary to `<binary>.new` in the same directory
   - Rename `<binary>.new` → `<binary>` (atomic rename on Unix)
8. Print success message

**JSON output:**
- Success: `{"success":true,"data":{"previous":"v1.2.0","current":"v1.2.1"}}`
- Error: `{"success":false,"error":{"code":"UPDATE_FAILED","message":"..."}}`

### Error Codes

| Code | Condition |
|------|-----------|
| `UPDATE_FAILED` | Generic update failure |
| `UPDATE_CHECKSUM_MISMATCH` | Archive digest doesn't match checksums.txt |
| `UPDATE_NETWORK_ERROR` | GitHub API or download failure |
| `UPDATE_PERMISSION_ERROR` | Cannot write to binary directory |
| `UPDATE_ALREADY_LATEST` | Already at latest version (not a hard error) |

## Version Injection

`pkg/update/updater.go` declares:

```go
package update

var Version = "dev"
```

### Build Integration

**.goreleaser.yml** — append to existing `ldflags`:
```yaml
ldflags:
  - -s -w
  - -X github.com/afadhitya/wallet-app/pkg/update.Version={{ .Version }}
```

**Makefile** — update `build` and `install`:
```makefile
VERSION := $(shell git describe --tags 2>/dev/null || echo dev)

build:
	go build -ldflags="-X github.com/afadhitya/wallet-app/pkg/update.Version=$(VERSION)" -o bin/wallet ./cmd/wallet

install:
	go install -ldflags="-X github.com/afadhitya/wallet-app/pkg/update.Version=$(VERSION)" ./cmd/wallet
```

## GitHub API Interaction

- Endpoint: `GET /repos/afadhitya/wallet-app/releases/latest`
- Filters: `prerelease: false` — only stable releases
- Response parsing: extract `tag_name`, iterate `assets` to find matching archive and `checksums.txt`
- Rate limiting: unauthenticated GitHub API allows 60 req/hr — sufficient for a CLI update command
- No auth token needed; if rate-limited, the command reports that the check is temporarily unavailable rather than failing

## Testing

### Unit Tests (`pkg/update/`)

- **Version parsing**: verify `semver` comparison functions
- **Asset matching**: given a list of assets and a GOOS/GOARCH, match the correct archive
- **Checksum verification**: test against a known checksums file and both valid/invalid digests
- **Archive extraction**: test tar.gz extraction with a fixture

### HTTP Tests (`pkg/update/`)

- `LatestRelease` tested with `httptest.NewServer` returning mock GitHub API responses
- Test cases: successful response, no assets, prerelease filtering, non-200 status, malformed JSON

### CLI Tests (`internal/cli/`)

- **version_test.go**: inject known version, verify output format for both text and JSON modes
- **update_test.go**: inject mock updater interface, verify success/error paths, JSON output structure

### Test Fixtures

- `pkg/update/testdata/wallet_darwin_amd64.tar.gz` — minimal valid archive
- `pkg/update/testdata/checksums.txt` — matching checksums file

### Coverage

- All business logic (version comparison, asset matching, checksum verification, archive extraction) must reach 100%
- OS-level failure branches (binary rename permission errors, network timeouts) go into coverignore.txt

## Implementation Order

1. Create `pkg/update/updater.go` with `Version` var, `CurrentVersion()`, `LatestRelease()`, `DownloadAndVerify()`, `ReplaceBinary()`
2. Add version injection to `Makefile` and `.goreleaser.yml`
3. Create `internal/cli/version.go` — `wallet version` command
4. Create `internal/cli/update.go` — `wallet update` command
5. Register both commands in `root.go`
6. Write unit tests for `pkg/update/`
7. Write CLI tests for `internal/cli/`
8. Run full test suite, lint, and coverage check
