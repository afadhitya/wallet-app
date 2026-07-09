# Fix Priorities — Wallet App E2E Test Results

> **Source:** 119 test scenarios across Part A (J1-J4) and Part B (D1-D11).
> **Results:** 75 ✅ PASSED · 2 ⚠️ PARTIAL · 2 ⏭️ SKIPPED · 42 ❌ FAILED
> **This document:** Classifies the 42 failures by severity and recommends fix order.

---

## 🔴 P1 — Critical Bugs (Fix Immediately)

Data integrity or core functionality issues that can corrupt or lose data.

| # | Part | Test ID | Issue | Suggested Fix |
|---|------|---------|-------|---------------|
| 1 | B | **D4.5** | System categories (Food & Dining, etc.) can be **deleted** without error. `is_system=1` flag exists but is never enforced. | Add validation in `service/category.go` to block deletion when `is_system=1`. Return `VALIDATION_ERROR`. |
| 2 | B | **D4.6** | Categories with active transactions can be **deleted** without error. No FK constraint check. | Add `has_transactions` check before deletion. Block or warn when transactions reference the category. |
| 3 | A | **J2.19** | `wallet bill rm <id>` crashes with raw SQL FK error when bill has been paid: `"FOREIGN KEY constraint failed (787)"`. | Change `rm` to soft-delete (`is_active=0`) like other resources, or handle FK gracefully. |
| 4 | A | **J2.16** | One-time bill is archived (`is_active=0`) instead of hard-deleted after payment. `PlannedPayment` record remains. | Either implement true hard-delete for one-time bills, or update expected behavior to match soft-delete pattern. |

---

## 🟠 P2 — Functional Bugs (High Priority)

Core features don't work as documented/expected.

| # | Part | Test ID | Issue | Suggested Fix |
|---|------|---------|-------|---------------|
| 5 | B | **D3.18** | Adjustment transactions appear in default `wallet list`. AGENTS.md says adjustments are "Excluded from all reports". | Add `WHERE is_adjustment=0` filter to the default list query. Show adjustments only with `--type adjustment`. |
| 6 | A | **J1.12** | `--archived` flag doesn't exist on `wallet list` or `wallet account list`. CLI uses `--all` instead. D2.5 same issue for accounts. | Add `--archived` as an alias for `--all` on list commands, or update docs. |
| 7 | A | **J4.5** | `wallet report --export csv --output <file>` returns JSON with `file_path` claim but **no file is written** to disk. | Implement actual file writing in `--output` path, or remove `file_path` from JSON response. |
| 8 | B | **D3.2** | `wallet add expense -1000` fails with Cobra flag parsing error (`-1` as unknown flag). | Accept negative amounts as positional args after `--`. Document the required syntax for negative values. |
| 9 | B | **D10.2** | Some error paths (Cobra arg validation) don't produce JSON output even with `--json` flag. | Catch arg validation errors and wrap them in the standard JSON error envelope. |

---

## 🟡 P3 — Multi-Currency Gaps (Medium Priority)

Multi-currency conversion works at transaction creation and in reports, but not in derived features.

| # | Part | Test ID | Issue | Suggested Fix |
|---|------|---------|-------|---------------|
| 10 | A | **J3.10 / D6.8** | Budget `check` uses raw amounts (50000 IDR + 10 USD = 50010) instead of converted (50000 + 160000 = 210000 IDR). | Use `COALESCE(base_amount, amount)` when summing budget spending. |
| 11 | A | **J3.11-12 / D9.4-5** | Forecast and forecast bills don't convert bill amounts. Hosting 10 USD treated as 10 IDR. | Convert bill amounts using account's exchange rate before forecast totals. |
| 12 | A | **J3.7 / D2.8** | `wallet account list` shows only individual raw balances. No converted total or aggregate net worth. | Add a summary row with converted total (using stored rates) at the bottom of account list. |
| 13 | B | **D3.14** | `wallet list --category <parent>` shows 0 results. Parent filter doesn't include child categories. | Expand category filter to include all descendants when a parent is specified. |
| 14 | A | **J3.8 / D2.9 / D9.6** | No warnings when exchange rates are missing. Built-in defaults mask rate removal. | Compare against config file (not built-in defaults) to detect missing rates. |

---

## 🔵 P4 — CLI ID vs Name Inconsistencies (Medium Priority)

Some commands accept names, others require numeric IDs. Creates confusing UX.

| # | Part | Test ID | Issue | Suggested Fix |
|---|------|---------|-------|---------------|
| 15 | B | **D2.3 / D2.4 / D2.6** | `wallet account edit/archive` expects `<id>` (integer) not name. User gets `INVALID_INPUT: invalid account ID: Checking`. | Add name-based lookup, or show available names on error. |
| 16 | B | **D4.3 / D4.4** | `wallet category edit/rm` expects `<id>` not name. | Add name-based lookup, or show available names on error. |
| 17 | B | **D6.3 / D6.10** | `wallet budget edit/rm` expects `<id>` not name. | Add name-based lookup, or show available names on error. |
| 18 | B | **D4.2** | `wallet category add --parent` expects numeric ID (`--parent 8`) not name (`--parent Income`). | Implement name-based parent lookup. |

---

## ⚪ P5 — Test Scenario Flag Mismatches (Low Priority)

The CLI works correctly; the scenarios use wrong flags or category names. Fix the test document.

| # | Part | Test IDs | Issue | Fix |
|---|------|----------|-------|-----|
| 19 | A+B | **J1.5, J2.1, D1.3, D3.1a, D3.6, D3.14, D6.1** | `-c Food` / `-c Transport` — wrong category name. Should be "Food & Dining" / "Transportation". | Update test scenarios |
| 20 | A+B | **J2.8, J2.15, D7.1, D7.2, D7.6** | `--recurrence monthly` / `--due-date` — flags don't exist in `bill add`. Should be `--monthly --day N -c <cat>`. | Update test scenarios |
| 21 | A+B | **J2.3, D6.5, D6.6** | `wallet budget check` requires `--all` or `-b <id>`. Bare command returns error. | Add `--all` to scenarios |
| 22 | A | **J2.10** | `wallet bill due --overdue` — bill not overdue yet. Should use bare `wallet bill due`. | Fix flag or precondition in scenario |
| 23 | A | **J4.6** | `wallet forecast` defaults to 1 month. Expected 6. Should use `-n 6`. | Add `-n 6` to scenario |
| 24 | A | **J1.2, D1.2** | `wallet init` on existing DB returns success, not error. App is idempotent by design. | Update expected result |
| 25 | A | **J3.1-2, D8.1-2** | App has built-in default rates. "No rates" precondition impossible. | Update expected results |
| 26 | B | **D1.4** | App auto-recreates missing DB. No error possible. | Update expected result |
| 27 | B | **D1.1** | `config.toml` not created during init. Only `rates.toml` is. | Either implement config.toml creation or update expected result |
| 28 | B | **D7.7** | Missing `-c` (category) flag validates before account. | Add `-c <category>` to scenario |

---

## ℹ️ Skipped Tests (Not Failures)

These 2 scenarios cannot be tested via CLI in a single session — they require time passage.

| ID | Reason |
|----|--------|
| **J2.6** | Requires waiting until next month to test one_time budget period check |
| **J2.7** | Requires monthly period rollover to test auto-creation of new period |

---

## Summary by Priority

| Priority | Count | Part A | Part B | Key Problems |
|----------|-------|--------|--------|--------------|
| 🔴 P1 | 4 | 2 | 2 | System cat delete, used cat delete, FK crash on bill rm, one-time bill archive |
| 🟠 P2 | 5 | 2 | 3 | Adjustments in list, --archived missing, CSV export broken, negative amounts, JSON envelope |
| 🟡 P3 | 5 | 4 | 1 | Multi-currency gaps (budget, forecast, account total, category filter, warnings) |
| 🔵 P4 | 4 | 0 | 4 | CLI ID vs name (account, category, budget, parent) |
| ⚪ P5 | 10 | 5 | 5 | Test scenario flag mismatches |
| ⏭️ Skipped | 2 | 2 | 0 | Time-constrained tests |

### Recommended Fix Order

1. **🔴 P1** — Data integrity first
2. **🟠 P2** — Core UX broken in visible ways
3. **🔵 P4** — Name-based lookup (low effort, high UX impact)
4. **🟡 P3** — Multi-currency completeness
5. **⚪ P5** — Update test scenarios to match actual CLI behavior

---

*Generated from e2e-testing-scenarios.md — 119 test cases evaluated on 2026-07-09.*
