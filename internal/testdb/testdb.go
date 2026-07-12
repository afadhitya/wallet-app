package testdb

import (
	"database/sql"
	"log/slog"
	"os"
	"testing"

	"github.com/afadhitya/wallet-app/internal/db"
)

type TB interface {
	Helper()
	Fatalf(format string, args ...interface{})
	Cleanup(func())
}

var (
	openDB     = db.Open
	migrateDB  = db.Migrate
	createTemp = os.CreateTemp
	removeFile = os.Remove
)

func Open(t testing.TB, logger *slog.Logger) *sql.DB {
	t.Helper()

	database, err := openDB(":memory:", logger)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	if err := migrateDB(database, logger); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return database
}

func OpenFile(t testing.TB, logger *slog.Logger) (*sql.DB, func()) {
	t.Helper()

	f, err := createTemp("", "wallet-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	path := f.Name()
	_ = f.Close()

	database, err := openDB(path, logger)
	if err != nil {
		_ = removeFile(path)
		t.Fatalf("failed to open file database: %v", err)
	}

	if err := migrateDB(database, logger); err != nil {
		_ = database.Close()
		_ = removeFile(path)
		t.Fatalf("failed to migrate file database: %v", err)
	}

	cleanup := func() {
		_ = database.Close()
		_ = removeFile(path)
	}

	return database, cleanup
}
