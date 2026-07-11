## ADDED Requirements

### Requirement: Service layer is organized into domain sub-packages
The `internal/service/` directory SHALL contain sub-packages for each domain (transaction, plannedpayment, budget, report) and a shared sub-package for cross-cutting utilities.

#### Scenario: Each domain has its own sub-package
- **WHEN** a developer navigates `internal/service/`
- **THEN** they SHALL find directories `transaction/`, `plannedpayment/`, `budget/`, `report/`, and `shared/`

#### Scenario: No file exceeds 500 lines in any sub-package
- **WHEN** counting lines in any `.go` file within sub-packages
- **THEN** each file SHALL be under 500 lines

### Requirement: Cross-cutting types live in shared sub-package
Error types (`NotFoundError`, `ValidationError`), sentinel errors, resolver helpers, currency helpers, and date parsing SHALL reside in `internal/service/shared/` and SHALL NOT import the parent `service` package.

#### Scenario: Shared package has no dependency on parent service
- **WHEN** checking imports of `internal/service/shared/`
- **THEN** no file SHALL import `internal/service`

### Requirement: Sub-package Managers embed into Service struct
Each domain manager (`*transaction.Manager`, `*plannedpayment.Manager`, `*budget.Manager`, `*report.Manager`) SHALL be embedded as promoted fields in the `Service` struct so all methods are accessible via `*Service` without delegation methods.

#### Scenario: CLI callers see no API change
- **WHEN** CLI code calls `svc.AddExpense(params)`
- **THEN** the call SHALL resolve through the embedded `*transaction.Manager` without code changes in `internal/cli/`

### Requirement: Existing behavior is preserved
All CLI commands, business rules, outputs, and test assertions SHALL remain identical to the pre-refactor state. The refactor is purely structural.

#### Scenario: Full test suite passes with identical assertions
- **WHEN** running `make coverage-check` after the refactor
- **THEN** all tests SHALL pass with 100% coverage maintained
