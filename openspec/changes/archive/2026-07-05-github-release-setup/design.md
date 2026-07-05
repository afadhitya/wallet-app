## Context

The wallet-app has no release infrastructure. The existing CI (`go-quality.yml`) handles lint, test, and coverage on push/PR to main, but stops there. The Makefile has a `build` target (`go build -o bin/wallet ./cmd/wallet`) that produces a single binary for the current OS/arch only. The README references a GitHub Releases page and includes a release badge, but no releases exist — the badge currently returns 404.

The project is feature-complete (12 capabilities, 100% coverage, full CLI, documented) and ready for its first public release. Users need a way to download pre-built binaries for their platform without installing Go toolchain.

## Goals / Non-Goals

**Goals:**
- Automatically cross-compile and publish release binaries on version tag push
- Provide a human-readable CHANGELOG documenting what shipped in each release
- Support the three major platforms: Linux, macOS, Windows (amd64 + arm64 for Linux/macOS)
- Generate checksums for binary integrity verification
- Use v1.0.0 as the first release version (semver)

**Non-Goals:**
- Homebrew tap or formula distribution
- Docker image publishing
- APT/RPM/other package manager repos
- Automatic version bumping from conventional commits
- Signed/notarized macOS binaries (requires Apple Developer account)
- Release towing (creating releases from any branch other than the tag's commit)
- Nightly/edge/canary release channels

## Decisions

### 1. Use goreleaser for release automation

**Choice**: goreleaser
**Alternatives considered**:
- Manual bash script + `gh release create` — simple but requires maintaining cross-compilation logic, checksum generation, and GitHub API calls. Becomes a maintenance burden.
- GoReleaser — de facto standard for Go projects. Handles cross-compilation, binary naming (`wallet_Linux_x86_64.tar.gz`), checksums (SHA256), and GitHub Release creation in a single declarative YAML config. One CI step, zero shell script maintenance.

**Rationale**: goreleaser is purpose-built for this, widely adopted in Go ecosystem, and eliminates script maintenance. The config file is self-documenting.

### 2. Cross-compilation matrix

**Choice**:
| OS | Architectures |
|----|---------------|
| linux | amd64, arm64 |
| darwin | amd64, arm64 |
| windows | amd64 |

**Rationale**: This covers 99%+ of developer/CLI-tool users. Dropped 32-bit (386) and darwin arm64 covers Apple Silicon. Windows arm64 excluded due to negligible Go CLI tool user base.

### 3. Manually curated CHANGELOG

**Choice**: Hand-written `CHANGELOG.md` using Keep a Changelog-inspired format, grouped by conventional commit type (Added/Changed)
**Alternatives considered**:
- Auto-generated from git log — noisy, includes chore/docs/refactor commits users don't care about
- goreleaser changelog integration — relies on conventional commit parsing; less curated but automated

**Rationale**: This is the first release. One curated CHANGELOG entry covering all features since project inception is more user-friendly than an auto-generated dump of 22 commits. Future releases can reconsider goreleaser's built-in changelog if the manual process becomes burdensome.

### 4. Release trigger: Git tag push

**Choice**: CI workflow triggered `on: push: tags: ['v*']`
**Rationale**: Tags are the canonical release trigger in Git. A push to `v1.0.0` unambiguously means "ship this commit." No need for additional GitHub UI interaction — goreleaser creates the GitHub Release automatically.

### 5. Version: v1.0.0

**Choice**: Start at `v1.0.0` (semver)
**Rationale**: The project is not pre-release. It has a stable CLI surface, comprehensive documentation, 100% test coverage, linting enforcement, and all 12 planned capabilities shipped. A `v0.x.y` would signal instability that doesn't reflect reality.

## Risks / Trade-offs

**[Risk] Release CI fails silently on tag push** → Debugging requires inspecting CI logs. Mitigation: CI workflow runs tests before release step — a failing test blocks the release. Also, the workflow runs on `push` to any branch matching `v*`, allowing pre-release testing on a branch before creating the actual tag.

**[Risk] goreleaser version drift** → CI uses `goreleaser/goreleaser-action` which auto-selects latest. Mitigation: pin to a major version (`goreleaser/goreleaser-action@v6`) for reproducible builds.

**[Risk] First release discoverability** → The release badge in README currently 404s. Mitigation: Once v1.0.0 is published, the badge resolves automatically. No code change needed.

**[Trade-off] Manual CHANGELOG vs automated** → Manual is more curated but requires remembering to update. For a solo-maintainer project shipping infrequently, this is acceptable. If release cadence increases, revisit goreleaser's changelog integration.
