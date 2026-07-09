# Fix Priorities — Wallet App E2E Test Results

> **Source:** 119 test scenarios across Part A (J1-J4) and Part B (D1-D11).
> **Results:** 75 ✅ PASSED · 2 ⚠️ PARTIAL · 2 ⏭️ SKIPPED · 42 ❌ FAILED
> **This document:** Classifies the 42 failures by severity and recommends fix order.

---

## 🔴 P1 — Critical Bugs (Fix Immediately)

Data integrity or core functionality issues that can corrupt or lose data.

| # | Test ID | Issue | Impact | Suggested Fix |
|---|---------|-------|--------|---------------|
| 1 | **D4.5** | System categories (Food & Dining, etc.) can be **deleted** without error. `is_system=1` flag exists but is never enforced. | User can permanently delete base categories. All child categories become orphaned. | Add validation in `service/category.go` to block deletion when `is_system=1`. Return `VALIDATION_ERROR`. |
| 2 | **D4.6** | Categories with active transactions can be **deleted** without error. No FK constraint check. | Referential integrity broken. Transactions reference non-existent categories. | Add `has_transactions` check before deletion. Block or warn when transactions reference the category. |
| 3 | **J2.19** | `wallet bill rm <id>` crashes with raw SQL FK error when bill has been paid: `"FOREIGN KEY constraint failed (787)"`. | Exposes internal SQL error to user. Bill cannot be removed if it has payments. | Change `rm` to soft-delete (`is_active=0`) like other resources, or handle FK gracefully with a user-friendly message. |

---

## 🟠 P2 — Functional Bugs (High Priority)

Core features don't work as documented/expected.

| # | Test ID | Issue | Impact | Suggested Fix |
|---|---------|-------|--------|---------------|
| 4 | **D3.18** | Adjustment transactions appear in default `wallet list`. AGENTS.md says adjustments are "Excluded from all reports". | Users see balance corrections mixed with real transactions. Confusing. | Add `WHERE is_adjustment=0` filter to the default list query. Only show adjustments with `--type adjustment`. |
| 5 | **J1.12 / D2.5** | `--archived` flag doesn't exist on `wallet list` or `wallet account list`. CLI uses `--all` instead. | Users can't view archived items via the documented flag name. | Add `--archived` as an alias for `--all` on list commands, or update docs. |
| 6 | **J4.5** | `wallet report --export csv --output <file>` returns JSON with `file_path` claim but **no file is written** to disk. | CSV export is non-functional. Users get JSON instead of a real CSV file. | Implement actual file writing in `--output` path, or remove `file_path` from JSON response. |
| 7 | **D3.2** | `wallet add expense -1000` fails with Cobra flag parsing error (`-1` as unknown flag). | Users cannot enter negative amounts naturally. Requires workaround (`--` separator). | Accept negative amounts as positional args after `--`. Document the required syntax for negative values. |
| 8 | **D10.2** | Some error paths (Cobra arg validation) don't produce JSON output even with `--json` flag. | Inconsistent error handling — machine consumers can't parse all errors. | Catch arg validation errors and wrap them in the standard JSON error envelope. |

---

## 🟡 P3 — Multi-Currency Gaps (Medium Priority)

Multi-currency support is partially implemented — conversion works at transaction creation and in reports, but not in derived features.

| # | Test ID | Issue | Impact | Suggested Fix |
|---|---------|-------|--------|---------------|
| 9 | **J3.10 / D6.8** | Budget `check` uses raw amounts for multi-currency spending (50000 IDR + 10 USD = 50010) instead of converted values (50000 + 160000 = 210000 IDR). | Budget tracking is wrong for users with foreign currency accounts. | Use `COALESCE(base_amount, amount)` when summing budget spending, falling back to raw amount only when `base_amount` is null. |
| 10 | **J3.11-12 / D9.4-5** | Forecast and forecast bills don't convert bill amounts. Hosting 10 USD is treated as 10 IDR. | Forecast projections are incorrect for multi-currency users. | Convert bill amounts using account's configured exchange rate before adding to forecast totals. |
| 11 | **J3.7 / D2.8** | `wallet account list` shows only individual raw balances. No converted total or aggregate net worth. | Users with multiple currencies can't see their total net worth in one place. | Add a summary row with converted total (using stored rates) at the bottom of account list. |
| 12 | **D3.14** | `wallet list --category <parent>` shows 0 results. Parent category filter doesn't include child categories. | Users filtering by e.g. "Food & Dining" miss transactions in Restaurant, Groceries, etc. | Expand category filter to include all descendants when a parent is specified. |
| 13 | **J3.8 / D2.9 / D9.6** | No warnings when exchange rates are missing. Built-in defaults mask rate removal. | Users don't know when their configured rates are removed or fallback rates are used. | Compare against config file (not built-in defaults) to detect missing rates. Surface warnings in account list, forecast, and budget check. |

---

## 🔵 P4 — CLI Flag Inconsistencies (Medium Priority)

CLI commands use different accessor patterns — some by name, others by numeric ID — creating a confusing UX.

| # | Test ID | Issue | Impact | Suggested Fix |
|---|---------|-------|--------|---------------|
| 14 | **D2.3 / D2.4 / D2.6** | `wallet account edit/archive` expects `<id>` (integer) but users naturally use account names. | Users get `INVALID_INPUT: invalid account ID: Checking` instead of a helpful error. | Add name-based lookup to `account edit` and `account archive`, or add a clear error with available names. |
| 15 | **D4.3 / D4.4** | `wallet category edit/rm` expects `<id>` but categories have names. Same as accounts. | Users can't use category names directly for edit/delete operations. | Add name-based lookup or suggest correct numeric ID with category list on error. |
| 16 | **D6.3 / D6.10** | `wallet budget edit/rm` expects `<id>` but budgets have names. | Users expect to reference budgets by name. | Same fix: add name-based lookup or improve error message. |
| 17 | **D4.2** | `wallet category add --parent` expects numeric ID (e.g. `--parent 8`) not category name (`--parent Income`). | Users get `VALIDATION_ERROR: invalid parent category ID`. | Implement name-based parent lookup: search categories by name when `--parent` receives a non-numeric value. |

---

## ⚪ P5 — Flag Mismatches in Test Scenarios (Low Priority)

These failures are in the test document, not in the app. The CLI works correctly; the scenarios use wrong flags or category names. They should be corrected in the test document.

| # | Test IDs | Issue | Fix |
|---|----------|-------|-----|
| 18 | **J1.5, J2.1, D1.3, D3.1a, D3.6, D3.14, D6.1** | `-c Food` / `-c Transport` — wrong category name. Should be "Food & Dining" / "Transportation". | Update test scenarios |
| 19 | **J2.8, J2.15, D7.1, D7.2, D7.6** | `--recurrence monthly` / `--due-date` — flags don't exist in `bill add`. Should be `--monthly --day N -c <cat>`. | Update test scenarios |
| 20 | **J2.3, D6.5, D6.6** | `wallet budget check` requires `--all` or `-b <id>`. Bare command returns error. | Add `--all` to scenarios |
| 21 | **J2.10** | `wallet bill due --overdue` — bill not overdue yet. Should use bare `wallet bill due`. | Fix flag or precondition |
| 22 | **J4.6** | `wallet forecast` defaults to 1 month. Expected 6. Should use `-n 6`. | Add `-n 6` to scenario |
| 23 | **J1.2, D1.2** | `wallet init` on existing DB returns success, not error. App is idempotent. | Update expected result to reflect idempotent behavior |
| 24 | **J3.1-2, D8.1-2** | App has built-in default rates. "No rates" precondition impossible. | Update expected results |
| 25 | **D1.4** | App auto-recreates missing DB. No error possible. | Update expected result |
| 26 | **D1.1** | `config.toml` not created during init. Only `rates.toml` is. | Either implement config.toml creation or update expected result |

---

## Summary by Priority

| Priority | Count | Key Problems |
|----------|-------|--------------|
| 🔴 P1 | 3 | Data integrity (delete system categories, delete used categories, FK crash) |
| 🟠 P2 | 5 | Adjustments in list, --archived missing, CSV export broken, negative amounts, JSON error envelope |
| 🟡 P3 | 5 | Multi-currency gaps (budget, forecast, account total, category filter, warnings) |
| 🔵 P4 | 4 | CLI ID vs name inconsistencies (account, category, budget, parent) |
| ⚪ P5 | 8 | Test document flag mismatches and erroneous expected results |

### Recommended Fix Order

1. **P1 items 1-3** — Data integrity is non-negotiable
2. **P2 items 4-8** — Core UX broken in visible ways
3. **P4 items 14-17** — Low effort, high UX impact (name-based lookup)
4. **P3 items 9-13** — Multi-currency affects a smaller user base but is important for completeness
5. **P5 items 18-26** — Update test scenarios to match actual CLI behavior

---

*Generated from e2e-testing-scenarios.md — 119 test cases evaluated on 2026-07-09.*
