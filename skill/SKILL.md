---
name: wallet
description: CLI wallet for tracking personal finances. Supports transactions, budgets, bills, forecasts, and multi-currency exchange rates.
license: MIT
metadata:
  author: wallet-app
  version: "1.0"
  triggers:
    - expense
    - income
    - transfer
    - budget
    - bill
    - due
    - forecast
    - balance
    - spending
    - wallet
    - finance
    - money
    - payment
    - planned payment
    - recurring
    - exchange rate
    - category
    - tag
    - report
    - net worth
    - savings
    - subscription
    - allocate
    - overview
    - statement
    - track
    - record
    - "how much"
    - "how many"
    - remaining
    - leftover
    - paid
    - unpaid
    - payday
    - paycheck
---

## Wallet CLI Agent Skill

This skill instructs AI agents how to invoke the `wallet` CLI tool for personal finance operations. The wallet app supports structured JSON output via a global `--json` flag.

### Core Principles

1. **Always use `--json`** on every wallet command.
2. **Parse the JSON envelope**: success responses have `success`, `data`, and `meta`. Errors have `success: false` and an `error` object with `code`, `message`, and optional `suggestion`.
3. **Never auto-create resources**: if the user references a tag, category, or account that doesn't exist, tell them. Suggest the appropriate `wallet tag add`, `wallet category add`, or `wallet init` command.
4. **Present friendly output**: parse the JSON, then format a human-readable reply. Never dump raw JSON.
5. **Amounts are integers in smallest currency units**: `50000` = Rp 50,000 (IDR) or $500.00 (USD). Format with thousand separators and currency symbols for display.

### JSON Envelope

**Success** (stdout):
```json
{"success": true, "data": { /* payload */ }, "meta": {"command": "wallet <cmd>", "timestamp": "2026-07-04T12:00:00Z"}}
```

**Error** (stderr):
```json
{"success": false, "error": {"code": "CATEGORY_NOT_FOUND", "message": "...", "suggestion": "..."}}
```

Parse stdout first. If `success` is false, parse stderr for `error.code`, `error.message`, `error.suggestion`. Always relay suggestions to the user.

### Intent Mapping

| User says | Command |
|-----------|---------|
| "I spent/paid/bought X on Y" | `wallet add expense <amount> "<desc>" -c <cat> -a <acct> --json` |
| "I earned/got paid/received X" | `wallet add income <amount> "<desc>" -c <cat> -a <acct> --json` |
| "Transfer/move X from A to B" | `wallet add transfer <amount> --from A --to B --json` |
| "Show/list transactions [filter]" | `wallet list [-c <cat>] [-m <month>] [-a <acct>] [-t <tag>] [--type T] [-n <limit>] --json` |
| "Edit/change transaction #N" | `wallet edit N [--amount X] [--desc Y] [--add-tag Z] --json` |
| "Delete/remove/undo transaction #N" | `wallet rm N --force --json` |
| "Adjust/reconcile/correct account" | `wallet adjust <acct> <target> "<desc>" --json` |
| "How's my budget / am I on track?" | `wallet budget check --all --json` |
| "Set/create a budget for X limit Y" | `wallet budget set "<name>" <amount> -c <cat> --period monthly --json` |
| "List/show budgets" | `wallet budget list [--all] --json` |
| "Edit/change budget #N" | `wallet budget edit N [flags] --json` |
| "Delete/remove budget #N" | `wallet budget rm N --json` |
| "What bills are due?" | `wallet bill due [--overdue\|--week\|--next N] --json` |
| "List/show bills / subscriptions" | `wallet bill list [--paused\|--all] --json` |
| "Add/schedule a bill" | `wallet bill add "<name>" <amount> -c <cat> -a <acct> [--monthly\|--weekly\|--daily\|--yearly] --day N --json` |
| "Pay/settle bill #N" | `wallet bill pay N [--amount X] [--date YYYY-MM-DD] --json` |
| "Skip/postpone bill #N" | `wallet bill skip N --json` |
| "Pause/freeze/hold bill #N" | `wallet bill pause N --json` |
| "Resume/unpause bill #N" | `wallet bill resume N --json` |
| "Edit/change bill #N" | `wallet bill edit N [--name X] [--amount Y] [--recurrence Z] --json` |
| "Delete/remove bill #N" | `wallet bill rm N --json` |
| "Forecast/predict/project" | `wallet forecast -n 3 [--account A] --json` |
| "Upcoming bills / bill forecast" | `wallet forecast bills -n 3 --json` |
| "Report/summary/overview [for X]" | `wallet report [-m <month>] [-a <acct>] [-c <cat>] --json` |
| "Exchange rates" | `wallet rate list --json` |
| "Add/update rate for X = Y" | `wallet rate add <currency> <rate> --json` |
| "Categories list" | `wallet category list --json` |
| "Tags list" | `wallet tag list --json` |

### Command Quick Reference

| Command | Key Flags | Response `data` |
|---------|-----------|-----------------|
| `add expense <amt> <desc>` | `-c -a` (required), `-t -d -n` | `id,type,amount,currency,description,date,tags,base_amount,base_currency` |
| `add income <amt> <desc>` | `-c -a` (required), `-t -d -n` | same as expense, type="income" |
| `add transfer <amt>` | `--from --to` (required), `-d -n` | `id,type,amount,from_account,to_account,date,warning` |
| `list` | `-m -a -c -t --type --from --to -n` | `transactions[],total,count,base_total,base_currency` |
| `edit <id>` | `--amount -c -a -d --desc -n --add-tag --remove-tag` | `id,type,amount,description,date,tags` |
| `rm <id>` | `-f` (skip confirmation) | `status,id` |
| `adjust <acct> <target> <desc>` | `-n` | `account,old_balance,new_balance,difference,description` |
| `budget set <name> <amt>` | `-c -t --period --from --to --notify` | `id,name,amount,period,period_start,period_end,notify_at_pct,is_active,categories[],tags[]` |
| `budget list` | `--all` | array of `id,name,amount,spent,remaining,period,period_start,period_end,is_active` |
| `budget check` | `-b --all` | array of `id,name,limit,spent,remaining,percent_used,status` (ok/warning/over) |
| `budget edit <id>` | `--amount --name --notify --add-category --remove-category --add-tag --remove-tag` | same as set |
| `budget rm <id>` | none | `status,id` |
| `bill add <name> <amt>` | `-c -a` (required), `--daily\|--weekly\|--monthly\|--yearly\|--custom --rrule --from --day` | PlannedPayment object |
| `bill list` | `--paused --all` | `planned_payments[],count` |
| `bill due` | `--overdue --week --next N` | `due[],total,count` |
| `bill pay <id>` | `--date --amount` | `Transaction{},PlannedPayment{},NextDueDate` |
| `bill skip <id>` | none | PlannedPayment object (updated next_due_date) |
| `bill pause <id>` | none | PlannedPayment or `{status,id}` |
| `bill resume <id>` | none | PlannedPayment or `{status,id}` |
| `bill edit <id>` | `--name --amount -c -a --rrule --recurrence --from --day` | PlannedPayment object |
| `bill rm <id>` | none | `status,id` |
| `forecast` | `-n -a` | `horizon,forecast[],planned_payments[],category_breakdown[],warnings[],account` |
| `forecast bills` | `-n` | `horizon,bills[],total_amount,count,warnings[]` |
| `report` | `-m -a -c --from --to` | `base_currency,total_income,total_expense,net,by_category[],by_account[]` |
| `category list` | none | array of `id,name,type,icon,parent_id` |
| `category add <name>` | `-p --icon` | Category object |
| `category edit <id>` | `-n` (required) `--icon` | Category object |
| `category rm <id>` | none | `status,id` |
| `tag list` | none | array of `id,name` |
| `tag add <name>` | none | Tag object |
| `tag rm <name>` | none | `status,name` |
| `rate list` | none | `base_currency,rates{}` |
| `rate add <cur> <rate>` | none | `status,currency,rate,base` |
| `rate set <cur> <rate>` | none | `status,currency,rate,base` |
| `rate rm <cur>` | none | `status,currency` |
| `init` | none | `status,database,accounts,categories,message` |

### Error Codes

| Code | Meaning | Recovery |
|------|---------|----------|
| `CATEGORY_NOT_FOUND` | Category doesn't exist | List categories, suggest `wallet category add` |
| `ACCOUNT_NOT_FOUND` | Account doesn't exist | Check setup, suggest `wallet init` |
| `TAG_NOT_FOUND` | Tag doesn't exist | Suggest `wallet tag add <name>` (do NOT auto-create) |
| `TRANSACTION_NOT_FOUND` | Wrong transaction ID | List transactions to find the right ID |
| `BUDGET_NOT_FOUND` | Wrong budget ID/name | List budgets |
| `PLANNED_PAYMENT_NOT_FOUND` | Wrong bill ID | List bills |
| `INVALID_AMOUNT` | Amount is zero/negative | Ask for a positive amount |
| `INVALID_DATE` | Bad date format | Use YYYY-MM-DD |
| `INVALID_INPUT` / `VALIDATION_ERROR` | Generic validation | Read the `message` field |
| `BILL_PAUSED` | Bill is paused | User must `wallet bill resume <id>` first |
| `EXCHANGE_RATE_NOT_FOUND` | No rate for currency | Suggest `wallet rate add <currency> <rate>` |
| `EXCHANGE_RATE_CONFIG_MISSING` | No rate config | Run `wallet init` |
| `EXCHANGE_RATE_INVALID` | Negative/zero rate | Rate must be positive integer |
| `DB_ERROR` | Database failure | Suggest `wallet init` |
| `INTERNAL_ERROR` | Unexpected error | Report the message to the user |

Always relay `error.suggestion` when present.

### Common Workflows

**Monthly checkup:**
```bash
wallet report -m july --json
wallet budget check --all --json
wallet bill due --json
```

**Track subscriptions:**
```bash
wallet bill add "Netflix" 149000 -c Subscriptions -a bca --monthly --day 15 --json
wallet bill add "Spotify" 54990 -c Subscriptions -a bca --monthly --day 1 --json
wallet bill list --json
wallet forecast bills -n 3 --json
```

**Trip spending with a tag:**
```bash
wallet tag add "japan-2026" --json
wallet add expense 350000 "Shinkansen" -c Travel -a bca -t japan-2026 --json
wallet add expense 120000 "Ramen" -c Restaurant -a gopay -t japan-2026 --json
wallet list -t japan-2026 --json
```

**Payday routine:**
```bash
wallet add income 15000000 "Salary" -c Salary -a bca --json
wallet budget check --all --json
wallet bill due --next 30 --json
wallet forecast -n 1 --json
```

**Pay all due bills:** Get IDs from `wallet bill due --json`, then `wallet bill pay N --json` for each.

### Data Model

- **Accounts**: bank accounts/e-wallets with currencies. Created on init.
- **Categories**: spending/income categories (Food, Salary, etc). Pre-seeded on init.
- **Tags**: flexible labels for trips/projects. User-created only.
- **Transactions**: individual expense/income/transfer/adjustment entries. Belong to one account, one category, many tags.
- **Budgets**: spending limits per category/tag within a period (monthly/weekly/yearly/one_time). Auto-renews.
- **Planned Payments (Bills)**: scheduled recurring or one-time payments. Lifecycle: Active → (pay → next due | skip → next due | pause → Paused). One-time bills archive after payment.

### Multi-Currency

Each account has a currency. Transactions on foreign-currency accounts use that currency. Auto-converted to base currency (default IDR) when exchange rates are configured:
```bash
wallet rate add USD 15800 --json
wallet rate add EUR 17200 --json
```
Missing rate → `EXCHANGE_RATE_NOT_FOUND`. Tell the user to add it.

### Rules

- **Never auto-create tags, categories, or accounts.** Tell the user what's missing and how to create it.
- **Check `success` first** before reading `data`. Handle errors before presenting results.
- **Use `--force` with `rm`** in automated contexts to skip interactive prompts.
- **IDs are stable integers.** When a user says "#3" or "transaction 3", use that number.
- **Dates are YYYY-MM-DD.** Months accept names ("july") or numeric ("2026-07").
- **Paused bills block payment.** If `BILL_PAUSED`, tell user to `wallet bill resume <id>` first.
- **One-time bills archive after payment** and won't appear in `bill due` again.
- **Graceful degradation**: on failure, list available resources (categories, tags, bills) and suggest fixes. Don't give up on the first error.
