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

### Documentation Files

This skill is split into focused files for faster lookup:

| File | Purpose |
|------|---------|
| `COMMANDS.md` | Concise command inventory — signatures only, grouped by domain. Use `wallet <command> --help` to discover flags and `--json` for output shapes |
| `ERRORS.md` | All error codes with meaning, cause, recovery action, and common recovery patterns |
| `EXAMPLES.md` | Ready-to-use command sequences for common multi-step workflows |

Agents should consult these files for detailed command syntax, error handling, and workflow guidance.

Always relay `error.suggestion` when present.

### Common Workflows

See `EXAMPLES.md` for ready-to-use command sequences covering:
- Recording expenses, income, and transfers
- Payday routine and monthly checkup
- Subscription tracking (setup, review, pay)
- Trip spending with tags
- Budget management (setup, check, edit)
- Multi-currency setup and transfers
- Account setup and reconciliation
- Transaction management (list, edit, archive)
- Tag and category management
- Forecasting and report export

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
- **Flags before `--`, positional args after.** When passing negative amounts or other values that look like flags, use `--` to separate flags from positional args.

  ```
  wallet adjust "Bunga Bank" --json -- -3612 "Initial balance"
  │─────────┬────────────│─┬──│ │──┬──│ │──────┬──────│
       command + args    │ flags │ │  pos-1  │ │  pos-2  │
                         │        │ │(amount) │ │(desc)   │
                         │        │
                    -- ends flag parsing
  ```
- **Confirm before destructive operations.** Deleting, archiving, or adjusting financial data is hard to undo. Before running destructive commands (`wallet rm`, `wallet adjust`, `wallet account archive`, `wallet budget rm`, `wallet bill rm`, `wallet category rm`, `wallet tag rm`, `wallet rate rm`), show the user a brief summary of what will be affected (name, amount, description) and ask for explicit confirmation. For batch operations, describe the scope. Read-only queries and data creation commands are exempt.
- **Never touch the database directly.** Do not open the SQLite file, write raw SQL, or create scripts that manipulate database data. Always use the `wallet` CLI for all data operations (inserts, queries, updates, deletes).
