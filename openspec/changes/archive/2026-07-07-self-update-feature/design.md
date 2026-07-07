## Context

The wallet CLI is distributed via GitHub Releases with platform-specific tar.gz archives and SHA256 checksums. Users currently have no built-in mechanism to check their installed version or upgrade to the latest release. This design adds self-update capabilities following the existing layered architecture (CLI → service → data access), with the update logic in a dedicated `pkg/update/` library.

The app uses ldflags for compile-time configuration. The update flow must handle HTTP interaction, checksum verification, tar.gz extraction, and atomic file replacement — all using only Go standard library packages (no external dependencies).

## Goals / Non-Goals

**Goals:**
- Display current version via `wallet version` command
- Check for latest GitHub release with `wallet version --check`
- Download, verify (SHA256), extract, and atomically replace binary via `wallet update`
- Support `--force` flag to bypass version check
- Consistent JSON output format with machine-readable error codes
- 100% test coverage for business logic (version comparison, asset matching, checksum verification, archive extraction)

**Non-Goals:**
- Authenticated GitHub API requests (OAuth token, GITHUB_TOKEN) — unauthenticated 60 req/hr is sufficient
- Background auto-update or update-on-launch daemon
- Rolling back to previous version after update
- Cross-platform concerns beyond GOOS/GOARCH asset matching
- Support for archive formats other than tar.gz

## Decisions

### Decision 1: New `pkg/update/` package instead of extending CLI directly
**Rationale:** Update logic (API interaction, checksums, archive extraction, binary replacement) is distinct from command presentation. A separate package enables unit testing in isolation from Cobra commands, follows the existing pattern of `pkg/config/` for cross-cutting concerns, and keeps CLI files focused on argument parsing and output formatting.

**Alternatives considered:** Embedding all logic in `internal/cli/` — rejected because it would mix concerns and complicate testing (CLI tests would need to mock HTTP, filesystem, etc.).

### Decision 2: Go standard library only for update package
**Rationale:** The update flow needs HTTP (net/http), JSON (encoding/json), SHA256 (crypto/sha256), archive extraction (archive/tar, compress/gzip), and file I/O (os, io) — all available in stdlib. Adding a third-party HTTP or archive library adds dependency risk for no benefit.

**Alternatives considered:** Using `github.com/hashicorp/go-version` for semver — rejected because simple string-based comparison of tag names (e.g. `v1.2.0`) is sufficient and avoids a dependency.

### Decision 3: Atomic rename strategy for binary replacement
**Rationale:** Write the new binary to `<binary>.new`, then `os.Rename()` to `<binary>`. On Unix, rename is atomic within the same filesystem — a partially written binary is never observed. This is the same approach used by tools like `rustup` and `goreleaser`.

**Alternatives considered:** Using `os.Remove()` + `os.Rename()` — rejected because it introduces a window where no binary exists. Copy-overwrite — rejected because a crash during write leaves a corrupt binary.

### Decision 4: Version via ldflags instead of embedded file or runtime detection
**Rationale:** The app already uses ldflags for build-time configuration. Injecting `pkg/update.Version` via `-X` follows this pattern and requires no file I/O at startup. Dev builds default to `"dev"`.

**Alternatives considered:** Embedding version in an auto-generated Go file — rejected as it adds build complexity without benefit. Parsing binary metadata at runtime — rejected as unreliable and platform-specific.

### Decision 5: `checksums.txt` separate from archive download
**Rationale:** GoReleaser generates `checksums.txt` alongside assets. Downloading it separately and verifying before extraction ensures integrity even if the archive download is intercepted or corrupted. This is the standard GitHub Releases verification pattern.

### Decision 6: Cobra commands in `internal/cli/` with service-layer orchestration
**Rationale:** Follow existing architecture — `version.go` and `update.go` in `internal/cli/` are Cobra command files. The update service layer in `pkg/update/` provides stateless functions. The CLI layer handles `--json` output formatting and error classification (e.g., `classifyError()` for error codes). This maintains consistency with how `add.go`, `budget.go`, etc., are structured.

## Risks / Trade-offs

**[Risk] Binary replacement may fail if the process is running from a read-only filesystem**
→ Mitigation: Error code `UPDATE_PERMISSION_ERROR` with clear message; user can manually download.

**[Risk] GitHub API rate limit (60 req/hr unauthenticated) could block update checks during heavy usage**
→ Mitigation: Unlikely for a CLI tool; command reports "temporarily unavailable" rather than failing silently. Can add optional token support later.

**[Risk] Running the new binary after update requires process restart — `exec` is not used**
→ Mitigation: The command prints a success message. The user must re-run the command themselves. This is intentional — automatic exec would lose terminal state and is surprising behavior.

**[Risk] No rollback capability**
→ Mitigation: The previous binary is overwritten. Users can manually download older releases. A rollback feature is a non-goal for v1 and can be added later if needed.

**[Risk] Checksums.txt tampering (if GitHub Releases itself is compromised)**
→ Mitigation: This is a GitHub-level threat. HTTPS provides transport security. GPG signing could be added in the future for additional trust.

## Open Questions

- None. All technical decisions are resolved by the existing spec.
