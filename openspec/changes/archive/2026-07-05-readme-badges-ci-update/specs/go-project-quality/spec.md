## ADDED Requirements

### Requirement: Codecov upload restricted to main branch
The CI workflow SHALL upload coverage reports to Codecov only when running on the `main` branch, to prevent duplicate uploads from pull request branches.

#### Scenario: Coverage uploaded to Codecov on main push
- **WHEN** the quality workflow runs on a push to the `main` branch
- **THEN** the workflow uploads the coverage profile to Codecov using `codecov/codecov-action`

#### Scenario: Coverage not uploaded to Codecov on PR
- **WHEN** the quality workflow runs on a pull request to `main`
- **THEN** the workflow runs coverage checks locally but does NOT upload to Codecov

#### Scenario: Coverage not uploaded to Codecov on brainstorming push
- **WHEN** the quality workflow runs on a push to the `brainstorming` branch
- **THEN** the workflow runs coverage checks locally but does NOT upload to Codecov
