package testdb

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
)

func TestOpen(t *testing.T) {
	database := Open(t)

	var count int
	if err := database.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count); err != nil {
		t.Fatalf("query categories: %v", err)
	}
	if count != 32 {
		t.Errorf("expected 32 seed categories, got %d", count)
	}

	if err := database.Ping(); err != nil {
		t.Fatalf("ping failed: %v", err)
	}
}

func TestOpenFile(t *testing.T) {
	database, cleanup := OpenFile(t)
	defer cleanup()

	var count int
	if err := database.QueryRow("SELECT COUNT(*) FROM accounts").Scan(&count); err != nil {
		t.Fatalf("query accounts: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 accounts, got %d", count)
	}

	if err := database.Ping(); err != nil {
		t.Fatalf("ping: %v", err)
	}
}

func TestOpenFileCleanupRemovesFile(t *testing.T) {
	database, cleanup := OpenFile(t)

	var x int
	if err := database.QueryRow("SELECT 1").Scan(&x); err != nil {
		t.Fatalf("query: %v", err)
	}

	cleanup()

	if err := database.Ping(); err == nil {
		t.Error("expected error after cleanup closes database")
	}
}

func TestOpenWithQuery(t *testing.T) {
	database := Open(t)

	_, err := database.Exec("INSERT INTO accounts (name, type, currency) VALUES ('Test', 'checking', 'IDR')")
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	var name string
	if err := database.QueryRow("SELECT name FROM accounts WHERE id = 1").Scan(&name); err != nil {
		t.Fatalf("query: %v", err)
	}
	if name != "Test" {
		t.Errorf("expected 'Test', got '%s'", name)
	}
}

func TestOpen_OpenDBError(t *testing.T) {
	oldOpen := openDB
	defer func() { openDB = oldOpen }()
	openDB = func(path string) (*sql.DB, error) {
		return nil, fmt.Errorf("mock open failure")
	}

	fakeT := &failingT{}
	func() {
		defer func() { _ = recover() }()
		Open(fakeT)
	}()
	if !fakeT.failed {
		t.Error("expected test to fail")
	}
}

func TestOpen_MigrateError(t *testing.T) {
	oldMigrate := migrateDB
	defer func() { migrateDB = oldMigrate }()
	migrateDB = func(db *sql.DB) error {
		return fmt.Errorf("mock migrate failure")
	}

	fakeT := &failingT{}
	func() {
		defer func() { _ = recover() }()
		Open(fakeT)
	}()
	if !fakeT.failed {
		t.Error("expected test to fail")
	}
}

func TestOpenFile_CreateTempError(t *testing.T) {
	oldCreate := createTemp
	defer func() { createTemp = oldCreate }()
	createTemp = func(dir, pattern string) (*os.File, error) {
		return nil, fmt.Errorf("mock create temp failure")
	}

	fakeT := &failingT{}
	func() {
		defer func() { _ = recover() }()
		OpenFile(fakeT)
	}()
	if !fakeT.failed {
		t.Error("expected test to fail")
	}
}

func TestOpenFile_OpenDBError(t *testing.T) {
	oldCreate := createTemp
	oldOpen := openDB
	defer func() {
		createTemp = oldCreate
		openDB = oldOpen
	}()
	createTemp = func(dir, pattern string) (*os.File, error) {
		return os.CreateTemp("", "wallet-test-*.db")
	}
	openDB = func(path string) (*sql.DB, error) {
		return nil, fmt.Errorf("mock open file failure")
	}

	fakeT := &failingT{}
	func() {
		defer func() { _ = recover() }()
		OpenFile(fakeT)
	}()
	if !fakeT.failed {
		t.Error("expected test to fail")
	}
}

func TestOpenFile_MigrateError(t *testing.T) {
	oldMigrate := migrateDB
	defer func() { migrateDB = oldMigrate }()
	migrateDB = func(db *sql.DB) error {
		return fmt.Errorf("mock migrate file failure")
	}

	fakeT := &failingT{}
	func() {
		defer func() { _ = recover() }()
		OpenFile(fakeT)
	}()
	if !fakeT.failed {
		t.Error("expected test to fail")
	}
}

type failingT struct {
	testing.TB
	failed bool
}

func (f *failingT) Fatalf(format string, args ...interface{}) {
	f.failed = true
	panic("test failure")
}

func (f *failingT) Cleanup(func()) {}
func (f *failingT) Helper()        {}
