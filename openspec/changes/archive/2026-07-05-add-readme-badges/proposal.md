## Why

The README is the first thing users and contributors see. Adding professional badges gives an immediate visual signal of project health — Go version, CI status, code coverage, license, and latest release — building trust and credibility at a glance.

## What Changes

- Add 5 badges to README.md header: Go Version, CI Status, Code Coverage, License, Release
- Badges use flat style for a clean, minimal aesthetic
- Badge order: Version → Build → Coverage → License → Release (identity first)
- No code changes — documentation-only change

## Capabilities

### New Capabilities
- `readme-badges`: Add professional status badges to the README header showing Go version, CI status, code coverage, license, and latest release

### Modified Capabilities
<!-- No existing spec requirements are changing -->

## Impact

- **Affected files**: `README.md` (insert 5 badge links below the `# Wallet App` heading)
- **External dependencies**: GitHub Actions (already active), Codecov (already active), GitHub Releases (already active), shields.io (free badge service, no setup required)
