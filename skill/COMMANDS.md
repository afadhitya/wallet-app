# Wallet CLI — Command Reference

> Domain-grouped command reference for AI agents. Each entry lists the command, key flags, and the JSON response `data` shape.
> Always add `--json` to every command invocation.

## Transaction Commands

### `wallet add expense <amount> <description>`

Record a spending transaction.

| Flag | Required | Description |
|------|----------|-------------|
| `-c, --category` | Yes | Category name or ID |
| `-a, --account` | Yes | Account name |
| `-t, --tag` | No | Tag name (repeatable) |
| `-d, --date` | No | Date in YYYY-MM-DD (default: today) |
| `-n, --note` | No | Additional note/memo |

**Response `data`:**
```json
{
  "id": 1,
  "type": "expense",
  "amount": 50000,
  "currency": "IDR",
  "description": "Lunch",
  "date": "2026-07-05",
  "tags": ["food"],
  "base_amount": 50000,
  "base_currency": "IDR"
}
```

### `wallet add income <amount> <description>`

Record income.

| Flag | Required | Description |
|------|----------|-------------|
| `-c, --category` | Yes | Category name or ID |
| `-a, --account` | Yes | Account name |
| `-t, --tag` | No | Tag name (repeatable) |
| `-d, --date` | No | Date in YYYY-MM-DD (default: today) |
| `-n, --note` | No | Additional note/memo |

**Response `data`:** Same shape as expense, `type: "income"`.

### `wallet add transfer <amount>`

Transfer money between own accounts.

| Flag | Required | Description |
|------|----------|-------------|
| `--from` | Yes | Source account name |
| `--to` | Yes | Destination account name |
| `-d, --date` | No | Date in YYYY-MM-DD (default: today) |
| `-n, --note` | No | Additional note/memo |

**Response `data`:**
```json
{
  "id": 1,
  "type": "transfer",
  "amount": 100000,
  "from_account": "BCA Checking",
  "to_account": "GoPay",
  "date": "2026-07-05",
  "warning": "string, if exchange rate missing for cross-currency transfer"
}
```

### `wallet list`

List and filter transactions.

| Flag | Description |
|------|-------------|
| `-m, --month` | Month filter: name ("july") or "YYYY-MM" |
| `-a, --account` | Filter by account name |
| `-c, --category` | Filter by category name |
| `-t, --tag` | Filter by tag name |
| `--type` | Filter: `expense`, `income`, `transfer` |
| `--from` | Start date YYYY-MM-DD |
| `--to` | End date YYYY-MM-DD |
| `-n, --limit` | Max results |

**Response `data`:**
```json
{
  "transactions": [ /* Transaction objects */ ],
  "total": 50000,
  "count": 5,
  "base_total": 50000,
  "base_currency": "IDR"
}
```

### `wallet edit <id>`

Edit a transaction.

| Flag | Description |
|------|-------------|
| `--amount` | New amount |
| `-c, --category` | Change category |
| `-a, --account` | Change account |
| `-d, --date` | Change date |
| `--desc` | Change description |
| `--note` | Change note |
| `--add-tag` | Add a tag |
| `--remove-tag` | Remove a tag |

**Response `data`:** Transaction object with updated fields.

### `wallet rm <id>`

Archive (soft-delete) a transaction.

| Flag | Description |
|------|-------------|
| `-f, --force` | Skip confirmation prompt |

**Response `data`:**
```json
{
  "status": "archived",
  "id": 1
}
```

### `wallet adjust <account> <target_balance> <description>`

Set an account balance to a target value. Creates a balancing adjustment transaction.

| Flag | Description |
|------|-------------|
| `-n, --note` | Additional note/memo |

**Response `data`:**
```json
{
  "account": "BCA Checking",
  "old_balance": 100000,
  "new_balance": 50000,
  "difference": -50000,
  "description": "Reconcile bank statement"
}
```

---

## Account Commands

### `wallet account add <name>`

Add a new account.

| Flag | Required | Description |
|------|----------|-------------|
| `--type` | Yes | Account type: `checking`, `savings`, `credit_card`, `cash`, `ewallet` |
| `--currency` | No | Currency code (default: base currency from config) |

**Response `data`:** Account object.

### `wallet account list`

List all active (non-archived) accounts.

**Response `data`:** Array of account objects with `id`, `name`, `type`, `currency`, `balance`, `sort_order`, `is_archived`.

### `wallet account edit <id>`

Edit an account.

| Flag | Description |
|------|-------------|
| `--name` | New name |
| `--type` | New type |
| `--currency` | New currency |

**Response `data`:** Updated account object.

### `wallet account archive <id>`

Archive an account (soft-delete). Leaves transaction history intact.

**Response `data`:**
```json
{
  "status": "archived",
  "id": 1
}
```

---

## Category Commands

### `wallet category list`

List all categories.

**Response `data`:** Array of `{ id, name, type, icon, parent_id }`.

### `wallet category add <name>`

Add a custom category.

| Flag | Description |
|------|-------------|
| `-p, --parent` | Parent category name (for hierarchy) |
| `--icon` | Icon name |

**Response `data`:** Category object.

### `wallet category edit <id>`

Edit a category.

| Flag | Required | Description |
|------|----------|-------------|
| `-n, --name` | Yes | New name |
| `--icon` | No | New icon |

**Response `data`:** Updated category object.

### `wallet category rm <id>`

Archive a category (soft-delete).

**Response `data`:**
```json
{
  "status": "archived",
  "id": 1
}
```

---

## Tag Commands

### `wallet tag list`

List all tags.

**Response `data`:** Array of `{ id, name }`.

### `wallet tag add <name>`

Create a new tag.

**Response `data`:** Tag object `{ id, name }`.

### `wallet tag rm <name>`

Delete a tag.

**Response `data`:**
```json
{
  "status": "deleted",
  "name": "japan-2026"
}
```

---

## Budget Commands

### `wallet budget set <name> <amount>`

Create a new budget.

| Flag | Required | Description |
|------|----------|-------------|
| `-c, --category` | No* | Category name (repeatable) |
| `-t, --tag` | No* | Tag name (repeatable) |
| `--period` | Yes | `monthly`, `weekly`, `yearly`, `one_time` |
| `--from` | No | Period start date |
| `--to` | No | Period end date |
| `--notify` | No | Notification threshold percentage (e.g., 80) |

*At least one category or tag is required.

**Response `data`:**
```json
{
  "id": 1,
  "name": "Food Budget",
  "amount": 2000000,
  "period": "monthly",
  "period_start": "2026-07-01",
  "period_end": "2026-07-31",
  "notify_at_pct": 80,
  "is_active": true,
  "categories": ["Restaurant", "Groceries"],
  "tags": []
}
```

### `wallet budget list`

List budgets.

| Flag | Description |
|------|-------------|
| `--all` | Include inactive budgets |

**Response `data`:** Array of `{ id, name, amount, spent, remaining, period, period_start, period_end, is_active }`.

### `wallet budget check`

Check budget status against spending.

| Flag | Description |
|------|-------------|
| `-b, --budget` | Check specific budget by ID |
| `--all` | Check all active budgets |

**Response `data`:** Array of:
```json
{
  "id": 1,
  "name": "Food Budget",
  "limit": 2000000,
  "spent": 500000,
  "remaining": 1500000,
  "percent_used": 25,
  "status": "ok"
}
```
Status values: `ok`, `warning`, `over`.

### `wallet budget edit <id>`

Edit a budget.

| Flag | Description |
|------|-------------|
| `--amount` | New amount |
| `--name` | New name |
| `--notify` | New notification threshold |
| `--add-category` | Add a category |
| `--remove-category` | Remove a category |
| `--add-tag` | Add a tag |
| `--remove-tag` | Remove a tag |

**Response `data`:** Updated budget object (same shape as `set`).

### `wallet budget rm <id>`

Deactivate a budget.

**Response `data`:**
```json
{
  "status": "deactivated",
  "id": 1
}
```

---

## Bill Commands (Planned Payments)

### `wallet bill add <name> <amount>`

Schedule a recurring or one-time bill.

| Flag | Required | Description |
|------|----------|-------------|
| `-c, --category` | Yes | Category name or ID |
| `-a, --account` | Yes | Account name |
| `--monthly` | No | Recur monthly |
| `--weekly` | No | Recur weekly |
| `--daily` | No | Recur daily |
| `--yearly` | No | Recur yearly |
| `--custom` | No | Custom recurrence |
| `--rrule` | No | Custom RRULE string (requires `--custom`) |
| `--from` | No | Start date |
| `--day` | No | Day of month/week for recurrence |

**Response `data`:** PlannedPayment object.

### `wallet bill list`

List planned payments.

| Flag | Description |
|------|-------------|
| `--paused` | Show only paused bills |
| `--all` | Include inactive bills |

**Response `data`:** `{ planned_payments: [...], count: N }`.

### `wallet bill due`

Show bills due.

| Flag | Description |
|------|-------------|
| `--overdue` | Show overdue bills |
| `--week` | Show bills due this week |
| `--next` | Show bills due in next N days |

**Response `data`:** `{ due: [...], total: N, count: N }`.

### `wallet bill pay <id>`

Pay a bill (creates an expense transaction).

| Flag | Description |
|------|-------------|
| `--date` | Payment date (default: today) |
| `--amount` | Payment amount (default: bill amount) |

**Response `data`:**
```json
{
  "transaction": { /* Transaction object */ },
  "planned_payment": { /* Updated PlannedPayment object */ },
  "next_due_date": "2026-08-01"
}
```

### `wallet bill skip <id>`

Skip a bill (advance to next due date without creating a transaction).

**Response `data`:** Updated PlannedPayment object.

### `wallet bill pause <id>`

Pause a bill.

**Response `data`:** Updated PlannedPayment or `{ status: "paused", id: N }`.

### `wallet bill resume <id>`

Resume a paused bill.

**Response `data`:** Updated PlannedPayment or `{ status: "resumed", id: N }`.

### `wallet bill edit <id>`

Edit a bill.

| Flag | Description |
|------|-------------|
| `--name` | New name |
| `--amount` | New amount |
| `-c, --category` | New category |
| `-a, --account` | New account |
| `--rrule` | New custom RRULE (with `--custom` recurrence) |
| `--recurrence` | New recurrence type |
| `--from` | New start date |
| `--day` | New day of month/week |

**Response `data`:** Updated PlannedPayment object.

### `wallet bill rm <id>`

Remove (deactivate) a bill.

**Response `data`:**
```json
{
  "status": "deactivated",
  "id": 1
}
```

---

## Forecast Commands

### `wallet forecast`

Project future account balances with bill deductions.

| Flag | Description |
|------|-------------|
| `-n, --months` | Number of months to project (default: 1) |
| `-a, --account` | Filter by account name |

**Response `data`:**
```json
{
  "horizon": "2026-07-05 to 2026-08-05",
  "forecast": [ /* Month projections with balances */ ],
  "planned_payments": [ /* Upcoming bills within horizon */ ],
  "category_breakdown": [ /* Projected spending by category */ ],
  "warnings": [ /* Low balance warnings, missing rates */ ],
  "account": "BCA Checking"
}
```

### `wallet forecast bills`

Show upcoming bill schedule.

| Flag | Description |
|------|-------------|
| `-n, --months` | Number of months to project |

**Response `data`:**
```json
{
  "horizon": "2026-07-05 to 2026-10-05",
  "bills": [ /* Upcoming bill events */ ],
  "total_amount": 5000000,
  "count": 12,
  "warnings": [ /* Any warnings */ ]
}
```

---

## Report Commands

### `wallet report`

Generate a financial summary report.

| Flag | Description |
|------|-------------|
| `-m, --month` | Month: name ("july") or "YYYY-MM" |
| `-a, --account` | Filter by account |
| `-c, --category` | Filter by category |
| `--from` | Start date |
| `--to` | End date |
| `--by` | Grouping: `category`, `account`, `tag` |
| `--export` | Export format: `csv` |

**Response `data`:**
```json
{
  "base_currency": "IDR",
  "total_income": 15000000,
  "total_expense": 8000000,
  "net": 7000000,
  "by_category": [ /* { category, amount, percentage } */ ],
  "by_account": [ /* { account, income, expense, net } */ ]
}
```

---

## Rate Commands

### `wallet rate list`

List all exchange rates.

**Response `data`:**
```json
{
  "base_currency": "IDR",
  "rates": {
    "USD": 15800,
    "EUR": 17200
  }
}
```

### `wallet rate add <currency> <rate>`

Add a new exchange rate.

**Response `data`:**
```json
{
  "status": "added",
  "currency": "USD",
  "rate": 15800,
  "base": "IDR"
}
```

### `wallet rate set <currency> <rate>`

Update an existing exchange rate.

**Response `data`:**
```json
{
  "status": "updated",
  "currency": "USD",
  "rate": 16000,
  "base": "IDR"
}
```

### `wallet rate rm <currency>`

Remove an exchange rate.

**Response `data`:**
```json
{
  "status": "removed",
  "currency": "USD"
}
```

---

## Init Command

### `wallet init`

Initialize the wallet database and create default configuration.

**Response `data`:**
```json
{
  "status": "initialized",
  "database": "/path/to/wallet.db",
  "accounts": [ /* Default account */ ],
  "categories": [ /* 32 seeded categories */ ],
  "message": "Wallet initialized successfully"
}
```
