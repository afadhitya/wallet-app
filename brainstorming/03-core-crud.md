# 03 — Core CRUD

> Depends on: [01-data-model](./01-data-model.md), [02-project-skeleton](./02-project-skeleton.md)
> Status: 🔴 pending review | Unblocks: 04-budget-engine

---

## Objective

Implement the core CRUD operations for transactions, categories, and tags — the backbone of every other feature. This phase produces the first usable CLI commands.

---

## Design Decisions

| # | Decision | Choice | Rationale |
|---|----------|--------|-----------|
| C1 | Transaction entry | Hybrid — positional (amount, desc), flags (-c, -a, -t) | Quick entry + flexibility |
| C2 | List/filter | Simple flags (--month, --category, --account, --tag) | Standard Go CLI, Cobra native, --help discoverable |
| C3 | Edit/delete | Separate commands (wallet edit <id>, wallet rm <id>) | Stateless, scriptable, Hermes-friendly |
| C4 | Category management | Full CRUD (add, edit, list, soft delete) | Seed data as default, user customizes over time |

---

## Commands

### `wallet init`

Creates the database, runs migrations, seeds default categories.

```
$ wallet init
✓ Database created at ~/.local/share/wallet/wallet.db
✓ Schema migrated (1 migration)
✓ Seeded 24 default categories
```

**Behavior:**
- Idempotent — safe to run multiple times
- If DB exists, check schema version and migrate if needed
- Creates config file at `~/.config/wallet/config.toml` if missing

---

### `wallet add expense`

Add an expense transaction.

```
$ wallet add expense 35000 "Lunch at Warung" -c food -a bca -t lunch -t work
✓ Recorded: Rp35.000 — Lunch at Warung (Food & Dining > Restaurant) [BCA]
```

**Flags:**

| Flag | Short | Required | Default | Description |
|------|-------|----------|---------|-------------|
| `--category` | `-c` | No | uncategorized | Category name or ID |
| `--account` | `-a` | No | config default | Account name or ID |
| `--tag` | `-t` | No | — | Tag name (repeatable) |
| `--date` | `-d` | No | today | Transaction date |
| `--description` | | No | — | Alias for positional desc |
| `--notes` | | No | — | Additional notes |
| `--json` | | No | false | JSON output |

**Validation:**
- Amount must be > 0
- Category must exist (error + suggest: "Did you mean 'food'?")
- Account must exist and not be archived
- Tags must exist (error if not: `Tag 'xyz' not found. Create it first with: wallet tag add xyz`)
- Date must be valid YYYY-MM-DD or alias (today, yesterday)

---

### `wallet add income`

Same as expense but `type=income`.

```
$ wallet add income 5000000 "Gaji Juli" -c salary -a bca
✓ Recorded: Rp5.000.000 — Gaji Juli (Income > Salary) [BCA]
```

---

### `wallet add transfer`

Transfer between accounts.

```
$ wallet add transfer 100000 --from bca --to gopay
✓ Transfer: Rp100.000 — BCA → GoPay
```

**Flags:**

| Flag | Short | Required | Default |
|------|-------|----------|---------|
| `--from` | | Yes | — |
| `--to` | | Yes | — |
| `--date` | `-d` | No | today |
| `--notes` | | No | — |
| `--json` | | No | false |

**Validation:**
- `--from` and `--to` must be different accounts
- Both accounts must exist
- Source account balance checked (warn if insufficient, don't block)

---

### `wallet list`

List transactions with filters.

```
$ wallet list --month july --category food
┌────┬────────────┬──────────────────────┬────────────┬───────────┐
│ ID │ Date       │ Description          │ Category   │ Amount    │
├────┼────────────┼──────────────────────┼────────────┼───────────┤
│ 42 │ 2026-07-01 │ Lunch at Warung      │ Restaurant │ Rp35.000  │
│ 38 │ 2026-07-01 │ Groceries            │ Groceries  │ Rp250.000 │
└────┴────────────┴──────────────────────┴────────────┴───────────┘
                                          Total: Rp285.000
```

**Filters:**

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

**Default:** Current month, all accounts, all categories, limit 20.

---

### `wallet edit <id>`

Edit an existing transaction.

```
$ wallet edit 42 --amount 40000 --category transport
✓ Updated transaction #42: Rp40.000 — Transport
```

**Editable fields:**
- `--amount` — amount
- `--category` — category
- `--account` — account
- `--date` — date
- `--description` — description
- `--notes` — notes
- `--add-tag` / `--remove-tag` — tag management

**Behavior:**
- Only update fields explicitly passed
- Update `updated_at` timestamp
- Recalculate account balance if amount or account changed

---

### `wallet rm <id>`

Delete a transaction (soft delete via `is_archived` or hard delete).

```
$ wallet rm 42
✓ Deleted transaction #42: Rp35.000 — Lunch at Warung
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--force` | Skip confirmation prompt |
| `--json` | JSON output |

**Behavior:**
- Soft delete — set `is_archived = 1` (OQ1: soft delete)
- Recalculate account balance
- Prompt for confirmation unless `--force`

---

### `wallet category list`

List all categories.

```
$ wallet category list
┌────┬─────────────────┬────────────────┬──────┬───────┐
│ ID │ Name            │ Parent         │ Type │ Icon  │
├────┼─────────────────┼────────────────┼──────┼───────┤
│  1 │ Food & Dining   │ —              │ exp  │ 🍔    │
│  2 │ Restaurant      │ Food & Dining  │ exp  │ 🍽️    │
│  3 │ Groceries       │ Food & Dining  │ exp  │ 🛒    │
└────┴─────────────────┴────────────────┴──────┴───────┘
```

---

### `wallet category add`

```
$ wallet category add "Kopi" --parent "Food & Dining" --icon ☕
✓ Category created: Food & Dining > Kopi ☕
```

---

### `wallet category edit <id>`

```
$ wallet category edit 3 --icon 🛒 --name "Groceries & Supermarket"
✓ Updated category #3
```

---

### `wallet tag list`

```
$ wallet tag list
┌────┬──────────────┬───────┐
│ ID │ Name         │ Color │
├────┼──────────────┼───────┤
│  1 │ lunch        │ —     │
│  2 │ work         │ —     │
│  3 │ japan-2026   │ —     │
└────┴──────────────┴───────┘
```

---

### `wallet tag add`

```
$ wallet tag add "japan-2026"
✓ Tag created: japan-2026
```

---

### `wallet tag rm <name>`

```
$ wallet tag rm "japan-2026"
✓ Tag deleted: japan-2026
```

---

### `wallet adjust`

Adjust account balance — not income/expense, just a correction.

```
$ wallet adjust bca 15000000 "Reconcile with bank statement"
✓ Balance adjusted: BCA = Rp15.000.000 (was Rp14.850.000, diff +Rp150.000)
  Adjustment recorded (tx #99)
```

**Flags:**

| Flag | Short | Required | Default | Description |
|------|-------|----------|---------|-------------|
| `--notes` | | No | — | Reason for adjustment |
| `--json` | | No | false | JSON output |

**Behavior:**
- Set account balance to exact amount
- Create transaction with `type='adjustment'`, `amount=|diff|`
- Positive diff: adjustment increases balance
- Negative diff: adjustment decreases balance
- Excluded from income/expense reports
- Visible in `wallet list --type adjustment`

**Examples:**

```
# Balance too high, need to decrease
$ wallet adjust gopay 50000 "Cash out not recorded"
✓ Balance adjusted: GoPay = Rp50.000 (was Rp120.000, diff -Rp70.000)

# Balance too low, need to increase
$ wallet adjust bca 15000000 "Found missing deposit"
✓ Balance adjusted: BCA = Rp15.000.000 (was Rp14.500.000, diff +Rp500.000)
```

---

## Service Layer

### TransactionService

```go
type TransactionService struct {
    db *gen.Queries
}

func (s *TransactionService) CreateExpense(ctx, params) (*Transaction, error)
func (s *TransactionService) CreateIncome(ctx, params) (*Transaction, error)
func (s *TransactionService) CreateTransfer(ctx, params) (*Transaction, error)
func (s *TransactionService) List(ctx, filters) ([]Transaction, error)
func (s *TransactionService) Update(ctx, id, params) (*Transaction, error)
func (s *TransactionService) Delete(ctx, id) error
func (s *TransactionService) RecalcBalance(ctx, accountID) error
```

### AccountService

```go
type AccountService struct {
    db *gen.Queries
}

func (s *AccountService) Create(ctx, params) (*Account, error)
func (s *AccountService) List(ctx) ([]Account, error)
func (s *AccountService) Update(ctx, id, params) (*Account, error)
func (s *AccountService) Archive(ctx, id) error
```

---

## sqlc Queries

### accounts.sql

```sql
-- name: GetAccount :one
SELECT * FROM accounts WHERE id = ?;

-- name: ListAccounts :many
SELECT * FROM accounts WHERE is_archived = 0 ORDER BY sort_order, name;

-- name: CreateAccount :one
INSERT INTO accounts (name, type, currency) VALUES (?, ?, ?) RETURNING *;

-- name: UpdateAccount :exec
UPDATE accounts SET name = ?, type = ?, sort_order = ?, updated_at = datetime('now') WHERE id = ?;

-- name: ArchiveAccount :exec
UPDATE accounts SET is_archived = 1, updated_at = datetime('now') WHERE id = ?;

-- name: UpdateBalance :exec
UPDATE accounts SET balance = ?, updated_at = datetime('now') WHERE id = ?;
```

### transactions.sql

```sql
-- name: GetTransaction :one
SELECT t.*, GROUP_CONCAT(tg.name) as tags
FROM transactions t
LEFT JOIN transaction_tags tt ON t.id = tt.transaction_id
LEFT JOIN tags tg ON tt.tag_id = tg.id
WHERE t.id = ?
GROUP BY t.id;

-- name: ListTransactions :many
SELECT t.*, GROUP_CONCAT(tg.name) as tags
FROM transactions t
LEFT JOIN transaction_tags tt ON t.id = tt.transaction_id
LEFT JOIN tags tg ON tt.tag_id = tg.id
WHERE t.date >= ? AND t.date <= ?
  AND t.is_archived = 0
  AND (? = 0 OR t.account_id = ?)
  AND (? = 0 OR t.category_id = ?)
  AND (? = '' OR t.type = ?)
GROUP BY t.id
ORDER BY t.date DESC, t.id DESC
LIMIT ?;

-- name: CreateTransaction :one
INSERT INTO transactions (account_id, category_id, type, amount, currency, base_amount, base_currency, description, notes, transfer_to_id, date, is_planned, planned_payment_id)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: UpdateTransaction :exec
UPDATE transactions SET category_id = ?, amount = ?, description = ?, notes = ?, date = ?, updated_at = datetime('now') WHERE id = ?;

-- name: DeleteTransaction :exec
UPDATE transactions SET is_archived = 1, updated_at = datetime('now') WHERE id = ?;

-- name: SumByAccount :one
SELECT COALESCE(SUM(CASE
    WHEN type = 'income' THEN amount
    WHEN type = 'adjustment' THEN amount
    WHEN type = 'transfer' AND transfer_to_id = ? THEN amount
    ELSE -amount
END), 0)
FROM transactions WHERE (account_id = ? OR transfer_to_id = ?) AND is_archived = 0;
```

---

## Error Handling

| Error | Message | Exit code |
|-------|---------|-----------|
| Category not found | `Category 'foo' not found. Did you mean 'food'?` | 1 |
| Account not found | `Account 'bar' not found.` | 1 |
| Invalid amount | `Amount must be a positive number.` | 1 |
| Invalid date | `Invalid date format. Use YYYY-MM-DD or 'today'/'yesterday'.` | 1 |
| Insufficient balance | `Warning: BCA balance (Rp50.000) < transfer amount (Rp100.000)` | 0 (warn only) |
| Transaction not found | `Transaction #99 not found.` | 1 |

---

## Testing Strategy

- **Unit tests:** Service layer with in-memory SQLite (`:memory:`)
- **Integration tests:** CLI commands via `exec.Command`, verify stdout + exit code
- **Coverage:** 100% (enforced by CI)
- **Test helpers:** `testutil.SetupDB(t)` returns clean DB + seed data
- **Fixtures:** Test accounts, categories, tags created per test

---

## Resolved Questions

| # | Question | Resolution |
|---|----------|------------|
| OQ1 | Soft delete vs hard delete? | Soft delete — mark archived, keep data |
| OQ2 | Auto-create tags on add? | No — error if tag not exist |
| OQ3 | Transaction ID format? | Auto-increment integer |
