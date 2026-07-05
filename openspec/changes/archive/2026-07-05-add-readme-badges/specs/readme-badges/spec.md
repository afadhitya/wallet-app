## ADDED Requirements

### Requirement: README displays project health badges
The README.md file SHALL display 5 badges immediately below the `# Wallet App` heading, showing Go version, CI status, code coverage, license, and latest release.

#### Scenario: Badges are visible on GitHub
- **WHEN** a user visits the repository on GitHub
- **THEN** they SHALL see 5 badges in flat style ordered as: Go Version, CI status, Code Coverage, License, Release

#### Scenario: Badges link to relevant resources
- **WHEN** a user clicks a badge
- **THEN** the Go Version badge SHALL link to go.dev, the CI badge SHALL link to the GitHub Actions workflow, the Coverage badge SHALL link to Codecov, the License badge SHALL link to the LICENSE file, and the Release badge SHALL link to GitHub Releases

#### Scenario: Go version badge reflects current go.mod
- **WHEN** the go.mod Go version is updated
- **THEN** the Go Version badge SHALL automatically reflect the new version from go.mod
