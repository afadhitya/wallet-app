## 1. CHANGELOG

- [x] 1.1 Create `CHANGELOG.md` with `## [v1.0.0]` section
- [x] 1.2 Document all features grouped by domain (accounts, transactions, budgets, planned payments, forecasting, reports, multi-currency, AI-native CLI)

## 2. Goreleaser Configuration

- [x] 2.1 Create `.goreleaser.yml` with project metadata and build configuration
- [x] 2.2 Configure cross-compilation targets: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
- [x] 2.3 Configure archive naming (`wallet_<OS>_<Arch>.tar.gz` / `.zip`) and SHA256 checksums

## 3. Release CI Workflow

- [x] 3.1 Create `.github/workflows/release.yml`
- [x] 3.2 Configure workflow trigger: `on: push: tags: ['v*']`
- [x] 3.3 Add lint, test, and goreleaser steps (test must pass before release)
- [x] 3.4 Verify `GITHUB_TOKEN` permissions allow release creation

## 4. Tag and Publish

- [x] 4.1 Create and push `v1.0.0` annotated tag on main
- [ ] 4.2 Verify CI runs and publishes release with binaries and checksums
- [ ] 4.3 Verify README release badge resolves correctly
