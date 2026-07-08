# Wallet App — E2E Testing Scenarios

> **Purpose:** Define end-to-end testing scenarios for manual execution (via agent or human tester).
> **Scope:** All features — accounts, transactions, categories, tags, budgets, bills, multi-currency, reports, forecasts, init/config, JSON output, self-update.
> **Format:** Each scenario is a table row with ID, preconditions, action, expected result, test result, and reason/suggestion.
> **Note:** Test steps are not included — only scenario definitions.

---

## Part A: User Journeys

End-to-end workflows that simulate real user behavior across multiple features.

### J1: Fresh Start — Init, Accounts, Transactions

| ID | Preconditions | Scenario | Expected Result | Result | Reason / Suggestion |
|----|---------------|----------|-----------------|--------|---------------------|
| J1.1 | `~/.config/wallet/` does not exist | Run `wallet init` | DB created at default path, config file created, 32 categories seeded | ✅ PASSED | DB created at default path, 32 categories seeded as expected. |
| J1.2 | Wallet is initialized | Run `wallet init` again | Error: already initialized (non-destructive) | ❌ FAILED | CLI returns `success: true` with message "Wallet database initialized successfully" instead of an error. **Suggestion:** Add re-init detection to return an error/warning when the DB already exists. |
| J1.3 | Wallet initialized | `wallet account add "Checking" --type checking --currency IDR` | Account created with 0 balance | ✅ PASSED | Checking account created with balance 0 as expected. |
| J1.4 | Wallet initialized | `wallet account add "Savings" --type savings --currency IDR` | Account created | ✅ PASSED | Savings account created successfully. |
| J1.5 | Checking account exists | `wallet add expense 50000 "Lunch" -c Food -a Checking` | Transaction created, Checking balance = -50000 | ❌ FAILED | Category "Food" does not exist in seeded categories (only "Food & Dining"). CLI returns `INTERNAL_ERROR` with suggestion "Did you mean: [Food & Dining]?". Using `-c "Food & Dining"` passes. **Suggestion:** Fix test scenario to reference "Food & Dining" or add "Food" as a category alias. |
| J1.6 | Accounts exist | `wallet add income 2000000 "Salary" -c Income -a Checking` | Income created, Checking balance = 1950000 | ✅ PASSED | Income of 2,000,000 created correctly under Checking via Income category. |
| J1.7 | Two accounts exist | `wallet add transfer 500000 --from Checking --to Savings` | Transfer created, Checking -500000, Savings +500000 | ✅ PASSED | Transfer of 500,000 from Checking to Savings created successfully with correct balance changes. |
| J1.8 | Transactions exist | `wallet list` | All transactions listed with correct amounts, descriptions, categories | ✅ PASSED | All transactions listed with correct amounts, descriptions, and categories. |
| J1.9 | Transactions exist | `wallet list --month 2026-07` | Only July transactions shown | ✅ PASSED | Month filter works correctly, shows only July transactions. |
| J1.10 | Transactions exist | `wallet edit 1 --amount 55000 --desc "Lunch upgraded"` | Transaction updated, reflected in list | ✅ PASSED | `wallet edit` command updates amount and description correctly. Note: since J1.5 failed, the expense was not created, so J1.10 edited the next available transaction instead. The edit operation itself works as expected. |
| J1.11 | A transaction exists | `wallet rm 1` (confirm) | Transaction archived, removed from default list | ✅ PASSED | `wallet rm --force` archives transaction and removes it from default list. |
| J1.12 | An archived transaction | `wallet list --archived` | Archived transaction visible with archive flag | ❌ FAILED | `--archived` flag is not implemented in the CLI. `wallet list --help` shows no such flag. **Suggestion:** Implement `--archived` or `--include-archived` flag to display archived transactions in list output. |

### J2: Budget & Bills

| ID | Preconditions | Scenario | Expected Result | Result | Reason / Suggestion |
|----|---------------|----------|-----------------|--------|---------------------|
| J2.1 | Wallet initialized, accounts exist | `wallet budget set "Monthly Food" 1000000 -c Food --period monthly --notify 80` | Budget created with current month period | ❌ FAILED | Category "Food" does not exist (only "Food & Dining"). CLI returns `INTERNAL_ERROR` with suggestion. Using `-c "Food & Dining"` succeeds. **Suggestion:** Fix test scenario to use "Food & Dining". |
| J2.2 | Budget exists | `wallet budget list` | Budget shown with spent = 0, remaining = 1000000, status = ok | ✅ PASSED | Budget listed with spent=0, remaining=1000000. Status implied by data; no explicit `status` field in list output (status field is in `budget check` output). Functional result matches expectations. |
| J2.3 | Budget exists, no spending | `wallet budget check` | Budget status = ok, spending = 0 | ⚠️ PARTIAL | `wallet budget check` requires `--all` or `-b <id>` flag; bare command returns error. With `--all` flag, shows status="ok", spent=0 as expected. **Suggestion:** Fix test scenario to use `wallet budget check --all`. |
| J2.4 | Budget exists | Add expenses in Food category totaling 850000 | `wallet budget check` shows spent = 850000, status = warning (at 85%, above 80% notify threshold) | ✅ PASSED | After adding 850000 in expenses (Food & Dining category), `wallet budget check --all` shows spent=850000, percent_used=85, status="warning". |
| J2.5 | Budget overspent | Add another expense of 200000 in Food (total = 1050000) | `wallet budget check` shows spent = 1050000, status = over (105%) | ✅ PASSED | After adding 200000 more, `wallet budget check --all` shows spent=1050000, percent_used=105, status="over". |
| J2.6 | Budget with period = one_time | `wallet budget check` in next month | No new period created, original budget unchanged | ⏭️ SKIPPED | Requires waiting until next month. Cannot test in current session (time constraint). |
| J2.7 | Budget with period = monthly | Period rolls over | `wallet budget check` auto-creates new period for current month | ⏭️ SKIPPED | Requires period rollover. Cannot test in current session (time constraint). |
| J2.8 | Budget exists | `wallet bill add "Internet" 300000 --account Checking --recurrence monthly --due-date 2026-07-15` | Bill created with next_due = 2026-07-15 | ❌ FAILED | CLI does not have `--recurrence` or `--due-date` flags. Actual CLI uses `--monthly` (or `--daily`/`--weekly`/`--yearly`) and `--day N` flags; also requires `-c` (category) flag. With correct syntax `-a Checking -c Internet --monthly --day 15`, bill created with next_due=2026-07-15. **Suggestion:** Update test scenario to use actual CLI flags: `--monthly --day 15 -c Internet`. |
| J2.9 | Bills exist | `wallet bill list` | All bills listed with next due dates and status | ✅ PASSED | Bills listed with next due dates and status fields correctly. |
| J2.10 | Bill is due | `wallet bill due --overdue` | Internet bill shown as due | ❌ FAILED | `--overdue` only shows bills past their due date. Since next_due=2026-07-15 is in the future (today is 2026-07-08), it's not overdue. Bill appears with `wallet bill due` (no flags) or `--next 30`. **Suggestion:** Fix test scenario to use `wallet bill due` (no flags) or adjust the due date precondition. |
| J2.11 | Bill is due | `wallet bill pay 1 --amount 300000` | Expense transaction created for Internet, next_due advanced to next month | ✅ PASSED | Transaction created with id=5, amount=300000, category=Internet. next_due_date advanced to 2026-08-15. |
| J2.12 | Bill is active | `wallet bill pause 1` | Bill status = paused | ✅ PASSED | Bill paused (is_paused=1), status reflects correctly in list output. |
| J2.13 | Bill is paused | `wallet bill pay 1` | Error: bill is paused | ✅ PASSED | Correct `BILL_PAUSED` error returned with suggestion "unpause the planned payment first". |
| J2.14 | Bill is paused | `wallet bill resume 1` | Bill status = active | ✅ PASSED | Bill resumed (is_paused=0), back to active status. |
| J2.15 | Bill is active, one-time | `wallet bill add "Test One-Time" 50000 -a Checking --recurrence one_time --due-date 2026-07-20` | One-time bill created | ❌ FAILED | CLI does not have `--recurrence` or `--due-date` flags. One-time bill can be created by omitting recurrence flags (defaults to recurrence="none") and using `--from`/`--day`. With `-a Checking -c Internet --from 2026-07-20 --day 20`, a one-time bill with next_due=2026-07-20 is created. **Suggestion:** Update test scenario to use `--from 2026-07-20 --day 20` (omit recurrence flags for one-time). |
| J2.16 | One-time bill is paid | `wallet bill pay <id>` | Bill hard-deleted after payment | ❌ FAILED | Bill is archived (is_active=0), not hard-deleted. The PlannedPayment record remains with is_active=0. **Suggestion:** Update expected result to "Bill archived (is_active=0)" or implement actual hard-delete for one-time bills. |
| J2.17 | Bill is active | `wallet bill skip 1` | Next due date advanced, no transaction created | ✅ PASSED | Next due date advanced from 2026-08-15 to 2026-09-15. No transaction created. |
| J2.18 | Bill exists | `wallet bill edit 1 --amount 350000` | Bill amount updated | ✅ PASSED | Bill amount updated from 300000 to 350000 successfully. |
| J2.19 | Bill exists | `wallet bill rm 1` | Bill archived (is_active = 0) | ❌ FAILED | Returned `INTERNAL_ERROR` with raw SQL error: "FOREIGN KEY constraint failed (787)". The bill has existing payment transactions (via planned_payment_id FK), so hard-delete fails. **Suggestion:** Change `rm` to soft-delete (set is_active=0) instead of hard DELETE, or handle FK constraint gracefully. |

### J3: Multi-Currency

| ID | Preconditions | Scenario | Expected Result |
|----|---------------|----------|-----------------|
| J3.1 | Wallet initialized, IDR is base currency | `wallet rate list` | No rates configured |
| J3.2 | No rates configured | `wallet rate add USD 15800` | USD rate = 15800 (1 USD = 15800 IDR) |
| J3.3 | Rate exists | `wallet rate set USD 16000` | USD rate updated to 16000 |
| J3.4 | IDR & USD accounts exist | `wallet rate list` | Rates displayed correctly |
| J3.5 | USD account exists with 0 balance | `wallet add income 500 "Freelance" -c Income -a USD-Account` | Transaction created, base_amount = 500 × 16000 = 8000000 IDR |
| J3.6 | Transactions in multiple currencies | `wallet list` | All transactions shown with original amounts + base currency conversion |
| J3.7 | Transactions in multiple currencies | `wallet account list` | Accounts listed with converted balance total |
| J3.8 | USD rate missing | `wallet rate rm USD`, then list transactions | Graceful handling — error or warning about missing rate |
| J3.9 | Transactions in multiple currencies | `wallet report --month 2026-07` | Report totals in base currency (IDR), all conversions applied correctly |
| J3.10 | Transactions in IDR + USD in same category | Add expense 50000 IDR and 10 USD (rate=16000) in Food category | Budget check shows spent in base currency (50000 + 160000 = 210000 IDR) |
| J3.11 | Bills in multiple currencies | Create bill "Hosting" 10 USD monthly with USD account, and "Internet" 300000 IDR monthly with IDR account | `wallet forecast bills` shows all amounts converted to base currency, running total and total_amount in base currency |
| J3.12 | Multi-currency forecast | Accounts with balances in IDR and USD + bills in both currencies | `wallet forecast` shows all projected balances in base currency, totals correctly converted |

### J4: Period End — Reports & Forecast

| ID | Preconditions | Scenario | Expected Result |
|----|---------------|----------|-----------------|
| J4.1 | Transactions across multiple months/categories | `wallet report --month 2026-07` | Report shows total income, expenses, net, transfers — all in base currency |
| J4.2 | Multiple categories with transactions | `wallet report --month 2026-07 --by category` | Breakdown by category hierarchy (parent → child), correct subtotals |
| J4.3 | Multiple accounts with transactions | `wallet report --month 2026-07 --by account` | Breakdown by account, correct per-account totals |
| J4.4 | Transactions with tags | `wallet report --month 2026-07 --by tag` | Breakdown by tag, correct totals per tag |
| J4.5 | Report data exists | `wallet report --month 2026-07 --export csv --output /tmp/report.csv` | CSV file created with correct headers and data matching the text report |
| J4.6 | Transactions exist, base currency configured | `wallet forecast` | Shows projected balance for next 6 months based on past patterns |
| J4.7 | Bills exist with due dates | `wallet forecast bills` | Shows upcoming bills with dates and running total |
| J4.8 | Invalid month specified | `wallet report --month invalid` | Error: invalid month format |
| J4.9 | No transactions in month | `wallet report --month 2025-01` | Report shows all zeros (no income/expense) |
| J4.10 | Multi-currency transactions across accounts | Expenses in IDR account + USD account in same month | `wallet report --month 2026-07 --by account` shows each account income/expense in base currency |
| J4.11 | Multi-currency expense in budget category | Expense 50000 IDR + 10 USD (rate=16000) in Food category same month | `wallet report --month 2026-07` expense total = 50000 + 160000 = 210000 IDR |

---

## Part B: Domain-Specific Scenarios

Isolated feature tests covering edge cases and error paths.

### D1: Init & Config

| ID | Preconditions | Scenario | Expected Result |
|----|---------------|----------|-----------------|
| D1.1 | No config exists | `wallet init` | DB + config created, categories seeded |
| D1.2 | Already initialized | `wallet init` | Error, no destructive action |
| D1.3 | Config exists | `wallet add expense 10000 "test" -c Food -a Unknown` | Error: account not found |
| D1.4 | DB file deleted manually | Run any command | Error: unable to open database |

### D2: Accounts

| ID | Preconditions | Scenario | Expected Result |
|----|---------------|----------|-----------------|
| D2.1 | Wallet initialized | `wallet account add "Credit Card" --type credit --currency IDR` | Account created with type=credit, balance=0 |
| D2.2 | Account exists | `wallet account list` | Account listed with name, type, currency, balance, sort_order |
| D2.3 | Account exists | `wallet account edit "Checking" --name "Main Checking"` | Account name updated |
| D2.4 | Account has transactions | `wallet account archive "Main Checking"` | Account archived, not shown in default list |
| D2.5 | Archived account | `wallet account list --archived` | Archived accounts shown |
| D2.6 | Nonexistent account | `wallet account edit "Fake" --name "X"` | Error: account not found |
| D2.7 | Duplicate name | Create account with name that already exists | Error: account name already exists |
| D2.8 | Multiple accounts with different currencies | IDR account (balance 1000000) + USD account (balance 500, rate=16000) | Account list shows total converted = 1000000 + (500×16000) = 9000000 IDR |
| D2.9 | Multi-currency with missing rate | USD account exists, no rate configured | Account list shows USD balance raw, total excludes USD, warning about missing rate |

### D3: Transactions

| ID | Preconditions | Scenario | Expected Result |
|----|---------------|----------|-----------------|
| D3.1 | Account exists | `wallet add expense 0 "zero" -c Food -a Checking` | Error: invalid amount (must be > 0) |
| D3.1a | Account exists in non-base currency | Add expense 10 USD -c Food -a USD-Account | Transaction created with amount=10, currency=USD, base_amount=160000 (at rate 16000), base_currency=IDR |
| D3.2 | Account exists | `wallet add expense -1000 "negative" -c Food -a Checking` | Error: invalid amount |
| D3.3 | Account exists | `wallet add income abc "invalid" -c Income -a Checking` | Error: invalid amount |
| D3.4 | Category required for expense | `wallet add expense 10000 "no-cat" -a Checking` | Error: category required |
| D3.5 | Account exists | `wallet add expense 10000 "test" -c NonExistent -a Checking` | Error: category not found (with suggestions) |
| D3.6 | Expense exists | `wallet edit 1 --category Food` | Category updated |
| D3.7 | Transaction exists | `wallet edit 1 --add-tag "urgent" --add-tag "work"` | Tags added to transaction |
| D3.8 | Transaction has tags | `wallet edit 1 --remove-tag "urgent"` | Tag removed from transaction |
| D3.9 | Invalid transaction ID | `wallet edit 999` | Error: transaction not found |
| D3.10 | Transfer without `--from` | `wallet add transfer 100000 --to Savings` | Error: --from required |
| D3.11 | Transfer without `--to` | `wallet add transfer 100000 --from Checking` | Error: --to required |
| D3.12 | Adjustment | `wallet adjust Checking 1000000 "Correction"` | Adjustment transaction created, Checking balance = 1000000 |
| D3.13 | Filter with `--account` | `wallet list --account Checking` | Only Checking account transactions shown |
| D3.14 | Filter with `--category` | `wallet list --category Food` | Only Food category transactions shown |
| D3.15 | Filter with `--tag` | `wallet list --tag urgent` | Only transactions with tag "urgent" shown |
| D3.16 | Filter with `--from`/`--to` | `wallet list --from 2026-07-01 --to 2026-07-31` | Only July transactions shown |
| D3.17 | Adjustment transaction | `wallet list --type adjustment` | Only adjustment transactions shown |
| D3.18 | Adjustment transaction | `wallet list` (default) | Adjustments excluded from default list |

### D4: Categories

| ID | Preconditions | Scenario | Expected Result |
|----|---------------|----------|-----------------|
| D4.1 | Wallet initialized | `wallet category list` | 32 categories listed (8 parents + 24 children) |
| D4.2 | Wallet initialized | `wallet category add "Freelance" --parent Income` | Custom category created under Income |
| D4.3 | Category exists | `wallet category edit "Freelance" --name "Side Hustle"` | Category renamed |
| D4.4 | Category exists, no usage | `wallet category rm "Side Hustle"` | Category archived |
| D4.5 | System category | `wallet category rm Food` | Error: cannot delete system category |
| D4.6 | Category has transactions | `wallet category rm Freelance` with transactions | Error: category has transactions |
| D4.7 | Empty name | `wallet category add ""` | Error: name required |

### D5: Tags

| ID | Preconditions | Scenario | Expected Result |
|----|---------------|----------|-----------------|
| D5.1 | Wallet initialized | `wallet tag list` | Empty list (no tags) |
| D5.2 | Wallet initialized | `wallet tag add "groceries"` | Tag created |
| D5.3 | Tag exists | `wallet tag add "groceries"` | Error: duplicate name |
| D5.4 | Tag exists with transactions | `wallet tag rm "groceries"` | Tag deleted, junction rows removed |
| D5.5 | Nonexistent tag | `wallet tag rm "fake"` | Error: not found |

### D6: Budgets

| ID | Preconditions | Scenario | Expected Result |
|----|---------------|----------|-----------------|
| D6.1 | Wallet initialized | `wallet budget set "Transport" 500000 -c Transport --period monthly` | Budget created |
| D6.2 | Budget with tags | `wallet budget set "Work Expenses" 2000000 --period monthly --tag "work"` | Budget created with tag filter |
| D6.3 | Budget exists | `wallet budget edit "Transport" --amount 600000` | Budget amount updated |
| D6.4 | Budget exists | `wallet budget list` | Budget name, period, spent, remaining, status shown |
| D6.5 | Multiple budgets | `wallet budget check` | All budgets evaluated, statuses reported |
| D6.6 | Nonexistent budget | `wallet budget check` with no budgets | Message: no active budgets |
| D6.7 | Budget with categories + tags | Add expense matching both category + tag | Counted in budget |
| D6.8 | Budget exists, multi-currency transactions | Add Food expense 50000 IDR and Food expense 10 USD (rate=16000) in same budget period | Budget spending = 50000 + 160000 = 210000 IDR (converted to base currency), not raw 50000 + 10 |
| D6.9 | Budget exists, multi-currency with missing rate | Add Food expense in a currency without configured rate | Budget spending taken from base_amount if available, or skipped with warning |
| D6.10 | Budget exists | `wallet budget rm "Transport"` | Budget archived (is_active = 0) |

### D7: Bills

| ID | Preconditions | Scenario | Expected Result |
|----|---------------|----------|-----------------|
| D7.1 | Wallet initialized | `wallet bill add "Rent" 5000000 -a Checking --recurrence monthly --due-date 2026-08-01` | Bill created |
| D7.2 | Bill with RRULE | `wallet bill add "Gym" 200000 -a Checking --recurrence "FREQ=WEEKLY;BYDAY=MO"` | Bill created with weekly recurrence |
| D7.3 | Bill exists | `wallet bill due --next 30` | Bills due within 30 days shown |
| D7.4 | Bill is paused | `wallet bill due` | Paused bills excluded from due list |
| D7.5 | Bill pay with amount override | `wallet bill pay 1 --amount 5200000` | Transaction created with custom amount |
| D7.6 | Invalid recurrence | `wallet bill add "Bad" 1000 -a Checking --recurrence "invalid"` | Error: invalid recurrence |
| D7.7 | Bill with missing account | `wallet bill add "Test" 1000 -a NonExistent` | Error: account not found |
| D7.8 | Bill edited with `--recurrence` | `wallet bill edit 1 --recurrence yearly` | Recurrence updated |

### D8: Exchange Rates

| ID | Preconditions | Scenario | Expected Result |
|----|---------------|----------|-----------------|
| D8.1 | Wallet initialized, no rates | `wallet rate list` | Empty list or message: no rates |
| D8.2 | No rates exist | `wallet rate add USD 15800` | Rate added |
| D8.3 | Invalid amount | `wallet rate add EUR 0` | Error: rate must be > 0 |
| D8.4 | Rate exists | `wallet rate set USD 16000` | Rate updated |
| D8.5 | Rate exists | `wallet rate list` | Rate shown with currency code and value |
| D8.6 | Rate exists | `wallet rate rm USD` | Rate removed |
| D8.7 | Rate assigned to an active account | `wallet rate rm USD` while USD account has transactions | Rate removed (accounts become unconvertible gracefully) |

### D9: Forecast

| ID | Preconditions | Scenario | Expected Result |
|----|---------------|----------|-----------------|
| D9.1 | Accounts with balance + bills | `wallet forecast -n 3` | 3-month projection table |
| D9.2 | Specific account | `wallet forecast -a Checking` | Only Checking account forecasted |
| D9.3 | Bills exist | `wallet forecast bills -n 3` | Upcoming bills schedule with running total |
| D9.4 | Bills in multiple currencies | Bills in IDR (300000) + USD (10 USD = 160000 IDR at rate 16000) | `wallet forecast bills` total_amount = 460000 IDR, running total also in base currency |
| D9.5 | Multi-currency with USD and IDR accounts + bills in both currencies | `wallet forecast -n 3` | All monthly start/ending balances in base currency, all bill amounts converted, negative balance detection accurate |
| D9.6 | Multi-currency forecast with missing rate | USD has transactions but no rate configured | `wallet forecast` includes warning about missing rate for USD |
| D9.7 | Unknown account | `wallet forecast -a Nonexistent` | Error: account not found |
| D9.8 | Invalid months | `wallet forecast -n 0` | Error: months must be > 0 |

### D10: JSON Output

| ID | Preconditions | Scenario | Expected Result |
|----|---------------|----------|-----------------|
| D10.1 | Any successful command | `wallet account list --json` | JSON: `{"success": true, "data": [...]}` |
| D10.2 | Any error command | `wallet add expense abc -a Fake -c Food --json` | JSON: `{"success": false, "error": {...}}` with error code |

### D11: Self-Update

| ID | Preconditions | Scenario | Expected Result |
|----|---------------|----------|-----------------|
| D11.1 | Internet available | `wallet version --check` | Shows current version vs latest (or message if up to date) |
| D11.2 | Internet available | `wallet version` | Shows installed version only |
| D11.3 | Internet available | `wallet update` | Downloads latest binary and updates |
