package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func Open(path string) (*sql.DB, error) {
	dsn := path + "?_journal_mode=WAL&_foreign_keys=on"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	return db, nil
}

func Migrate(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_version (version INTEGER NOT NULL)`)
	if err != nil {
		return fmt.Errorf("create schema_version: %w", err)
	}

	var currentVersion int
	err = db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("get current version: %w", err)
	}

	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		var version int
		_, err := fmt.Sscanf(entry.Name(), "%d_", &version)
		if err != nil {
			continue
		}

		if version > currentVersion {
			content, err := migrationsFS.ReadFile("migrations/" + entry.Name())
			if err != nil {
				return fmt.Errorf("read migration %s: %w", entry.Name(), err)
			}

			_, err = db.Exec(string(content))
			if err != nil {
				return fmt.Errorf("execute migration %s: %w", entry.Name(), err)
			}

			_, err = db.Exec("INSERT OR REPLACE INTO schema_version (version) VALUES (?)", version)
			if err != nil {
				return fmt.Errorf("update schema_version: %w", err)
			}
		}
	}

	return nil
}
