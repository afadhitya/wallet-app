# Wallet App — Brainstorming Index

> Branch: `brainstorming` | Date: 2026-07-02 | Approved: 2026-07-02
> Approach: **Bottom-up, phase-by-phase** — each phase builds on the previous one.
> Scope: Single-user, SQLite, CLI-first, AI-native, open source.

---

## Phases

| # | Phase | Focus | Status |
|---|-------|-------|--------|
| 01 | Data Model | SQLite schema, entities, relationships, currencies | ✅ design approved |
| 02 | Project Skeleton | CLI framework, project structure, tooling, config | ✅ design approved |
| 03 | Core CRUD | Transactions (expense/income/transfer/adjustment), categories, tags | ✅ design approved |
| 04 | Budget Engine | Monthly budgets, alerts, progress tracking | ✅ design approved |
| 05 | Planned Payments | Recurring + one-time payments, due dates, cash flow impact | ✅ design approved |
| 06 | Forecasting | Projected balances (planned payments only), per-category breakdown | ✅ design approved |
| 07 | Multi-Currency | Exchange rates, conversion, mixed-currency reporting | ✅ design approved |
| 08 | AI-Native Layer | Hermes skill, JSON output mode, all commands wrapped | ✅ design approved |
| 09 | Reports & Export | Reports by category/account/tag, CSV export | ✅ design approved |
| 10 | Documentation | README, LICENSE (MIT), CONTRIBUTING, AGENTS.md | ✅ design approved |
| 11 | README Badges | Go Version, CI, Coverage, License, Release badges | 🔴 pending review |

---

## Guiding Principles

1. **Bottom-up** — Start from the database. No UI until data model is solid.
2. **Single responsibility** — Each phase produces a working, testable artifact.
3. **CLI-first** — Every feature is exposed as a CLI command. Hermes skill wraps CLI.
4. **SQLite only** — No external services. DB file = application state.
5. **English artifacts** — Code, comments, docs, CLI output, commit messages — all English.

---

## Decisions Log

| Decision | Answer |
|----------|--------|
| Language | **Go** |
| Interaction model | **CLI-first** (TUI optional) |
| Category model | **2-level hierarchy** (parent→child) |
| Tags | **Both** — categories + freeform tags; budget by category AND/OR tag |
| Transfer model | **Single row** with transfer_to_id |
| Budget periods | **Snapshot per period** (clone) |
| Forecasting approach | **Planned payments only** — per-category breakdown, no tag breakdown |
| Multi-currency strategy | **Manual rate source** (config file), convert at transaction time |
| Balance adjustment | **New tx type `adjustment`** — not income/expense, tracked, excluded from reports |
| Delete behavior | **Soft delete** (is_archived) — preserve history |
| License | **MIT** |
| README style | **Standard OSS** |
| Badge count | **5** — Go Version, CI, Coverage, License, Release |
| Badge style | **Flat** |

---

## Reference Docs

- [Research: Wallet by BudgetBakers](./research-wallet-by-budgetbakers.md)
