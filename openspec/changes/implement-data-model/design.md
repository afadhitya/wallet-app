## Overview

Implement the approved wallet data model as the initial SQLite schema for the Go application. The schema will create the core tables for accounts, categories, tags, transactions, budgets, planned payments, exchange rates, and many-to-many junctions, plus seed default categories.

## Goals

- Provide a deterministic SQLite schema that can initialize a new wallet database.
- Preserve the approved modeling decisions from `brainstorming/01-data-model.md`.
- Keep amounts in integer minor units to avoid floating-point money errors.
- Support future features without requiring immediate CLI CRUD implementation.

## Non-Goals

- Build full account, transaction, budget, or reporting commands.
- Implement balance mutation business logic beyond schema support.
- Implement exchange-rate fetching integrations.
- Implement recurring payment fulfillment logic.

## Data Model

The initial schema will include these tables:

- `accounts`: money storage entities with type, currency, balance, archive flag, sort order, and timestamps.
- `categories`: two-level classification with optional parent, type, icon, color, system flag, and sort order.
- `tags`: unique freeform labels with optional color.
- `transactions`: canonical money movements, including expense, income, transfer, and adjustment types.
- `transaction_tags`: many-to-many relation between transactions and tags.
- `budgets`: per-period spending limits stored as snapshots.
- `budget_categories`: many-to-many relation between budgets and categories.
- `budget_tags`: many-to-many relation between budgets and tags.
- `planned_payments`: scheduled income or expense records with precomputed `next_due_date`.
- `exchange_rates`: cached conversion rates between currency pairs.

Default category seed data will include the approved parent and child categories for food, transportation, shopping, bills, entertainment, health, education, and income.

## Key Decisions

- Transfers are represented as a single `transactions` row with `type = 'transfer'` and `transfer_to_id` pointing at the destination account.
- Balance adjustments use `type = 'adjustment'` and are excluded from future income and expense reports by type.
- Budgets are period snapshots; recurring budgets will be cloned into future periods rather than mutating past periods.
- Budget targets are many-to-many category and tag relationships, with the application enforcing at least one target.
- Categories are limited to a parent-child hierarchy, avoiding recursive category trees.
- Multi-currency transactions store original `amount` and `currency`, plus optional `base_amount` and `base_currency` in the account currency.
- `planned_payments.next_due_date` is stored and updated during fulfillment instead of computed at query time.

## Implementation Approach

Add database initialization code that executes the schema in a stable order with foreign keys enabled. Seed data should be idempotent so initialization can run repeatedly without duplicating system categories; this may require uniqueness constraints or insert guards for system category names within their parent scope.

Tests should initialize an empty SQLite database, verify the schema exists, verify foreign key enforcement is enabled, and verify default categories are present. The test suite should also cover representative inserts for each major relationship, including a transfer transaction, transaction tags, and budget targets.

## Risks and Tradeoffs

- SQLite cannot enforce every domain rule cleanly, such as requiring each budget to have at least one category or tag target; these rules will be enforced in application code.
- Denormalized account balances require careful write-path logic in later changes, but they keep balance reads simple and fast.
- Category seed idempotency needs a stable uniqueness strategy to avoid duplicate rows when initialization is rerun.
