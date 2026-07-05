## Why

The CI workflow currently uploads coverage reports to Codecov on every push and PR, causing duplicate uploads and noisy Codecov data. Coverage upload should only happen on the `main` branch so the Codecov badge reflects the canonical main branch coverage, while PRs still run coverage checks locally without uploading.

## What Changes

- Add `if: github.ref == 'refs/heads/main'` to the Codecov upload step in `.github/workflows/go-quality.yml`
- PRs continue to run coverage checks and fail below 100%, but no longer upload to Codecov

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `go-project-quality`: Coverage upload to Codecov is now restricted to the `main` branch only, preventing duplicate uploads from PR branches

## Impact

- `.github/workflows/go-quality.yml` — one-line addition to the Codecov upload step
- No code or service changes
