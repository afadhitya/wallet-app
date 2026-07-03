PRAGMA foreign_keys = ON;

CREATE TABLE accounts (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT NOT NULL,
    type          TEXT NOT NULL DEFAULT 'checking',
    currency      TEXT NOT NULL DEFAULT 'IDR',
    balance       INTEGER NOT NULL DEFAULT 0,
    is_archived   INTEGER NOT NULL DEFAULT 0,
    sort_order    INTEGER NOT NULL DEFAULT 0,
    created_at    TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at    TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE categories (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT NOT NULL,
    parent_id     INTEGER REFERENCES categories(id),
    type          TEXT NOT NULL DEFAULT 'expense',
    icon          TEXT,
    color         TEXT,
    is_system     INTEGER NOT NULL DEFAULT 0,
    sort_order    INTEGER NOT NULL DEFAULT 0,
    created_at    TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE UNIQUE INDEX idx_categories_unique_parent ON categories(name) WHERE parent_id IS NULL;
CREATE UNIQUE INDEX idx_categories_unique_child ON categories(parent_id, name) WHERE parent_id IS NOT NULL;

CREATE TABLE tags (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT NOT NULL UNIQUE,
    color         TEXT,
    created_at    TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE transactions (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id        INTEGER NOT NULL REFERENCES accounts(id),
    category_id       INTEGER REFERENCES categories(id),
    type              TEXT NOT NULL,
    amount            INTEGER NOT NULL,
    currency          TEXT NOT NULL DEFAULT 'IDR',
    base_amount       INTEGER,
    base_currency     TEXT,
    description       TEXT,
    notes             TEXT,
    transfer_to_id    INTEGER REFERENCES accounts(id),
    date              TEXT NOT NULL,
    is_planned        INTEGER NOT NULL DEFAULT 0,
    planned_payment_id INTEGER REFERENCES planned_payments(id),
    created_at        TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at        TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE transaction_tags (
    transaction_id  INTEGER NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    tag_id          INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (transaction_id, tag_id)
);

CREATE TABLE budgets (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    name            TEXT,
    amount          INTEGER NOT NULL,
    currency        TEXT NOT NULL DEFAULT 'IDR',
    type            TEXT NOT NULL DEFAULT 'recurring',
    period_start    TEXT NOT NULL,
    period_end      TEXT NOT NULL,
    notify_at_pct   INTEGER DEFAULT 80,
    is_active       INTEGER NOT NULL DEFAULT 1,
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE budget_categories (
    budget_id    INTEGER NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    category_id  INTEGER NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (budget_id, category_id)
);

CREATE TABLE budget_tags (
    budget_id  INTEGER NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    tag_id     INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (budget_id, tag_id)
);

CREATE TABLE planned_payments (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id      INTEGER NOT NULL REFERENCES accounts(id),
    category_id     INTEGER REFERENCES categories(id),
    type            TEXT NOT NULL DEFAULT 'expense',
    amount          INTEGER NOT NULL,
    currency        TEXT NOT NULL DEFAULT 'IDR',
    name            TEXT NOT NULL,
    recurrence      TEXT NOT NULL DEFAULT 'none',
    recurrence_rule TEXT,
    start_date      TEXT NOT NULL,
    next_due_date   TEXT,
    is_paused       INTEGER NOT NULL DEFAULT 0,
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

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

INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Food & Dining', NULL, 'expense', '🍽️', 1, 0);
INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Transportation', NULL, 'expense', '🚌', 1, 1);
INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Shopping', NULL, 'expense', '🛍️', 1, 2);
INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Bills & Utilities', NULL, 'expense', '🧾', 1, 3);
INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Entertainment', NULL, 'expense', '🎬', 1, 4);
INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Health', NULL, 'expense', '💊', 1, 5);
INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Education', NULL, 'expense', '📚', 1, 6);
INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Income', NULL, 'income', '💰', 1, 7);

INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Restaurant', (SELECT id FROM categories WHERE name = 'Food & Dining' AND parent_id IS NULL), 'expense', '🍽️', 1, 0);
INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Groceries', (SELECT id FROM categories WHERE name = 'Food & Dining' AND parent_id IS NULL), 'expense', '🛒', 1, 1);
INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Coffee & Snacks', (SELECT id FROM categories WHERE name = 'Food & Dining' AND parent_id IS NULL), 'expense', '☕', 1, 2);

INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Fuel', (SELECT id FROM categories WHERE name = 'Transportation' AND parent_id IS NULL), 'expense', '⛽', 1, 0);
INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Public Transit', (SELECT id FROM categories WHERE name = 'Transportation' AND parent_id IS NULL), 'expense', '🚌', 1, 1);
INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Ride-hailing', (SELECT id FROM categories WHERE name = 'Transportation' AND parent_id IS NULL), 'expense', '🚕', 1, 2);

INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Clothing', (SELECT id FROM categories WHERE name = 'Shopping' AND parent_id IS NULL), 'expense', '👕', 1, 0);
INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Electronics', (SELECT id FROM categories WHERE name = 'Shopping' AND parent_id IS NULL), 'expense', '📱', 1, 1);
INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Household', (SELECT id FROM categories WHERE name = 'Shopping' AND parent_id IS NULL), 'expense', '🏠', 1, 2);

INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Electricity', (SELECT id FROM categories WHERE name = 'Bills & Utilities' AND parent_id IS NULL), 'expense', '⚡', 1, 0);
INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Internet', (SELECT id FROM categories WHERE name = 'Bills & Utilities' AND parent_id IS NULL), 'expense', '🌐', 1, 1);
INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Phone', (SELECT id FROM categories WHERE name = 'Bills & Utilities' AND parent_id IS NULL), 'expense', '📞', 1, 2);
INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Subscriptions', (SELECT id FROM categories WHERE name = 'Bills & Utilities' AND parent_id IS NULL), 'expense', '🔁', 1, 3);

INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Movies & Shows', (SELECT id FROM categories WHERE name = 'Entertainment' AND parent_id IS NULL), 'expense', '🎬', 1, 0);
INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Gaming', (SELECT id FROM categories WHERE name = 'Entertainment' AND parent_id IS NULL), 'expense', '🎮', 1, 1);
INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Travel', (SELECT id FROM categories WHERE name = 'Entertainment' AND parent_id IS NULL), 'expense', '✈️', 1, 2);

INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Medical', (SELECT id FROM categories WHERE name = 'Health' AND parent_id IS NULL), 'expense', '💊', 1, 0);
INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Fitness', (SELECT id FROM categories WHERE name = 'Health' AND parent_id IS NULL), 'expense', '🏋️', 1, 1);

INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Courses', (SELECT id FROM categories WHERE name = 'Education' AND parent_id IS NULL), 'expense', '📚', 1, 0);
INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Books', (SELECT id FROM categories WHERE name = 'Education' AND parent_id IS NULL), 'expense', '📖', 1, 1);

INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Salary', (SELECT id FROM categories WHERE name = 'Income' AND parent_id IS NULL), 'income', '💰', 1, 0);
INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Freelance', (SELECT id FROM categories WHERE name = 'Income' AND parent_id IS NULL), 'income', '💻', 1, 1);
INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Investment', (SELECT id FROM categories WHERE name = 'Income' AND parent_id IS NULL), 'income', '📈', 1, 2);
INSERT OR IGNORE INTO categories (name, parent_id, type, icon, is_system, sort_order) VALUES ('Other', (SELECT id FROM categories WHERE name = 'Income' AND parent_id IS NULL), 'income', '💵', 1, 3);
