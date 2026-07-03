## Why

The wallet application needs a durable SQLite data model before CLI, budgeting, forecasting, reporting, and AI-assisted features can be implemented safely. Establishing the schema now gives future phases one source of truth for accounts, transactions, budgets, planned payments, tags, categories, and exchange rates.

## What Changes

- Add a SQLite schema for the core wallet domain.
- Add account storage with denormalized balances in minor units.
- Add two-level categories and freeform tags for transaction classification.
- Add transactions that support expense, income, transfer, and balance adjustment movements.
- Add budget snapshots with category and tag targets.
- Add planned payments with precomputed next due dates.
- Add cached exchange rates for multi-currency transactions.
- Seed default system categories for common income and expense use cases.

## Capabilities

### New Capabilities
- `wallet-data-model`: Defines the SQLite schema and seed data required for the wallet application's core financial records.

### Modified Capabilities

## Impact

- Affected code: database schema/migration code, seed data, tests, and any future repository or CLI code that reads/writes wallet entities.
- APIs: establishes table and column contracts for future application features.
- Dependencies: may introduce a SQLite driver and migration approach if not already present.
- Systems: local wallet database files will be initialized with the approved schema.
