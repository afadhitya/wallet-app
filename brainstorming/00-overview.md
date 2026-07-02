# Wallet App — Brainstorming Index

> Branch: `brainstorming` | Date: 2026-07-02
> Approach: **Bottom-up, phase-by-phase** — each phase builds on the previous one.
> Scope: Single-user, SQLite, CLI-first, AI-native, open source.

---

## Phases

| # | Phase | Focus | Status |
|---|-------|-------|--------|
| 01 | Data Model | SQLite schema, entities, relationships, currencies | ✅ design approved |
| 02 | Project Skeleton | CLI framework, project structure, tooling, config | ⬜ blocked |
| 03 | Core CRUD | Transactions (expense/income/transfer), categories, tags | ⬜ blocked |
| 04 | Budget Engine | Monthly budgets, alerts, rollover, progress tracking | ⬜ blocked |
| 05 | Planned Payments | Recurring + one-time payments, due dates, cash flow impact | ⬜ blocked |
| 06 | Forecasting | Projected balances, trend analysis, bill calendar | ⬜ blocked |
| 07 | Multi-Currency | Exchange rates, conversion, mixed-currency reporting | ⬜ blocked |
| 08 | AI-Native Layer | Hermes skill, JSON output mode, NLP-friendly CLI | ⬜ blocked |
| 09 | Reports & TUI | Terminal dashboard, charts, export | ⬜ blocked |

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
| Tags | **Both** — categories + freeform tags; budget by category OR tag |
| Transfer model | **Single row** with transfer_to_id |
| Budget periods | **Snapshot per period** (clone) |
| Forecasting approach | TBD |
| Multi-currency strategy | TBD |

---

## Reference Docs

- [Research: Wallet by BudgetBakers](./research-wallet-by-budgetbakers.md)
