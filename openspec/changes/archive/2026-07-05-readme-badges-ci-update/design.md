## Context

The `.github/workflows/go-quality.yml` workflow uploads coverage to Codecov on every run — pushes to `main`/`brainstorming` and PRs to `main`. This causes duplicate uploads per commit (the PR branch upload AND the merge commit upload) and skews Codecov data with intermediate PR coverage states.

## Goals / Non-Goals

**Goals:**
- Upload coverage to Codecov only when pushing to `main`
- PR workflow still runs coverage checks and fails below 100%, but does not upload

**Non-Goals:**
- Changing the coverage threshold or test execution
- Changing the Codecov action version or configuration

## Decisions

### D1: Branch Gate Approach

**Choice:** Use `if: github.ref == 'refs/heads/main'` on the Codecov step.

**Rationale:** This is the simplest, most idiomatic GitHub Actions approach. It matches the brainstorming doc's recommendation and is consistent with the existing `upload-artifact` step which already uses `if: always()`.

### D2: Keep Coverage Check on PRs

**Choice:** The coverage check (`Check coverage` step) continues to run on all events. Only the Codecov upload is gated.

**Rationale:** PRs must still enforce 100% coverage. Removing that would degrade quality. The Codecov upload is purely for the badge and historical tracking.

## Risks / Trade-offs

- None. This is a one-line conditional addition with no breaking changes.
