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
---

## Wallet CLI Agent Skill

This skill instructs AI agents how to invoke the `wallet` CLI tool for personal finance operations. The wallet app supports structured JSON output via a global `--json` flag, providing machine-readable responses for programmatic use.

### Core Principles

1. **Always use `--json`**: append the `--json` flag to every wallet command. This ensures structured, parseable responses.
2. **Parse the JSON envelope**: every successful response contains `success`, `data`, and `meta` fields. Errors contain `success: false` and an `error` object with `code`, `message`, and optional `suggestion`.
3. **Do not auto-create missing tags**: if the user references a tag that does not exist, do not create it implicitly. Ask the user or direct them to run `wallet tag add <name>`.
4. **Present results in friendly language**: after parsing the JSON response, format a human-friendly reply.

### JSON Response Envelope

```json
{
  "success": true,
  "data": { /* command-specific payload */ },
  "meta": {
    "command": "wallet <subcommand>",
    "timestamp": "2026-07-04T12:00:00Z"
  }
}
```

**Error envelope:**

```json
{
  "success": false,
  "error": {
    "code": "CATEGORY_NOT_FOUND",
    "message": "category 'food' not found",
    "suggestion": ""
  }
}
```

### Command Mapping

#### Recording Expenses

User: "I spent 50000 on lunch"
```bash
wallet add expense 50000 "Lunch" -c food -a bca --json
```

User: "Add 250000 expense for groceries at gopay with tag weekly"
```bash
wallet add expense 250000 "Groceries" -c groceries -a gopay -t weekly --json
```

User: "Record income of 5000000 salary"
```bash
wallet add income 5000000 "Salary" -c salary -a bca --json
```

#### Listing Transactions

User: "Show my recent transactions"
```bash
wallet list -n 10 --json
```

User: "Show me July expenses in category food"
```bash
wallet list -c food -m july --json
```

#### Budgets

User: "How is my budget doing?"
```bash
wallet budget check --all --json
```

User: "What budgets do I have?"
```bash
wallet budget list --json
```

User: "Set a monthly budget of 2000000 for food"
```bash
wallet budget set "Food Budget" 2000000 -c food --period monthly --json
```

#### Bills and Planned Payments

User: "What bills are due?"
```bash
wallet bill due --json
```

User: "Show overdue bills"
```bash
wallet bill due --overdue --json
```

User: "List my planned payments"
```bash
wallet bill list --json
```

User: "Pay bill #1"
```bash
wallet bill pay 1 --json
```

User: "Add a monthly bill for Netflix 149000"
```bash
wallet bill add "Netflix" 149000 -c entertainment -a bca --monthly --day 15 --json
```

#### Forecasts

User: "Forecast my balance for next 3 months"
```bash
wallet forecast -n 3 --json
```

User: "Show upcoming bills"
```bash
wallet forecast bills -n 3 --json
```

#### Reports

User: "Give me a financial report for July"
```bash
wallet report -m july --json
```

#### Exchange Rates

User: "Show exchange rates"
```bash
wallet rate list --json
```

### Error Handling

When a command returns `success: false`, inspect `error.code` for the specific failure. Common error codes:

| Code | Meaning |
|------|---------|
| `CATEGORY_NOT_FOUND` | Referenced category does not exist |
| `ACCOUNT_NOT_FOUND` | Referenced account does not exist |
| `TAG_NOT_FOUND` | Referenced tag does not exist |
| `BUDGET_NOT_FOUND` | Referenced budget does not exist |
| `PLANNED_PAYMENT_NOT_FOUND` | Referenced bill does not exist |
| `INVALID_AMOUNT` | Amount is invalid or not positive |
| `INVALID_DATE` | Date format is wrong |
| `VALIDATION_ERROR` | Input validation failed |
| `BILL_PAUSED` | Bill is paused and cannot be paid |
| `EXCHANGE_RATE_NOT_FOUND` | Exchange rate not configured |
| `DB_ERROR` | Database operation failed |
| `INTERNAL_ERROR` | Unexpected internal error |

If `error.suggestion` is present, share it with the user as helpful guidance.

### Important Rules

- **Never auto-create tags**: If the user mentions a tag that doesn't exist, say "The tag '<name>' doesn't exist. You can create it with `wallet tag add <name>`."
- **Always append `--json`**: Every wallet command should use `--json` for structured output.
- **Confirm data before displaying**: Always parse `success` first. If `false`, present the error to the user.
- **Format amounts**: Amounts are in the smallest currency unit (e.g., Rp). For display, you may format them as human-readable currency.
