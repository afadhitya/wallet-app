# Changelog

All notable changes to this project will be documented in this file.

The format is inspired by [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v1.4.0] - 2026-07-18

### Added

- **`--all-categories` flag for `budget set` and `budget edit`** — New flag to create or edit a budget that applies to all categories. When enabled, spending is calculated across all categories regardless of category-level scope.
- **Structured logging with `log/slog`** — Replaced ad-hoc logging with Go's structured `log/slog` package. Added `-v`/`--verbose` and `--log-file` flags for configurable log levels and JSON file output.

### Fixed

- **Past-due date advancement for planned payments** — Fixed `computeInitialDueDate` to advance the due date to the next valid period when it falls before the start date. Fixed `monthsBetween` day comparison for accurate month intervals.
- **Budget spending deduplication** — Fixed budget spending calculation to avoid double-counting when a transaction matches both a category and a tag within the same budget.

## [v1.3.0] - 2026-07-12

### Added

- **Self-update commands** — `wallet version` displays current version with optional `--check` flag to query the latest GitHub release. `wallet update` downloads, verifies SHA256 checksums, and atomically replaces the binary.
- **Destructive operation confirmation** — AI agents must now confirm before running destructive commands (delete, archive, adjust) with a summary of affected resources.
- **Removed planned_payment_id foreign key** — Simplified the transactions schema by dropping the `planned_payment_id` FK constraint and `is_planned` column. Planned payments now create regular transactions instead.
- **BYDAY RRULE support** — Custom WEEKLY RRULEs with `BYDAY` (e.g. `FREQ=WEEKLY;BYDAY=TU,WE`) now advance to the next matching weekday instead of blindly adding 7 days.

### Fixed

- **Multi-currency forecast** — Fixed account balance and planned payment amount conversion to base currency in `ForecastBalance`. Payments with unconfigured exchange rates are skipped with a warning.
- **CSV export JSON output** — Fixed `report export --json` printing output before the CSV file was written.

## [v1.2.0] - 2026-07-06

### Added

- **Auto-generated CLI docs** — CLI documentation is now auto-generated for all commands and subcommands.
- **AI agent documentation split** — AGENTS.md and CONTRIBUTING.md now have clear separation of concerns for AI agents working on the codebase.
- **CLI separator rule** — Added formatting rule for CLI help text separators.
- **DB access boundary rule** — Documented the data access boundary constraint for the service layer.
- **Skill installation docs** — Added documentation for installing and configuring OpenSpec skills.

### Changed

- Updated AGENTS.md and CONTRIBUTING.md with refined agent guidelines.

### Fixed

- Fixed in-memory rate config setup in service tests to use proper test configuration.

## [v1.1.0] - 2026-07-05

### Added

- **Account List Table** — New `converted_balance` column in account list showing balances converted to the base currency using configured exchange rates.
- **Multi-Currency Total** — Account list now computes a total row across all accounts, converted to the base currency for a unified financial snapshot.

## [v1.0.0] - 2026-07-05

### Added

- **Accounts** — Multi-account management with balances. Create, list, edit, and archive accounts with currency and sort order. Account types: checking, savings, cash, credit card, and e-wallet.
- **Transactions** — Record expenses, income, transfers between accounts, and balance adjustments. Filter by month, category, type, tags, and account. Soft-delete with archive recovery.
- **Categories** — 32 system-seeded categories organized in a 2-level parent-child hierarchy. Add custom categories, rename, and archive unused ones.
- **Tags** — Freeform labels with unique names. Many-to-many relationship with transactions for flexible organization.
- **Budgets** — Recurring (daily, weekly, monthly, yearly) and one-time budgets with configurable notification thresholds. Scope budgets to specific categories and tags. Check spending against active budgets.
- **Planned Payments** — Recurring bills with RRULE support (daily, weekly, monthly, yearly, custom cron patterns). Lifecycle states: active, paused, paid, skipped. Track due dates, auto-create expenses on payment.
- **Forecasting** — Project future account balances based on recurring income, bills, and spending patterns. Upcoming bill schedule with multi-month horizon.
- **Reports** — Monthly financial summaries with breakdowns by category, account, and tags. Export reports to CSV or structured JSON.
- **Multi-currency** — Configure exchange rates for currency conversion. Manage rates via CLI with list, add, set, and remove commands. All amounts stored in integer minor units.
- **AI-native CLI** — `--json` global flag for structured machine-readable JSON output. Stable error codes for programmatic consumption. Cobra-based command structure with comprehensive help text.
- **Configuration** — TOML-based config for database path, display preferences, currency defaults, and AI output mode. Works without configuration using sensible defaults.
- **Infrastructure** — SQLite database with WAL journal mode and embedded migrations. 100% test coverage enforcement. golangci-lint integration. sqlc-generated type-safe data access layer.

### Changed

- Initial public release.
