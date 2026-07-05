# Changelog

All notable changes to this project will be documented in this file.

The format is inspired by [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
