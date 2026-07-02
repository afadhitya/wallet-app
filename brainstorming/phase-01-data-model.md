# Phase 01 — Data Model

> Status: 🔴 pending | Depends on: none | Unblocks: Phase 02

---

## Objective

Define the SQLite schema that will serve as the single source of truth for the entire application. Every future phase (transactions, budgeting, forecasting, AI integration) reads from and writes to these tables.

---

## Schema Design

### `accounts`

Represents a money storage entity: checking account, savings, cash wallet, credit card, e-wallet, etc.

```sql
CREATE TABLE accounts (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT NOT NULL,                    -- "BCA Checking", "GoPay", "Cash"
    type          TEXT NOT NULL DEFAULT 'checking', -- checking | savings | cash | credit_card | ewallet | investment
    currency      TEXT NOT NULL DEFAULT 'IDR',      -- ISO 4217
    balance       INTEGER NOT NULL DEFAULT 0,       -- stored in minor units (cents/sen)
    is_archived   INTEGER NOT NULL DEFAULT 0,       -- soft delete
    sort_order    INTEGER NOT NULL DEFAULT 0,
    created_at    TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at    TEXT NOT NULL DEFAULT (datetime('now'))
);
```

**Design notes:**
- `balance` is **computed from transactions** at query time or cached as a denormalized field refreshed periodically.
- `type` drives UI behavior: credit cards have negative balance semantics.
- All monetary values stored as integers (minor units) to avoid floating-point errors.

### `categories`

Hierarchical category tree for classifying transactions.

```sql
CREATE TABLE categories (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT NOT NULL,                    -- "Food & Dining"
    parent_id     INTEGER REFERENCES categories(id),
    type          TEXT NOT NULL DEFAULT 'expense',  -- expense | income | both
    icon          TEXT,                             -- emoji or icon code: "🍔"
    color         TEXT,                             -- hex: "#FF5733"
    is_system     INTEGER NOT NULL DEFAULT 0,       -- built-in, not deletable
    sort_order    INTEGER NOT NULL DEFAULT 0,
    created_at    TEXT NOT NULL DEFAULT (datetime('now'))
);
```

**Default categories (seed data):**

| Category | Type | Icon |
|----------|------|------|
| Food & Dining | expense | 🍔 |
| Transportation | expense | 🚗 |
| Shopping | expense | 🛍️ |
| Bills & Utilities | expense | 📄 |
| Entertainment | expense | 🎬 |
| Health | expense | 💊 |
| Education | expense | 📚 |
| Housing | expense | 🏠 |
| Income | income | 💰 |
| Investments | income | 📈 |
| Transfer | both | 🔄 |

### `tags`

Freeform labels for cross-cutting transaction organization.

```sql
CREATE TABLE tags (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT NOT NULL UNIQUE,             -- "tax-deductible", "vacation", "reimbursable"
    color         TEXT,                             -- hex
    created_at    TEXT NOT NULL DEFAULT (datetime('now'))
);
```

### `transactions`

The core entity — every money movement.

```sql
CREATE TABLE transactions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id      INTEGER NOT NULL REFERENCES accounts(id),
    category_id     INTEGER REFERENCES categories(id),
    type            TEXT NOT NULL,                   -- expense | income | transfer
    amount          INTEGER NOT NULL,                -- always positive, minor units
    currency        TEXT NOT NULL DEFAULT 'IDR',     -- ISO 4217
    description     TEXT,                            -- "Lunch at Warung Sederhana"
    notes           TEXT,                            -- optional longer notes
    transfer_to_id  INTEGER REFERENCES accounts(id), -- for transfers (nullable)
    date            TEXT NOT NULL,                   -- YYYY-MM-DD (user-specified)
    is_planned      INTEGER NOT NULL DEFAULT 0,       -- 1 = from planned payment
    planned_payment_id INTEGER REFERENCES planned_payments(id),
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
);
```

**Design notes:**
- `currency` on the transaction allows mixed-currency recording. Reports normalize to account currency.
- `date` is the business date (when the transaction happened), not `created_at` (when it was entered).
- Transfers: `type='transfer'`, `transfer_to_id` points to destination account. Two rows per transfer? Or one row with source+dest? **Decision needed.**
- `is_planned` flag tracks whether this transaction originated from a planned payment.

### `transaction_tags`

Many-to-many junction between transactions and tags.

```sql
CREATE TABLE transaction_tags (
    transaction_id  INTEGER NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    tag_id          INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (transaction_id, tag_id)
);
```

### `budgets`

Monthly spending limits per category (or group of categories).

```sql
CREATE TABLE budgets (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id     INTEGER REFERENCES categories(id), -- NULL = "all categories"
    name            TEXT,                                -- "Monthly Food Budget"
    amount          INTEGER NOT NULL,                    -- minor units
    currency        TEXT NOT NULL DEFAULT 'IDR',
    period          TEXT NOT NULL DEFAULT 'monthly',     -- weekly | monthly | yearly
    start_date      TEXT NOT NULL,                       -- YYYY-MM-DD
    end_date        TEXT,                                -- NULL = indefinite
    rollover        INTEGER NOT NULL DEFAULT 0,          -- unused amount carries over
    notify_at       INTEGER,                             -- alert at X% (e.g., 80 = 80%)
    is_active       INTEGER NOT NULL DEFAULT 1,
    created_at      TEXT NOT NULL DEFAULT (datetime('now'))
);
```

**Design notes:**
- One budget row = one period's limit for one category. For recurring monthly budgets, a new row is created each period (or we use a template + instance pattern — **decision needed**).
- Alternative: single budget row with recurring flag, and a separate `budget_periods` table for actual period instances.

### `planned_payments`

Recurring bills, subscriptions, and one-time future expenses/income.

```sql
CREATE TABLE planned_payments (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id      INTEGER NOT NULL REFERENCES accounts(id),
    category_id     INTEGER REFERENCES categories(id),
    type            TEXT NOT NULL DEFAULT 'expense',     -- expense | income
    amount          INTEGER NOT NULL,                    -- minor units
    currency        TEXT NOT NULL DEFAULT 'IDR',
    description     TEXT NOT NULL,                       -- "Netflix Subscription"
    recurrence      TEXT NOT NULL DEFAULT 'none',        -- none | daily | weekly | monthly | yearly | custom
    recurrence_rule TEXT,                                -- RRULE format (RFC 5545) for complex patterns
    start_date      TEXT NOT NULL,                       -- YYYY-MM-DD
    end_date        TEXT,                                -- NULL = indefinite
    next_due_date   TEXT,                                -- computed, updated after each fulfillment
    is_paused       INTEGER NOT NULL DEFAULT 0,
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
);
```

**Recurrence examples:**

| Description | recurrence | recurrence_rule |
|-------------|-----------|-----------------|
| Netflix monthly | monthly | — |
| Gym every 2 weeks | custom | `FREQ=WEEKLY;INTERVAL=2` |
| Rent, 1st of month | monthly | `FREQ=MONTHLY;BYMONTHDAY=1` |
| Dentist April 15 | none | — |

### `exchange_rates`

Cached exchange rates for multi-currency support.

```sql
CREATE TABLE exchange_rates (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    from_currency   TEXT NOT NULL,                     -- "USD"
    to_currency     TEXT NOT NULL,                     -- "IDR"
    rate            REAL NOT NULL,                     -- 1 USD = 15700 IDR
    source          TEXT DEFAULT 'manual',             -- manual | api
    fetched_at      TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(from_currency, to_currency, fetched_at)
);
```

---

## Entity Relationship Diagram

```
                    ┌──────────────┐
                    │   accounts   │
                    └──────┬───────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
    ┌─────────▼──────┐ ┌──▼──────────┐ ┌▼──────────────┐
    │  transactions  │ │   budgets   │ │planned_payments│
    └──┬──────┬──────┘ └────┬───────┘ └───────┬───────┘
       │      │              │                  │
  ┌────▼──┐ ┌─▼────────┐    │                  │
  │ tags  │ │categories│◄───┘                  │
  └───┬───┘ └──────────┘                       │
      │                                         │
  ┌───▼───────────┐                             │
  │transaction_tags│                            │
  └───────────────┘    ┌────────────────┐       │
                       │ exchange_rates │       │
                       └────────────────┘       │
                                                │
                         ┌──────────────────────┘
                         │
                         ▼
                  ┌──────────────┐
                  │ transactions │ (is_planned=1)
                  └──────────────┘
```

---

## Open Design Decisions

### D1: Transfer representation
**Option A:** One transaction row with `type='transfer'` and `transfer_to_id`
**Option B:** Two transaction rows (expense from source + income to destination), linked by `transfer_group_id`

→ **Recommendation: A** — simpler, single row, balance query subtracts from source and adds to dest.

### D2: Budget period model
**Option A:** Single `budgets` row = template, new row cloned per period
**Option B:** Single row, period filtering via `start_date + period` in queries

→ **Recommendation: A** — cleaner, each period has its own snapshot of limit/spent.

### D3: Balance caching
**Option A:** Compute balance from transactions every query (SUM)
**Option B:** Denormalized `accounts.balance` updated on every transaction write

→ **Recommendation: B for MVP** — single user, write volume is low, simpler queries. Option A if performance becomes an issue (won't).

### D4: Category hierarchy depth
**Flat** vs **2-level** (parent → child) vs **unlimited nesting**

→ **Recommendation: 2-level** — `parent_id` supports 1 nesting level. Unlimited nesting adds query complexity with recursive CTEs for marginal UX benefit.

---

## Next Steps

- [ ] Decide D1–D4
- [ ] Validate schema against all Phase 03–07 requirements
- [ ] Proceed to Phase 02: Project Skeleton
