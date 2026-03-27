// Package db manages the SQLite database that stores scan snapshots over time.
// It uses the pure-Go modernc.org/sqlite driver — no CGo, no external libraries.
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // register the sqlite driver
)

// Open creates or opens the SQLite database at path, applies all pending
// migrations, and sets connection pragmas optimised for a single-writer CLI.
//
// The parent directory is created with 0700 permissions. The file itself is
// set to 0600 after creation to prevent other local users from reading the
// history store.
func Open(path string) (*sql.DB, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	// Touch the file so we can chmod before handing it to SQLite.
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("create db file: %w", err)
	}
	f.Close()

	if err := os.Chmod(path, 0600); err != nil {
		return nil, fmt.Errorf("set db file permissions: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := applyPragmas(db); err != nil {
		db.Close()
		return nil, err
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate db: %w", err)
	}

	return db, nil
}

// applyPragmas sets connection-level SQLite settings.
// WAL mode allows readers and the single writer to proceed concurrently.
// The cache_size value is in kibibytes (negative sign convention).
func applyPragmas(db *sql.DB) error {
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA temp_store=MEMORY",
		"PRAGMA cache_size=-32000",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return fmt.Errorf("pragma %q: %w", p, err)
		}
	}
	return nil
}
