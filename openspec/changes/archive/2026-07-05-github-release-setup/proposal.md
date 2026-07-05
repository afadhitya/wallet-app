## Why

The wallet-app is feature-complete with 12 major capabilities (accounts, transactions, budgets, planned payments, forecasting, reports, multi-currency, and AI-native CLI), 100% test coverage, full documentation, and linting enforcement. However, there is no way for users to discover or download it — no git tags exist, no CHANGELOG documents what shipped, no CI pipeline builds release binaries, and the README's release badge links to an empty releases page. This change establishes the release infrastructure needed to ship v1.0.0 to users.

## What Changes

- Add `CHANGELOG.md` documenting all features from inception to v1.0.0
- Add `.goreleaser.yml` configuring automated cross-compilation (linux/darwin/windows, amd64/arm64), checksums, and GitHub Release creation
- Add `.github/workflows/release.yml` CI workflow triggered on `v*` tags that runs tests/lint, then builds and publishes via goreleaser
- Tag `v1.0.0` (semver) as the first public release

## Capabilities

### New Capabilities

- `release-automation`: Automated GitHub Release pipeline — cross-compiled binaries, changelog-driven release notes, checksums, and CI-triggered publishing on version tags

### Modified Capabilities

None. This is purely additive release infrastructure — no existing specs, APIs, or CLI commands change.

## Impact

- New files: `CHANGELOG.md`, `.goreleaser.yml`, `.github/workflows/release.yml`
- New git tag: `v1.0.0`
- CI: Release workflow runs independently from existing `go-quality.yml`; no modification to existing CI
- README release badge resolves automatically once a release exists on GitHub
