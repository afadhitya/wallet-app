package testdb

import (
	"database/sql"
	"os"
	"testing"

	"github.com/afadhitya/wallet-app/internal/db"
)

func Open(t *testing.T) *sql.DB {
	t.Helper()

	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	if err := db.Migrate(database); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return database
}

func OpenFile(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	f, err := os.CreateTemp("", "wallet-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	path := f.Name()
	_ = f.Close()

	database, err := db.Open(path)
	if err != nil {
		_ = os.Remove(path)
		t.Fatalf("failed to open file database: %v", err)
	}

	if err := db.Migrate(database); err != nil {
		_ = database.Close()
		_ = os.Remove(path)
		t.Fatalf("failed to migrate file database: %v", err)
	}

	cleanup := func() {
		_ = database.Close()
		_ = os.Remove(path)
	}

	return database, cleanup
}
