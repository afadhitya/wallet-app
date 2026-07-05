## Purpose

TBD - Captures release automation requirements including goreleaser configuration, CI release workflow, changelog format, version tagging, and release badge.

## Requirements

### Requirement: goreleaser configuration for cross-compilation
The project SHALL include a `.goreleaser.yml` configuration that produces release binaries for linux (amd64, arm64), darwin (amd64, arm64), and windows (amd64) with SHA256 checksums, tar.gz archives, and GitHub Release creation.

#### Scenario: goreleaser config targets all supported platforms
- **WHEN** `.goreleaser.yml` is parsed
- **THEN** it defines builds for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, and windows/amd64
- **AND** archives are named `wallet_<OS>_<Arch>.tar.gz` (`.zip` for windows)
- **AND** a `checksums.txt` file is generated with SHA256 hashes

### Requirement: Release CI workflow
The project SHALL include a `.github/workflows/release.yml` workflow that triggers on `v*` tag pushes, runs lint and tests, then builds and publishes release artifacts via goreleaser.

#### Scenario: Release workflow triggers on version tag
- **WHEN** a tag matching `v*` is pushed (e.g., `v1.0.0`)
- **THEN** the release workflow executes in CI
- **AND** it runs `go test` and `golangci-lint` before the release step

#### Scenario: Release workflow creates GitHub Release
- **WHEN** tests and lint pass on a `v*` tag push
- **THEN** goreleaser creates a GitHub Release with the tag name
- **AND** cross-compiled binaries and checksums are attached as release assets

#### Scenario: Failing tests block release
- **WHEN** tests fail during the release workflow
- **THEN** goreleaser does not execute
- **AND** no GitHub Release is created

### Requirement: CHANGELOG format
The project SHALL include a `CHANGELOG.md` following a Keep a Changelog-inspired format, with sections grouped by category (Added, Changed, Fixed) and version headers.

#### Scenario: v1.0.0 changelog entry documents all features
- **WHEN** reading `CHANGELOG.md`
- **THEN** it contains a `## [v1.0.0]` section with subsections for Added features
- **AND** all major features are listed: accounts, transactions, budgets, planned payments, forecasting, reports, multi-currency, AI-native CLI

### Requirement: First release version
The project's first release SHALL be tagged `v1.0.0` following semantic versioning.

#### Scenario: v1.0.0 tag exists
- **WHEN** listing git tags
- **THEN** `v1.0.0` is present
- **AND** it points to the latest commit on main

### Requirement: README release badge resolves
The README's release badge SHALL resolve to the latest GitHub Release without modification after the first release is published.

#### Scenario: Badge shows v1.0.0 after release
- **WHEN** visiting the GitHub Releases page after `v1.0.0` is published
- **THEN** the release badge in README displays `v1.0.0`
