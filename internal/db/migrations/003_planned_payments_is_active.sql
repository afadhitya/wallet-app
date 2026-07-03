ALTER TABLE planned_payments ADD COLUMN is_active INTEGER NOT NULL DEFAULT 1;

CREATE INDEX idx_planned_payments_account_id ON planned_payments(account_id);
CREATE INDEX idx_planned_payments_next_due_date ON planned_payments(next_due_date);
