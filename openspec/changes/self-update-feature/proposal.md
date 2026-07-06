## Why

Users currently have no way to check their installed wallet CLI version or upgrade to the latest release without manually visiting GitHub Releases and downloading the binary. This adds `wallet version` and `wallet update` commands for discoverability and automatic self-update.

## What Changes

- Add `wallet version` command showing the current version (injected via ldflags at build time)
- Add `wallet version --check` to compare against the latest GitHub stable release and show an upgrade hint
- Add `wallet update` command to download, verify (SHA256 checksum), and atomically replace the running binary from GitHub Releases
- Add `wallet update --force` to reinstall even when already at latest
- Introduce `pkg/update/` library for version reading, GitHub Releases API interaction, archive download/extraction, checksum verification, and atomic binary replacement
- Inject version string via `-ldflags` in Makefile and `.goreleaser.yml`

## Capabilities

### New Capabilities

- `self-update`: Display current CLI version, check for newer releases from GitHub, and atomically self-update the running binary with checksum integrity verification.

### Modified Capabilities

<!-- No existing capability requirements are changing -->

## Impact

- New package: `pkg/update/` — core update logic, no external dependencies beyond Go stdlib
- New CLI commands: `internal/cli/version.go`, `internal/cli/update.go`
- Modified: `internal/cli/root.go` — register new subcommands
- Modified: `Makefile` — version variable, ldflags for `build` and `install` targets
- Modified: `.goreleaser.yml` — append version ldflag to existing config
- Modified: `coverignore.txt` — OS-level failure branches (permission errors, network timeouts)
- New test fixtures: `pkg/update/testdata/` — tar.gz archive and checksums.txt
