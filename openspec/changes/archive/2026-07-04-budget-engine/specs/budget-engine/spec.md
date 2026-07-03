## ADDED Requirements

### Requirement: Budget Set Command
The system SHALL provide `wallet budget set` to create or update a budget with a name, positive amount, one or more category or tag targets, period selection, optional explicit dates, notification percentage, and JSON output support.

#### Scenario: Create monthly budget with categories
- **WHEN** the user runs `wallet budget set "Monthly Food" 2000000 -c food -c transport --period monthly`
- **THEN** the system creates an active budget named `Monthly Food` with amount `2000000`, currency `IDR`, period type `monthly`, current-month start and end dates, and notification threshold `80`
- **AND** links the resolved `food` and `transport` categories as budget targets
- **AND** prints a success message or JSON representation according to the output mode

#### Scenario: Create budget with mixed category and tag targets
- **WHEN** the user runs `wallet budget set "Japan Trip" 10000000 -c travel -t japan-2026 --period one_time --from 2026-01-01 --to 2026-12-31`
- **THEN** the system creates a budget with both the resolved category and tag targets
- **AND** later budget spending calculations consider both target sets

#### Scenario: Update existing budget for same name and period
- **WHEN** a budget named `Monthly Food` already exists for the selected period
- **AND** the user runs `wallet budget set "Monthly Food" 2500000 -c food --notify 75 --period monthly`
- **THEN** the system updates that budget's amount, notification threshold, period metadata, and target links
- **AND** does not create a duplicate budget row for the same name and period

#### Scenario: Reject budget without targets
- **WHEN** the user runs `wallet budget set "Untargeted" 1000000 --period monthly`
- **THEN** the system exits with a non-zero status
- **AND** reports that a budget must have at least one category or tag
- **AND** does not create a budget

#### Scenario: Reject invalid budget amount
- **WHEN** the user runs `wallet budget set "Food" 0 -c food`
- **THEN** the system exits with a non-zero status
- **AND** reports that the amount must be positive
- **AND** does not create a budget

#### Scenario: Reject one-time budget without explicit dates
- **WHEN** the user runs `wallet budget set "Trip" 10000000 -t japan-2026 --period one_time`
- **THEN** the system exits with a non-zero status
- **AND** reports that one-time budgets require both `--from` and `--to`

#### Scenario: Reject unsupported budget period
- **WHEN** the user runs `wallet budget set "Food" 1000000 -c food --period daily`
- **THEN** the system exits with a non-zero status
- **AND** reports the supported periods `monthly`, `weekly`, `yearly`, and `one_time`

### Requirement: Budget Period Calculation
The system SHALL calculate budget period boundaries from the requested period type unless explicit dates are required or supplied.

#### Scenario: Monthly period defaults to current month
- **WHEN** the user creates or checks a monthly budget without `--from` or `--to`
- **THEN** the system uses the first day of the current month as `period_start`
- **AND** uses the last day of the current month as `period_end`

#### Scenario: Weekly period defaults to current week
- **WHEN** the user creates or checks a weekly budget without `--from` or `--to`
- **THEN** the system uses Monday of the current week as `period_start`
- **AND** uses Sunday of the current week as `period_end`

#### Scenario: Yearly period defaults to current year
- **WHEN** the user creates or checks a yearly budget without `--from` or `--to`
- **THEN** the system uses January 1 of the current year as `period_start`
- **AND** uses December 31 of the current year as `period_end`

#### Scenario: Explicit dates override calculated dates
- **WHEN** the user creates a non-`one_time` budget with both `--from` and `--to`
- **THEN** the system stores the supplied dates as the budget period boundaries

### Requirement: Budget List Command
The system SHALL provide `wallet budget list` to list budgets with their limit, spent amount, remaining amount, active state filtering, and JSON output support.

#### Scenario: List active budgets by default
- **WHEN** the user runs `wallet budget list`
- **THEN** the system lists active budgets only
- **AND** includes each budget's ID, name, limit, spent amount, and remaining amount in text output
- **AND** calculates spent from current budget target matches before rendering

#### Scenario: List all budgets
- **WHEN** the user runs `wallet budget list --all`
- **THEN** the system includes inactive and expired budgets in the result

#### Scenario: List budgets as JSON
- **WHEN** the user runs `wallet budget list --json`
- **THEN** the system writes a JSON array with stable fields for ID, name, amount, spent, remaining, period, targets, notification threshold, and active state

### Requirement: Budget Check Command
The system SHALL provide `wallet budget check` to report spending progress for one budget or all active budgets.

#### Scenario: Check all active budgets
- **WHEN** the user runs `wallet budget check --all`
- **THEN** the system checks every active budget for the current applicable period
- **AND** reports each budget's limit, spent amount, percent used, and status

#### Scenario: Check one budget by ID or name
- **WHEN** the user runs `wallet budget check --budget "Monthly Food"`
- **THEN** the system checks only the matching active budget
- **AND** exits with a non-zero status if no matching budget exists

#### Scenario: Status is ok below notification threshold
- **WHEN** a budget has notification threshold `80` and spending is less than `80%` of its limit
- **THEN** budget check reports status `ok`

#### Scenario: Status is warning at or above notification threshold
- **WHEN** a budget has notification threshold `80` and spending is at least `80%` and less than `100%` of its limit
- **THEN** budget check reports status `warning`

#### Scenario: Status is over at or above limit
- **WHEN** spending is greater than or equal to `100%` of the budget limit
- **THEN** budget check reports status `over`

#### Scenario: Check budgets as JSON
- **WHEN** the user runs `wallet budget check --all --json`
- **THEN** the system writes a JSON array with stable fields for budget ID, name, limit, spent, remaining, percent used, and status

### Requirement: Budget Spending Calculation
The system SHALL calculate budget spending from non-archived expense transactions that match the budget's category targets, tag targets, or both within the budget period.

#### Scenario: Category target spending is included
- **WHEN** a budget targets category `Food`
- **AND** an expense transaction in the budget period uses category `Food`
- **THEN** the transaction amount contributes to the budget spent amount

#### Scenario: Tag target spending is included
- **WHEN** a budget targets tag `japan-2026`
- **AND** an expense transaction in the budget period is linked to tag `japan-2026`
- **THEN** the transaction amount contributes to the budget spent amount

#### Scenario: Income, transfer, adjustment, and archived transactions are excluded
- **WHEN** transactions match a budget's category or tag targets but are not non-archived expenses
- **THEN** those transactions do not contribute to the budget spent amount

#### Scenario: Mixed target overlap is double-counted
- **WHEN** one expense transaction matches both a budget category target and a budget tag target
- **THEN** the transaction amount may be counted once for the category target and once for the tag target
- **AND** the system does not deduplicate overlap between target types

### Requirement: Budget Recurring Period Auto-Generation
The system SHALL auto-create a current recurring budget period during budget check when an active monthly, weekly, or yearly budget has a previous period but no current period.

#### Scenario: Auto-create current monthly period
- **WHEN** a monthly budget has a prior period and no budget row for the current month
- **AND** the user checks that budget or all budgets
- **THEN** the system creates a new current-month budget row with the same name, amount, currency, notification threshold, active state, and target links from the most recent prior period
- **AND** returns status for the newly created current period

#### Scenario: Do not auto-create one-time budgets
- **WHEN** a one-time budget period has expired
- **AND** the user runs `wallet budget check --all`
- **THEN** the system does not create a new budget period for that one-time budget

#### Scenario: Do not duplicate existing current period
- **WHEN** a recurring budget already has a row for the current period
- **AND** the user runs `wallet budget check --all`
- **THEN** the system uses the existing current-period row
- **AND** does not create a duplicate period

### Requirement: Budget Edit Command
The system SHALL provide `wallet budget edit <id>` to update explicitly supplied budget fields and target links while preserving unspecified fields.

#### Scenario: Edit amount and notification threshold
- **WHEN** the user runs `wallet budget edit 1 --amount 2500000 --notify 75`
- **THEN** the system updates budget `1` with the new amount and notification threshold
- **AND** preserves the budget name, period, active state, and target links

#### Scenario: Edit name and target links
- **WHEN** the user runs `wallet budget edit 1 --name "Monthly Essentials" --add-category bills --remove-tag food`
- **THEN** the system updates the budget name
- **AND** adds the resolved category target
- **AND** removes the resolved tag target

#### Scenario: Reject editing missing budget
- **WHEN** the user runs `wallet budget edit 99 --amount 2500000` and budget `99` does not exist
- **THEN** the system exits with a non-zero status
- **AND** reports that budget `99` was not found

### Requirement: Budget Removal Command
The system SHALL provide `wallet budget rm <id>` to remove a budget from active budget workflows.

#### Scenario: Remove active budget
- **WHEN** the user runs `wallet budget rm 1`
- **THEN** the system marks budget `1` inactive or otherwise removes it from default active budget results
- **AND** prints a success message identifying the removed budget

#### Scenario: Removed budget is excluded from active list and check
- **WHEN** a budget has been removed
- **THEN** `wallet budget list` excludes it by default
- **AND** `wallet budget check --all` excludes it from active checks

#### Scenario: Reject removing missing budget
- **WHEN** the user runs `wallet budget rm 99` and budget `99` does not exist
- **THEN** the system exits with a non-zero status
- **AND** reports that budget `99` was not found

### Requirement: Budget Query And Service Support
The system SHALL implement sqlc queries and service-layer operations for budget creation, updates, listing, checking, target management, recurring period generation, and spending calculation.

#### Scenario: Services validate and persist budgets
- **WHEN** CLI commands invoke budget service methods
- **THEN** the services validate domain inputs before writing
- **AND** use sqlc-generated queries for database access
- **AND** return typed results or clear errors for the CLI to render

#### Scenario: Generated queries are current
- **WHEN** budget sqlc query files are added or changed
- **THEN** generated code in `internal/gen` is regenerated and compiles with the service implementation

### Requirement: Budget Testing
The system SHALL include unit and integration tests for budget service and CLI behavior with deterministic local databases and SHALL continue to satisfy the repository's Go coverage gate after approved generated-code and documented OS/infrastructure exclusions are applied.

#### Scenario: Service tests cover budget behavior
- **WHEN** budget service tests run
- **THEN** they verify create/update, target validation, listing, checking, spending calculation, status thresholds, recurring auto-generation, editing, and removal behavior against isolated migrated SQLite databases

#### Scenario: CLI tests cover budget commands
- **WHEN** CLI integration tests run budget commands
- **THEN** they verify exit codes, stable text output, JSON output, validation errors, missing budget errors, and database side effects

#### Scenario: Coverage gate passes after budget implementation
- **WHEN** the repository coverage command runs
- **THEN** total included Go test coverage remains exactly `100%`
- **AND** any uncovered budget service, CLI, database, output, validation, or error path is either covered by focused tests or removed as unreachable code
