package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func Open(path string, logger *slog.Logger) (*sql.DB, error) {
	logger.Debug("opening database", slog.String("path", path))
	dsn := dsnForPath(path)
	db, _ := sql.Open("sqlite", dsn)
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}
	logger.Info("database opened", slog.String("journal_mode", "wal"))
	return db, nil
}

func dsnForPath(path string) string {
	if path == ":memory:" {
		return "file::memory:"
	}
	return "file:" + path + "?_journal_mode=WAL"
}

func Migrate(db *sql.DB, logger *slog.Logger) error {
	return migrateFS(db, migrationsFS, logger)
}

func migrateFS(db *sql.DB, fsys fs.FS, logger *slog.Logger) error {
	currentVersion, err := prepareSchemaVersion(db)
	if err != nil {
		return err
	}

	entries, err := fs.ReadDir(fsys, "migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		var version int
		if _, err := fmt.Sscanf(entry.Name(), "%d_", &version); err != nil {
			continue
		}
		if version <= currentVersion {
			logger.Debug("migration skipped", slog.String("file", entry.Name()), slog.Int("version", version))
			continue
		}

		content, err := fs.ReadFile(fsys, "migrations/"+entry.Name())
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}

		if err := execMigration(db, content); err != nil {
			return fmt.Errorf("execute migration %s: %w", entry.Name(), err)
		}

		setSchemaVersion(db, version)
		logger.Info("migration applied", slog.String("file", entry.Name()), slog.Int("version", version))
	}

	return nil
}

func prepareSchemaVersion(db *sql.DB) (int, error) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_version (version INTEGER NOT NULL)`)
	if err != nil {
		return 0, fmt.Errorf("create schema_version: %w", err)
	}

	var version int
	_ = db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&version)
	return version, nil
}

func execMigration(db *sql.DB, content []byte) error {
	_, err := db.Exec(string(content))
	return err
}

func setSchemaVersion(db *sql.DB, version int) {
	_, _ = db.Exec("INSERT OR REPLACE INTO schema_version (version) VALUES (?)", version)
}
