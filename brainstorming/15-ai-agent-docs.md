# 15 — AI Agent Documentation

> Depends on: [08-ai-native-layer](./08-ai-native-layer.md), [10-documentation](./10-documentation.md)
> Status: 🔴 pending review | Unblocks: implementation

---

## Objective

Create comprehensive documentation for AI agents in the `skill/` directory. Enables any AI agent to effectively use the wallet CLI.

---

## Decisions

### D1: Documentation Scope
| Option | Description |
|--------|-------------|
| A: Command Reference | Command list + flags + JSON output |
| B: Full Agent Guide | Commands + workflows + error codes + examples |
| **C: Both (multiple files)** | Separate files for different concerns |

→ **C — Multiple files in `skill/`:**

| File | Content |
|------|---------|
| `SKILL.md` | Existing — Hermes integration, provider-agnostic |
| `COMMANDS.md` | Command reference + JSON output format |
| `ERRORS.md` | Error codes + handling |
| `EXAMPLES.md` | Common workflows (add, list, report, etc.) |

### D2: COMMANDS.md Structure
| Option | Description |
|--------|-------------|
| A: Flat list | Alphabetical |
| **B: Grouped by domain** | Account/Transaction/Budget/etc. |
| C: Grouped by workflow | Record → List → Report |

→ **B — Grouped by domain.** AI agents think in terms of "I need to work with accounts."

---

## File: `skill/COMMANDS.md`

### Structure

```markdown
# Wallet CLI — Command Reference

## Transaction Commands

### `wallet add expense <amount> <description>`
Add an expense transaction.

**Flags:**
| Flag | Short | Required | Default | Description |
|------|-------|----------|---------|-------------|
| `--category` | `-c` | No | uncategorized | Category name or ID |
| `--account` | `-a` | No | config default | Account name or ID |
| `--tag` | `-t` | No | — | Tag name (repeatable) |
| `--date` | `-d` | No | today | Transaction date |
| `--json` | | No | false | JSON output |

**JSON Output:**
```json
{
  "success": true,
  "data": {
    "transaction": {
      "id": 42,
      "type": "expense",
      "amount": 35000,
      "currency": "IDR",
      "description": "Lunch at Warung",
      "date": "2026-07-05",
      "account_id": 1,
      "category_id": 2
    },
    "tags": ["lunch", "work"]
  }
}
```

### `wallet add income <amount> <description>`
[Similar structure...]

### `wallet add transfer <amount> --from <account> --to <account>`
[Similar structure...]

### `wallet list`
List transactions with filters.

**Flags:**
| Flag | Short | Description |
|------|-------|-------------|
| `--month` | `-m` | Month name or YYYY-MM |
| `--account` | `-a` | Account name or ID |
| `--category` | `-c` | Category name or ID |
| `--tag` | `-t` | Tag name |
| `--type` | | expense / income / transfer / adjustment |
| `--from` | | Start date (YYYY-MM-DD) |
| `--to` | | End date (YYYY-MM-DD) |
| `--limit` | `-n` | Max results (default: 20) |
| `--json` | | JSON output |

**JSON Output:**
```json
{
  "success": true,
  "data": {
    "transactions": [...],
    "count": 20,
    "total": -285000
  }
}
```

---

## Account Commands

### `wallet account add <name>`
### `wallet account list`
### `wallet account edit <id>`
### `wallet account archive <id>`

---

## Category Commands

### `wallet category list`
### `wallet category add <name>`
### `wallet category edit <id>`

---

## Tag Commands

### `wallet tag list`
### `wallet tag add <name>`
### `wallet tag rm <name>`

---

## Budget Commands

### `wallet budget create <name>`
### `wallet budget check`
### `wallet budget list`

---

## Bill Commands (Planned Payments)

### `wallet bill add <name>`
### `wallet bill list`
### `wallet bill pay <id>`
### `wallet bill skip <id>`
### `wallet bill pause <id>`

---

## Forecast Commands

### `wallet forecast`

---

## Report Commands

### `wallet report`
### `wallet report --by category`
### `wallet report --export csv`

---

## Rate Commands

### `wallet rate list`
### `wallet rate set`
```

---

## File: `skill/ERRORS.md`

### Structure

```markdown
# Wallet CLI — Error Codes

## Error Format

All errors in JSON mode follow this format:
```json
{
  "success": false,
  "error": {
    "code": "CATEGORY_NOT_FOUND",
    "message": "Category 'food' not found. Did you mean 'Food & Dining'?"
  }
}
```

## Error Codes

| Code | Message | Cause | Fix |
|------|---------|-------|-----|
| `CATEGORY_NOT_FOUND` | Category 'X' not found | Invalid category name | Use `wallet category list` to see options |
| `ACCOUNT_NOT_FOUND` | Account 'X' not found | Invalid account name | Use `wallet account list` to see options |
| `TAG_NOT_FOUND` | Tag 'X' not found | Tag doesn't exist | Create first: `wallet tag add X` |
| `INVALID_AMOUNT` | Amount must be a positive number | Amount <= 0 or non-numeric | Use positive integer (minor units) |
| `INVALID_DATE` | Invalid date format | Wrong format | Use YYYY-MM-DD or 'today'/'yesterday' |
| `DUPLICATE_NAME` | Name already exists | Duplicate account/category/tag | Use a different name |
| `BILL_PAUSED` | Bill is paused | Trying to pay paused bill | Unpause first: `wallet bill unpause <id>` |
| `BILL_NOT_FOUND` | Bill not found | Invalid bill ID | Use `wallet bill list` to find ID |
| `BUDGET_NOT_FOUND` | Budget not found | Invalid budget ID | Use `wallet budget list` to find ID |
| `BUDGET_EXCEEDED` | Budget exceeded | Over budget | Warning, not error |
| `TRANSACTION_NOT_FOUND` | Transaction not found | Invalid transaction ID | Use `wallet list` to find ID |
| `INSUFFICIENT_BALANCE` | Insufficient balance | Transfer amount > balance | Warning only, transfer still proceeds |

## Handling Errors

### For AI Agents
1. Parse JSON error response
2. Check `error.code` for specific handling
3. Use `error.message` to inform user
4. Suggest corrective action based on code

### Common Recovery Patterns

| Error | Recovery |
|-------|----------|
| `CATEGORY_NOT_FOUND` | List categories, suggest closest match |
| `ACCOUNT_NOT_FOUND` | List accounts, suggest closest match |
| `TAG_NOT_FOUND` | Offer to create tag first |
| `INVALID_AMOUNT` | Ask user for correct amount |
| `DUPLICATE_NAME` | Suggest alternative name |
```

---

## File: `skill/EXAMPLES.md`

### Structure

```markdown
# Wallet CLI — Common Workflows

## Recording Expenses

### Simple expense
```bash
wallet add expense 35000 "Lunch at Warung" --json
```

### Expense with category and account
```bash
wallet add expense 150000 "Groceries" -c groceries -a bca --json
```

### Expense with tags
```bash
wallet add expense 50000 "Coffee" -c coffee -t work -t morning --json
```

---

## Recording Income

### Salary
```bash
wallet add income 5000000 "July Salary" -c salary -a bca --json
```

---

## Transferring Money

### Between accounts
```bash
wallet add transfer 100000 --from bca --to gopay --json
```

---

## Checking Balances

### All accounts
```bash
wallet account list --json
```

### Specific account
```bash
wallet account list | grep "BCA"
```

---

## Listing Transactions

### This month
```bash
wallet list --json
```

### Specific month
```bash
wallet list --month 2026-06 --json
```

### By category
```bash
wallet list --category food --json
```

### By type
```bash
wallet list --type expense --json
```

---

## Budget Management

### Create budget
```bash
wallet budget create "Monthly Food" --amount 3000000 -c food -c restaurant --json
```

### Check budget status
```bash
wallet budget check --json
```

---

## Planned Payments

### Add recurring bill
```bash
wallet bill add "Netflix" --amount 149000 -a bca -c entertainment --recurrence monthly --json
```

### Pay a bill
```bash
wallet bill pay 1 --json
```

### Skip next occurrence
```bash
wallet bill skip 1 --json
```

---

## Forecasting

### Next 3 months
```bash
wallet forecast --months 3 --json
```

---

## Reports

### Monthly summary
```bash
wallet report --json
```

### By category
```bash
wallet report --by category --json
```

### Export to CSV
```bash
wallet report --export csv
```

---

## Balance Adjustment

### Reconcile with bank statement
```bash
wallet adjust bca 15000000 "Reconcile with bank statement" --json
```

---

## Multi-Currency

### Add foreign currency account
```bash
wallet account add "USD Savings" --type savings --currency USD
```

### Record foreign expense
```bash
wallet add expense 100 "Dinner" -c food -a usd_savings --json
```

### Set exchange rate
```bash
wallet rate set USD 15800
```
```

---

## File Structure

```
skill/
├── SKILL.md          # Existing — Hermes integration
├── COMMANDS.md       # Command reference (grouped by domain)
├── ERRORS.md         # Error codes + handling
└── EXAMPLES.md       # Common workflows
```

---

## Dependencies

- Phase 08: AI-Native Layer (JSON output, error codes)
- Phase 10: Documentation structure

---

## Ready to Review

Check:
- [ ] Multiple files approach OK?
- [ ] Grouped by domain OK?
- [ ] Error codes comprehensive?
- [ ] Workflows cover common use cases?
