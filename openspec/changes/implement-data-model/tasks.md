## 1. Database Schema

- [x] 1.1 Choose and add the SQLite driver/migration approach used by the Go application.
- [x] 1.2 Add database initialization code that creates all approved wallet tables in dependency order.
- [x] 1.3 Enable SQLite foreign key enforcement for application database connections.
- [x] 1.4 Add indexes or uniqueness constraints needed for tag uniqueness, exchange-rate uniqueness, and idempotent category seeding.

## 2. Seed Data

- [x] 2.1 Add idempotent seed logic for approved default parent categories.
- [x] 2.2 Add idempotent seed logic for approved default child categories.
- [x] 2.3 Mark seeded categories as system categories.

## 3. Verification

- [x] 3.1 Add tests that initialize an empty SQLite database and verify all expected tables exist.
- [x] 3.2 Add tests that verify foreign key constraints are enforced.
- [x] 3.3 Add tests that verify default categories are seeded once when initialization is rerun.
- [x] 3.4 Add tests for representative relationship inserts: transfer transaction, transaction tags, budget categories, and budget tags.
- [x] 3.5 Run formatting, tests, and linting using the repository's documented quality commands.
