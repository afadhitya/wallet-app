## MODIFIED Requirements

### Requirement: Planned Payment JSON Output
The system SHALL support AI-native JSON output for planned-payment commands when `--json` is supplied and SHALL use the shared JSON response envelope for successes and failures.

#### Scenario: Render planned payment JSON
- **WHEN** the user runs a planned-payment command with `--json`
- **THEN** the system writes a machine-readable JSON response containing `success: true`
- **AND** the response contains `data` with command result fields
- **AND** the response contains `meta.command` and `meta.timestamp`
- **AND** the response does not include table formatting in the response

#### Scenario: Bill due returns envelope JSON
- **WHEN** the user runs `wallet bill due --json`
- **THEN** the system writes a JSON envelope with `success: true`
- **AND** `data.due` contains active, unpaused planned payments due in the selected window
- **AND** `data.total_due` contains the total due amount
- **AND** `data.count` contains the number of due payments

#### Scenario: Planned payment errors return envelope JSON
- **WHEN** the user runs a planned-payment command with `--json` and references a missing, paused, already-paid, or invalid bill
- **THEN** the system exits with a non-zero status
- **AND** writes a JSON envelope with `success: false`
- **AND** `error.code` identifies the planned-payment failure condition
- **AND** `error.message` describes the failure without table formatting
