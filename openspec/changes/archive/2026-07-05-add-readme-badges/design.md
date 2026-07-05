## Context

The README.md at the repo root already has a `# Wallet App` heading followed by a project description. The goal is to insert 5 professional badges between the heading and the description. No infrastructure changes are needed — all external services (GitHub Actions, Codecov, GitHub Releases) are already active.

## Goals / Non-Goals

**Goals:**
- Add 5 badges showing Go version, CI status, code coverage, license, and latest release
- Use flat style for clean minimal aesthetic matching the project
- Order badges as: Version → CI → Coverage → License → Release
- Links must point to correct GitHub/Codecov/shields.io URLs for the `afadhitya/wallet-app` repo

**Non-Goals:**
- Adding more than 5 badges (no Go Report Card, last commit, PRs welcome, etc.)
- Changing README structure or content beyond badge insertion
- Setting up new CI/CD or Codecov integration (already done)

## Decisions

### D1: Badge Selection
5 badges selected: Go Version, CI Status, Code Coverage, License, Release.
Rejected: Go Report Card (uses deprecated golint), Last Commit (not essential), PRs Welcome (not essential).

### D2: Badge Style
Chose flat style (`?style=flat`). Alternative: flat-square or for-the-badge. Flat is the most minimal and matches the project's clean CLI aesthetic.

### D3: Badge Order
Chose Version → CI → Coverage → License → Release. Shows identity first (what version), then health (CI, coverage), then legal (license), then discovery (release). Alternative ordering (health-first, legal-first) was considered but identity-first is more user-centric.

### D4: Placement
Badges go on a single line immediately after `# Wallet App`, before the description paragraph. Each badge is an image link wrapped in a URL link. No table or list formatting — inline images keep the visual compact.

## Risks / Trade-offs

- [Badges depend on external services] → shields.io is a well-established free service with high uptime. GitHub badges are served by GitHub itself. All URLs are static and will gracefully degrade to broken images if services are down — no functionality impact.
- [Release badge requires at least one GitHub Release] → The project already has releases, so the badge will resolve immediately.
- [Go version badge reads from go.mod] → Will auto-update when go.mod is bumped. No maintenance burden.
