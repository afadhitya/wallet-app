## ADDED Requirements

### Requirement: Display current version
The system SHALL provide a `wallet version` command that prints the current application version injected via ldflags at build time.

#### Scenario: Display version in text mode
- **WHEN** user runs `wallet version`
- **THEN** the current version string (e.g. `v1.2.0`) is printed to stdout

#### Scenario: Display version in JSON mode
- **WHEN** user runs `wallet version --json`
- **THEN** output is `{"success":true,"data":{"version":"<current>"}}`

### Requirement: Check for latest version
The system SHALL support a `--check` flag on `wallet version` that queries the GitHub Releases API for the latest stable release and reports whether an update is available.

#### Scenario: Update available in text mode
- **WHEN** user runs `wallet version --check` and a newer stable release exists
- **THEN** the current version is printed along with the latest version and an upgrade hint

#### Scenario: Already latest in text mode
- **WHEN** user runs `wallet version --check` and the current version matches the latest stable release
- **THEN** a message indicates the CLI is up to date

#### Scenario: Check in JSON mode
- **WHEN** user runs `wallet version --check --json` and a newer release exists
- **THEN** output is `{"success":true,"data":{"version":"<current>","latest":"<latest>","update_available":true}}`

#### Scenario: Network failure during check
- **WHEN** user runs `wallet version --check` and the GitHub API is unreachable or rate-limited
- **THEN** an error is reported indicating the check is temporarily unavailable

### Requirement: Self-update from GitHub Releases
The system SHALL provide a `wallet update` command that downloads the latest stable release archive for the current platform from GitHub Releases, verifies its SHA256 checksum, extracts the binary, and atomically replaces the running executable.

#### Scenario: Successful update from older version
- **WHEN** user runs `wallet update` and a newer stable release exists
- **THEN** the latest archive is downloaded, checksum verified, binary extracted, and the running binary is atomically replaced

#### Scenario: Already at latest version
- **WHEN** user runs `wallet update` and the current version is already the latest stable release
- **THEN** a message indicates no update is needed and the command exits successfully

#### Scenario: Update with checksum mismatch
- **WHEN** user runs `wallet update` and the downloaded archive's SHA256 does not match the entry in checksums.txt
- **THEN** the command fails with error code `UPDATE_CHECKSUM_MISMATCH` and the running binary is not modified

#### Scenario: Update with network failure
- **WHEN** user runs `wallet update` and the GitHub API or download fails
- **THEN** the command fails with error code `UPDATE_NETWORK_ERROR`

#### Scenario: Update with permission error
- **WHEN** user runs `wallet update` and the process lacks write permission to the binary directory
- **THEN** the command fails with error code `UPDATE_PERMISSION_ERROR`

#### Scenario: Force update when already latest
- **WHEN** user runs `wallet update --force`
- **THEN** the update proceeds regardless of whether the current version matches the latest release

#### Scenario: JSON output on success
- **WHEN** user runs `wallet update --json` and the update succeeds
- **THEN** output is `{"success":true,"data":{"previous":"<old>","current":"<new>"}}`

#### Scenario: JSON output on failure
- **WHEN** user runs `wallet update --json` and the update fails
- **THEN** output is `{"success":false,"error":{"code":"<ERROR_CODE>","message":"<description>"}}`

### Requirement: Atomic binary replacement
The system SHALL replace the running binary atomically by writing the new binary to a temporary file in the same directory and performing an atomic rename, ensuring a partially written binary is never observed.

#### Scenario: New binary written alongside existing
- **WHEN** the new binary is extracted from the archive
- **THEN** it is written to `<binary>.new` in the same directory as the current executable

#### Scenario: Atomic rename
- **WHEN** the new binary is fully written and verified
- **THEN** `<binary>.new` is atomically renamed to `<binary>` via os.Rename

### Requirement: Version string injection at build time
The system SHALL read its version from a compile-time variable (`pkg/update.Version`) set via Go linker flags (`-ldflags`).

#### Scenario: Version injected at build
- **WHEN** the binary is built with `-ldflags "-X github.com/afadhitya/wallet-app/pkg/update.Version=v1.2.0"`
- **THEN** `wallet version` prints `v1.2.0`

#### Scenario: Dev version when not injected
- **WHEN** the binary is built without a version ldflag (e.g. `go build` with no flags)
- **THEN** `wallet version` prints `dev`

### Requirement: Platform-aware asset matching
The system SHALL match the correct GitHub Release asset for the current platform by constructing the filename pattern `wallet_{{ .Os }}_{{ .Arch }}.tar.gz` from runtime GOOS and GOARCH constants.

#### Scenario: Darwin amd64 asset matched
- **WHEN** the system is running on darwin/amd64
- **THEN** the asset `wallet_darwin_amd64.tar.gz` is selected from the release assets

#### Scenario: No matching asset found
- **WHEN** user runs `wallet update` and no release asset matches the current platform
- **THEN** the command fails with error code `UPDATE_FAILED` indicating no compatible binary is available
