package db

import (
	"database/sql"
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func openDiskDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	f, err := os.CreateTemp("", "wallet-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	path := f.Name()
	_ = f.Close()

	db, err := Open(path)
	if err != nil {
		_ = os.Remove(path)
		t.Fatalf("Open(file) error = %v", err)
	}
	cleanup := func() {
		_ = db.Close()
		_ = os.Remove(path)
	}
	return db, cleanup
}

func testMigrateFS(db *sql.DB, fsys fs.FS) error {
	return migrateFS(db, fsys)
}

func testPrepareSchemaVersion(db *sql.DB) (int, error) {
	return prepareSchemaVersion(db)
}

func testExecMigration(db *sql.DB, content []byte) error {
	return execMigration(db, content)
}

func TestOpenMemory(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open(:memory:) error = %v", err)
	}
	defer func() { _ = db.Close() }()

	var enabled int
	err = db.QueryRow("PRAGMA foreign_keys").Scan(&enabled)
	if err != nil {
		t.Fatalf("failed to check foreign_keys: %v", err)
	}
	if enabled != 1 {
		t.Errorf("foreign_keys should be enabled, got %d", enabled)
	}
}

func TestOpenFileDB(t *testing.T) {
	db, cleanup := openDiskDB(t)
	defer cleanup()

	var enabled int
	err := db.QueryRow("PRAGMA foreign_keys").Scan(&enabled)
	if err != nil {
		t.Fatalf("failed to check foreign_keys: %v", err)
	}
	if enabled != 1 {
		t.Errorf("foreign_keys should be enabled, got %d", enabled)
	}
}

func TestOpenInvalidPath(t *testing.T) {
	db, err := Open("/\x00invalid/\x00test.db")
	if err == nil {
		_ = db.Close()
		t.Fatal("expected error for invalid path")
	}
	if !strings.Contains(err.Error(), "enable foreign keys") {
		t.Errorf("expected 'enable foreign keys' error, got: %v", err)
	}
}

func TestDSNForPathMemory(t *testing.T) {
	dsn := dsnForPath(":memory:")
	if dsn != "file::memory:" {
		t.Errorf("expected 'file::memory:', got '%s'", dsn)
	}
}

func TestDSNForPathFile(t *testing.T) {
	dsn := dsnForPath("/tmp/test.db")
	if dsn != "file:/tmp/test.db?_journal_mode=WAL" {
		t.Errorf("expected 'file:/tmp/test.db?_journal_mode=WAL', got '%s'", dsn)
	}
}

func TestMigrateCreatesAllTables(t *testing.T) {
	db := openTestDB(t)

	err := Migrate(db)
	if err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	expectedTables := []string{
		"accounts", "categories", "tags", "transactions",
		"transaction_tags", "budgets", "budget_categories",
		"budget_tags", "planned_payments", "exchange_rates",
	}

	for _, table := range expectedTables {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		if err != nil {
			t.Errorf("failed to check table %s: %v", table, err)
			continue
		}
		if count != 1 {
			t.Errorf("expected table %s to exist", table)
		}
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	db := openTestDB(t)

	err := Migrate(db)
	if err != nil {
		t.Fatalf("first Migrate() error = %v", err)
	}

	var categoryCount int
	err = db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&categoryCount)
	if err != nil {
		t.Fatalf("failed to count categories: %v", err)
	}

	err = Migrate(db)
	if err != nil {
		t.Fatalf("second Migrate() error = %v", err)
	}

	var categoryCountAfter int
	err = db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&categoryCountAfter)
	if err != nil {
		t.Fatalf("failed to count categories after second migrate: %v", err)
	}

	if categoryCount != categoryCountAfter {
		t.Errorf("category count changed after second migrate: %d -> %d", categoryCount, categoryCountAfter)
	}

	parentCount := 8
	childCount := 24
	expectedTotal := parentCount + childCount
	if categoryCount != expectedTotal {
		t.Errorf("expected %d categories, got %d", expectedTotal, categoryCount)
	}
}

func TestCategoriesAreMarkedSystem(t *testing.T) {
	db := openTestDB(t)

	err := Migrate(db)
	if err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	var nonSystemCount int
	err = db.QueryRow("SELECT COUNT(*) FROM categories WHERE is_system = 0").Scan(&nonSystemCount)
	if err != nil {
		t.Fatalf("failed to count non-system categories: %v", err)
	}
	if nonSystemCount != 0 {
		t.Errorf("expected all seeded categories to be system, got %d non-system", nonSystemCount)
	}

	var systemCount int
	err = db.QueryRow("SELECT COUNT(*) FROM categories WHERE is_system = 1").Scan(&systemCount)
	if err != nil {
		t.Fatalf("failed to count system categories: %v", err)
	}
	if systemCount != 32 {
		t.Errorf("expected 32 system categories, got %d", systemCount)
	}
}

func TestForeignKeyEnforcement(t *testing.T) {
	db := openTestDB(t)

	err := Migrate(db)
	if err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	_, err = db.Exec("INSERT INTO transactions (account_id, type, amount, currency, date) VALUES (999, 'expense', 10000, 'IDR', '2026-01-01')")
	if err == nil {
		t.Error("expected foreign key error when inserting transaction with nonexistent account")
	}
}

func TestTransferTransaction(t *testing.T) {
	db := openTestDB(t)

	err := Migrate(db)
	if err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	_, err = db.Exec("INSERT INTO accounts (name, type, currency) VALUES ('BCA Checking', 'checking', 'IDR')")
	if err != nil {
		t.Fatalf("failed to insert source account: %v", err)
	}

	_, err = db.Exec("INSERT INTO accounts (name, type, currency) VALUES ('GoPay', 'ewallet', 'IDR')")
	if err != nil {
		t.Fatalf("failed to insert destination account: %v", err)
	}

	var categoryID int
	err = db.QueryRow("SELECT id FROM categories WHERE name = 'Food & Dining' AND parent_id IS NULL").Scan(&categoryID)
	if err != nil {
		t.Fatalf("failed to get category: %v", err)
	}

	_, err = db.Exec(
		"INSERT INTO transactions (account_id, category_id, type, amount, currency, description, transfer_to_id, date) VALUES (1, ?, 'transfer', 100000, 'IDR', 'Transfer to GoPay', 2, '2026-07-01')",
		categoryID,
	)
	if err != nil {
		t.Fatalf("failed to insert transfer transaction: %v", err)
	}

	var txnType string
	var txnTransferToID int
	err = db.QueryRow("SELECT type, transfer_to_id FROM transactions WHERE id = 1").Scan(&txnType, &txnTransferToID)
	if err != nil {
		t.Fatalf("failed to query transfer transaction: %v", err)
	}
	if txnType != "transfer" {
		t.Errorf("expected type 'transfer', got '%s'", txnType)
	}
	if txnTransferToID != 2 {
		t.Errorf("expected transfer_to_id 2, got %d", txnTransferToID)
	}
}

func TestTransactionTags(t *testing.T) {
	db := openTestDB(t)

	err := Migrate(db)
	if err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	_, err = db.Exec("INSERT INTO accounts (name, type, currency) VALUES ('BCA Checking', 'checking', 'IDR')")
	if err != nil {
		t.Fatalf("failed to insert account: %v", err)
	}

	var categoryID int
	err = db.QueryRow("SELECT id FROM categories WHERE name = 'Food & Dining' AND parent_id IS NULL").Scan(&categoryID)
	if err != nil {
		t.Fatalf("failed to get category: %v", err)
	}

	_, err = db.Exec(
		"INSERT INTO transactions (account_id, category_id, type, amount, currency, description, date) VALUES (1, ?, 'expense', 50000, 'IDR', 'Lunch', '2026-07-01')",
		categoryID,
	)
	if err != nil {
		t.Fatalf("failed to insert transaction: %v", err)
	}

	_, err = db.Exec("INSERT INTO tags (name) VALUES ('lunch')")
	if err != nil {
		t.Fatalf("failed to insert tag: %v", err)
	}

	_, err = db.Exec("INSERT INTO tags (name) VALUES ('work-meal')")
	if err != nil {
		t.Fatalf("failed to insert second tag: %v", err)
	}

	_, err = db.Exec("INSERT INTO transaction_tags (transaction_id, tag_id) VALUES (1, 1)")
	if err != nil {
		t.Fatalf("failed to insert transaction_tag: %v", err)
	}

	_, err = db.Exec("INSERT INTO transaction_tags (transaction_id, tag_id) VALUES (1, 2)")
	if err != nil {
		t.Fatalf("failed to insert second transaction_tag: %v", err)
	}

	var tagCount int
	err = db.QueryRow("SELECT COUNT(*) FROM transaction_tags WHERE transaction_id = 1").Scan(&tagCount)
	if err != nil {
		t.Fatalf("failed to count transaction tags: %v", err)
	}
	if tagCount != 2 {
		t.Errorf("expected 2 transaction tags, got %d", tagCount)
	}

	_, err = db.Exec("DELETE FROM transactions WHERE id = 1")
	if err != nil {
		t.Fatalf("failed to delete transaction: %v", err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM transaction_tags WHERE transaction_id = 1").Scan(&tagCount)
	if err != nil {
		t.Fatalf("failed to count transaction tags after delete: %v", err)
	}
	if tagCount != 0 {
		t.Errorf("expected 0 transaction tags after cascade delete, got %d", tagCount)
	}
}

func TestBudgetCategoriesAndTags(t *testing.T) {
	db := openTestDB(t)

	err := Migrate(db)
	if err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	_, err = db.Exec("INSERT INTO budgets (name, amount, currency, type, period_start, period_end) VALUES ('Monthly Food', 2000000, 'IDR', 'recurring', '2026-07-01', '2026-07-31')")
	if err != nil {
		t.Fatalf("failed to insert budget: %v", err)
	}

	var foodID int
	err = db.QueryRow("SELECT id FROM categories WHERE name = 'Food & Dining' AND parent_id IS NULL").Scan(&foodID)
	if err != nil {
		t.Fatalf("failed to get food category: %v", err)
	}

	var transportID int
	err = db.QueryRow("SELECT id FROM categories WHERE name = 'Transportation' AND parent_id IS NULL").Scan(&transportID)
	if err != nil {
		t.Fatalf("failed to get transport category: %v", err)
	}

	_, err = db.Exec("INSERT INTO budget_categories (budget_id, category_id) VALUES (1, ?)", foodID)
	if err != nil {
		t.Fatalf("failed to insert budget_category: %v", err)
	}

	_, err = db.Exec("INSERT INTO budget_categories (budget_id, category_id) VALUES (1, ?)", transportID)
	if err != nil {
		t.Fatalf("failed to insert second budget_category: %v", err)
	}

	_, err = db.Exec("INSERT INTO tags (name) VALUES ('essential')")
	if err != nil {
		t.Fatalf("failed to insert tag: %v", err)
	}

	_, err = db.Exec("INSERT INTO budget_tags (budget_id, tag_id) VALUES (1, 1)")
	if err != nil {
		t.Fatalf("failed to insert budget_tag: %v", err)
	}

	var catCount int
	err = db.QueryRow("SELECT COUNT(*) FROM budget_categories WHERE budget_id = 1").Scan(&catCount)
	if err != nil {
		t.Fatalf("failed to count budget categories: %v", err)
	}
	if catCount != 2 {
		t.Errorf("expected 2 budget categories, got %d", catCount)
	}

	var tagCount int
	err = db.QueryRow("SELECT COUNT(*) FROM budget_tags WHERE budget_id = 1").Scan(&tagCount)
	if err != nil {
		t.Fatalf("failed to count budget tags: %v", err)
	}
	if tagCount != 1 {
		t.Errorf("expected 1 budget tag, got %d", tagCount)
	}

	_, err = db.Exec("DELETE FROM budgets WHERE id = 1")
	if err != nil {
		t.Fatalf("failed to delete budget: %v", err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM budget_categories WHERE budget_id = 1").Scan(&catCount)
	if err != nil {
		t.Fatalf("failed to count budget categories after delete: %v", err)
	}
	if catCount != 0 {
		t.Errorf("expected 0 budget categories after cascade delete, got %d", catCount)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM budget_tags WHERE budget_id = 1").Scan(&tagCount)
	if err != nil {
		t.Fatalf("failed to count budget tags after delete: %v", err)
	}
	if tagCount != 0 {
		t.Errorf("expected 0 budget tags after cascade delete, got %d", tagCount)
	}
}

func TestTagUniqueness(t *testing.T) {
	db := openTestDB(t)

	err := Migrate(db)
	if err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	_, err = db.Exec("INSERT INTO tags (name) VALUES ('vacation')")
	if err != nil {
		t.Fatalf("failed to insert tag: %v", err)
	}

	_, err = db.Exec("INSERT INTO tags (name) VALUES ('vacation')")
	if err == nil {
		t.Error("expected unique constraint error on duplicate tag name")
	}
}

func TestExchangeRateUniqueness(t *testing.T) {
	db := openTestDB(t)

	err := Migrate(db)
	if err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	_, err = db.Exec("INSERT INTO exchange_rates (from_currency, to_currency, rate) VALUES ('USD', 'IDR', 16200)")
	if err != nil {
		t.Fatalf("failed to insert exchange rate: %v", err)
	}

	_, err = db.Exec("INSERT INTO exchange_rates (from_currency, to_currency, rate) VALUES ('USD', 'IDR', 16200)")
	if err == nil {
		t.Error("expected unique constraint error on duplicate exchange rate")
	}
}

func TestMigrateFSReadDirError(t *testing.T) {
	db := openTestDB(t)

	err := testMigrateFS(db, fstest.MapFS{})
	if err == nil {
		t.Fatal("expected error from migrateFS with empty MapFS")
	}
	if !strings.Contains(err.Error(), "read migrations dir") {
		t.Errorf("expected 'read migrations dir' error, got: %v", err)
	}
}

func TestMigrateFSReadFileError(t *testing.T) {
	db := openTestDB(t)

	errFS := &readFileErrorFS{FS: migrationsFS}
	err := testMigrateFS(db, errFS)
	if err == nil {
		t.Fatal("expected error from migrateFS with failing ReadFile")
	}
	if !strings.Contains(err.Error(), "read migration") {
		t.Errorf("expected 'read migration' error, got: %v", err)
	}
}

func TestMigrateFSSscanfSkip(t *testing.T) {
	db := openTestDB(t)

	mapFS := fstest.MapFS{
		"migrations/README.sql":      &fstest.MapFile{Data: []byte{}},
		"migrations/001_initial.sql": &fstest.MapFile{Data: []byte(`CREATE TABLE IF NOT EXISTS schema_version (version INTEGER NOT NULL);`)},
	}
	err := testMigrateFS(db, mapFS)
	if err != nil {
		t.Fatalf("migrateFS with non-numeric filename error = %v", err)
	}
}

func TestMigrateFSBadSQL(t *testing.T) {
	db := openTestDB(t)

	mapFS := fstest.MapFS{
		"migrations/001_bad.sql": &fstest.MapFile{Data: []byte("THIS IS NOT VALID SQL;")},
	}
	err := testMigrateFS(db, mapFS)
	if err == nil {
		t.Fatal("expected error from migrateFS with bad SQL")
	}
	if !strings.Contains(err.Error(), "execute migration") {
		t.Errorf("expected 'execute migration' error, got: %v", err)
	}
}

func TestMigrateFSClosedDB(t *testing.T) {
	db := openTestDB(t)
	_ = db.Close()

	err := testMigrateFS(db, migrationsFS)
	if err == nil {
		t.Fatal("expected error from migrateFS on closed DB")
	}
}

func TestPrepareSchemaVersionClosedDB(t *testing.T) {
	db := openTestDB(t)
	_ = db.Close()

	_, err := testPrepareSchemaVersion(db)
	if err == nil {
		t.Fatal("expected error from prepareSchemaVersion on closed DB")
	}
	if !strings.Contains(err.Error(), "create schema_version") {
		t.Errorf("expected 'create schema_version' error, got: %v", err)
	}
}

func TestExecMigrationClosedDB(t *testing.T) {
	db := openTestDB(t)
	_ = db.Close()

	err := testExecMigration(db, []byte("SELECT 1"))
	if err == nil {
		t.Fatal("expected error from execMigration on closed DB")
	}
}

type readFileErrorFS struct {
	fs.FS
}

func (f *readFileErrorFS) Open(name string) (fs.File, error) {
	if strings.HasSuffix(name, ".sql") {
		return nil, errors.New("simulated read error")
	}
	return f.FS.Open(name)
}
