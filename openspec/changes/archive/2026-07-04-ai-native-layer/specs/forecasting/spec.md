## MODIFIED Requirements

### Requirement: Forecast JSON Output
The system SHALL support machine-readable AI-native JSON output for forecast commands when global `--json` is supplied and SHALL use the shared JSON response envelope.

#### Scenario: Render balance forecast JSON
- **WHEN** the user runs `wallet forecast --months 3 --json`
- **THEN** the system writes a JSON envelope with `success: true`
- **AND** `data` contains the forecast horizon, forecast rows, planned payment details, totals, and warnings
- **AND** `meta.command` identifies the forecast command
- **AND** the response does not include table formatting in the response

#### Scenario: Render bills forecast JSON
- **WHEN** the user runs `wallet forecast bills --json`
- **THEN** the system writes a JSON envelope with `success: true`
- **AND** `data` contains bill rows, running totals, total amount, count, and horizon
- **AND** `meta.command` identifies the forecast bills command
- **AND** the response does not include table formatting in the response

#### Scenario: Forecast errors return envelope JSON
- **WHEN** the user runs a forecast command with `--json` and provides an invalid horizon or unknown account
- **THEN** the system exits with a non-zero status
- **AND** writes a JSON envelope with `success: false`
- **AND** `error.code` identifies the invalid horizon or missing account
- **AND** `error.message` describes the failure without table formatting
