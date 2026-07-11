## MODIFIED Requirements

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

#### Scenario: Mixed target overlap is deduplicated
- **WHEN** one expense transaction matches both a budget category target and a budget tag target
- **THEN** the transaction amount is counted exactly once toward the budget spent amount
- **AND** the system deduplicates overlap between target types using a single query with OR-based matching
