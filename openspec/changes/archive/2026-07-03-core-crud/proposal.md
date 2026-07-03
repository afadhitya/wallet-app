## Why

The wallet app has a data model and project skeleton, but it is not yet usable for day-to-day bookkeeping. Core CRUD commands are needed now so users can initialize storage, record transactions, manage categories and tags, reconcile balances, and list or correct their data from the CLI.

## What Changes

- Add `wallet init` to create configuration, run database migrations, and seed default categories idempotently.
- Add transaction commands for expenses, income, transfers, listing, editing, soft deletion, and balance adjustments.
- Add category CRUD commands for listing, creating, editing, and soft deleting user-defined categories.
- Add tag commands for listing, creating, and deleting tags used by transactions.
- Add service-layer operations for transactions, accounts, categories, and tags that back the CLI commands.
- Add sqlc queries and tests for the new CRUD paths, validation, and balance recalculation behavior.

## Capabilities

### New Capabilities
- `core-crud`: CLI and service behavior for initializing the wallet database, managing transactions, categories, and tags, and reconciling account balances.

### Modified Capabilities

None.

## Impact

- Affects Cobra CLI command registration and handlers.
- Adds application services around sqlc-generated database access.
- Adds or extends SQL query files for accounts, transactions, categories, tags, and transaction-tag relationships.
- Uses the existing SQLite schema and seed data from the data-model change.
- Adds unit tests for services and integration tests for CLI commands.
