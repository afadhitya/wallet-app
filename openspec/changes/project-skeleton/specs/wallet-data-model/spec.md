## MODIFIED Requirements

### Requirement: Database Initialization
The application SHALL provide a way to initialize a SQLite wallet database with the approved core schema through embedded migrations.

#### Scenario: Initialize empty wallet database
- **WHEN** the database initialization runs against an empty SQLite database
- **THEN** it creates the core wallet tables for accounts, categories, tags, transactions, transaction tags, budgets, budget categories, budget tags, planned payments, and exchange rates
- **AND** foreign key enforcement is enabled for the connection used by the application
- **AND** the initial schema is applied from an embedded SQL migration file

#### Scenario: Track applied schema version
- **WHEN** the database initialization successfully applies a migration
- **THEN** it records the applied migration version so subsequent startup runs do not reapply the same migration
