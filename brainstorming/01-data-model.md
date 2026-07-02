# 01 — Data Model

> Depends on: none
> Status: ✅ design approved | Unblocks: 02-project-skeleton
> Decisions: Language=Go, Transfer=single-row, Budget=snapshot, Categories=2-level, Tags=both+categories

---

## Objective

Define the SQLite schema that serves as the single source of truth. Every future phase (CLI, transactions, budgeting, forecasting, AI integration) reads from and writes to these tables.

---

## Design Decisions (Approved)

| # | Decision | Choice | Rationale |
|---|----------|--------|-----------|
| D1 | Transfer representation | Single row with `transfer_to_id` | Transfers aren't income/expense; cleaner reporting |
| D2 | Budget period model | Snapshot per period (clone) | History preserved; limit changes don't corrupt past periods |
| D4 | Category hierarchy | 2-level (parent→child) | Grouped reporting without recursive CTE complexity |
| D5 | Tags + Categories | Both — budget by category OR tag | Categories=structure, Tags=cross-cutting + alternative budget target |

---

## Schema

### `accounts`

Money storage entities: bank account, cash, e-wallet, credit card.

```sql
CREATE TABLE accounts (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT NOT NULL,                    -- "BCA Checking", "GoPay", "Cash"
    type          TEXT NOT NULL DEFAULT 'checking', -- checking | savings | cash | credit_card | ewallet
    currency      TEXT NOT NULL DEFAULT 'IDR',      -- ISO 4217
    balance       INTEGER NOT NULL DEFAULT 0,       -- denormalized, minor units, updated on every write
    is_archived   INTEGER NOT NULL DEFAULT 0,
    sort_order    INTEGER NOT NULL DEFAULT 0,
    created_at    TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at    TEXT NOT NULL DEFAULT (datetime('now'))
);
```

### `categories`

2-level hierarchical classification. Flat list with optional parent.

```sql
CREATE TABLE categories (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT NOT NULL,                    -- "Restaurant", "Groceries"
    parent_id     INTEGER REFERENCES categories(id),-- NULL = top-level category
    type          TEXT NOT NULL DEFAULT 'expense',  -- expense | income | both
    icon          TEXT,                             -- "🍔"
    color         TEXT,                             -- "#FF5733"
    is_system     INTEGER NOT NULL DEFAULT 0,       -- built-in, non-deletable
    sort_order    INTEGER NOT NULL DEFAULT 0,
    created_at    TEXT NOT NULL DEFAULT (datetime('now'))
);
```

**Seed data (default categories):**

| Parent | Child | Type | Icon |
|--------|-------|------|------|
| Food & Dining | Restaurant | expense | 🍽️ |
| Food & Dining | Groceries | expense | 🛒 |
| Food & Dining | Coffee & Snacks | expense | ☕ |
| Transportation | Fuel | expense | ⛽ |
| Transportation | Public Transit | expense | 🚌 |
| Transportation | Ride-hailing | expense | 🚕 |
| Shopping | Clothing | expense | 👕 |
| Shopping | Electronics | expense | 📱 |
| Shopping | Household | expense | 🏠 |
| Bills & Utilities | Electricity | expense | ⚡ |
| Bills & Utilities | Internet | expense | 🌐 |
| Bills & Utilities | Phone | expense | 📞 |
| Bills & Utilities | Subscriptions | expense | 🔁 |
| Entertainment | Movies & Shows | expense | 🎬 |
| Entertainment | Gaming | expense | 🎮 |
| Entertainment | Travel | expense | ✈️ |
| Health | Medical | expense | 💊 |
| Health | Fitness | expense | 🏋️ |
| Education | Courses | expense | 📚 |
| Education | Books | expense | 📖 |
| Income | Salary | income | 💰 |
| Income | Freelance | income | 💻 |
| Income | Investment | income | 📈 |
| Income | Other | income | 💵 |

### `tags`

Freeform cross-cutting labels.

```sql
CREATE TABLE tags (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT NOT NULL UNIQUE,             -- "vacation", "reimbursable", "2026-japan"
    color         TEXT,
    created_at    TEXT NOT NULL DEFAULT (datetime('now'))
);
```

### `transactions`

Core entity — every money movement.

```sql
CREATE TABLE transactions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id      INTEGER NOT NULL REFERENCES accounts(id),
    category_id     INTEGER REFERENCES categories(id),
    type            TEXT NOT NULL,                   -- expense | income | transfer
    amount          INTEGER NOT NULL,                -- always positive, minor units
    currency        TEXT NOT NULL DEFAULT 'IDR',
    description     TEXT,                            -- "Lunch at Warung Sederhana"
    notes           TEXT,
    transfer_to_id  INTEGER REFERENCES accounts(id), -- D1: for transfers only
    date            TEXT NOT NULL,                   -- YYYY-MM-DD (business date)
    is_planned      INTEGER NOT NULL DEFAULT 0,       -- 1 = auto-generated from planned payment
    planned_payment_id INTEGER REFERENCES planned_payments(id),
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
);
```

### `transaction_tags`

M2M junction.

```sql
CREATE TABLE transaction_tags (
    transaction_id  INTEGER NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    tag_id          INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (transaction_id, tag_id)
);
```

### `budgets`

Per-period spending limit — target is category OR tag (polymorphic). D2: snapshot per period.

```sql
CREATE TABLE budgets (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id     INTEGER REFERENCES categories(id), -- D5: target = category (nullable if tag_id set)
    tag_id          INTEGER REFERENCES tags(id),        -- D5: target = tag (nullable if category_id set)
    name            TEXT,                                -- human-readable label
    amount          INTEGER NOT NULL,                    -- limit, minor units
    currency        TEXT NOT NULL DEFAULT 'IDR',
    period_start    TEXT NOT NULL,                       -- YYYY-MM-DD (first day of period)
    period_end      TEXT NOT NULL,                       -- YYYY-MM-DD (last day of period)
    rollover        INTEGER NOT NULL DEFAULT 0,          -- carry over unused amount to next period
    notify_at_pct   INTEGER DEFAULT 80,                  -- alert at X% (e.g., 80 = 80%)
    is_active       INTEGER NOT NULL DEFAULT 1,
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    CONSTRAINT target_check CHECK (
        (category_id IS NOT NULL AND tag_id IS NULL) OR
        (category_id IS NULL AND tag_id IS NOT NULL)
    )
);
```

**Example:** Monthly "Food" budget is `category_id = <Food & Dining parent id>`. Monthly "Japan Trip" budget is `tag_id = <#japan-2026 tag id>`.

### `planned_payments`

Recurring bills, subscriptions, and one-time future transactions.

```sql
CREATE TABLE planned_payments (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id      INTEGER NOT NULL REFERENCES accounts(id),
    category_id     INTEGER REFERENCES categories(id),
    type            TEXT NOT NULL DEFAULT 'expense',     -- expense | income
    amount          INTEGER NOT NULL,
    currency        TEXT NOT NULL DEFAULT 'IDR',
    description     TEXT NOT NULL,
    recurrence      TEXT NOT NULL DEFAULT 'none',        -- none | daily | weekly | monthly | yearly | custom
    recurrence_rule TEXT,                                -- RRULE (RFC 5545) for custom patterns
    start_date      TEXT NOT NULL,
    end_date        TEXT,
    next_due_date   TEXT,                                -- computed after each fulfillment
    is_paused       INTEGER NOT NULL DEFAULT 0,
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
);
```

### `exchange_rates`

Cached rates for multi-currency support.

```sql
CREATE TABLE exchange_rates (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    from_currency   TEXT NOT NULL,
    to_currency     TEXT NOT NULL,
    rate            REAL NOT NULL,
    source          TEXT DEFAULT 'manual',
    fetched_at      TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE UNIQUE INDEX idx_exchange_rates_unique
    ON exchange_rates(from_currency, to_currency, fetched_at);
```

---

## Entity Relationships

```
                    ┌──────────────┐
                    │   accounts   │
                    └──────┬───────┘
                           │ 1:N
       ┌───────────────────┼───────────────────┐
       │                   │                   │
  ┌────▼─────┐      ┌──────▼──────┐     ┌──────▼──────────┐
  │transactions│     │  budgets   │     │planned_payments │
  └──┬───┬────┘      └──┬────┬───┘     └──────┬──────────┘
     │   │               │    │                │
     │   │          ┌────┘    └────┐           │
     │   │          │              │           │
  ┌──▼─┐│    ┌──────▼──┐    ┌─────▼───┐       │
  │tags││    │categories│   │  tags   │       │
  └──┬─┘│    └──────────┘   └─────────┘       │
     │  │                                      │
  ┌──▼──▼──────┐                               │
  │trans_tags   │                               │
  └────────────┘                               │
                                               │
  ┌────────────────────────────────────────────┘
  │ (via planned_payment_id)
  ▼
transactions (is_planned=1)
```

---

## Open Questions

| # | Question | Status |
|---|----------|--------|
| OQ1 | Do we need `exchange_rates` in MVP? Multi-currency mungkin dimulai manual dulu | → Phase 07 |
| OQ2 | Transactions: store original amount+currency or convert to account currency? | → Phase 07 |
| OQ3 | `planned_payments.next_due_date` — computed at query time or pre-calculated on fulfill? | → Phase 05 |
