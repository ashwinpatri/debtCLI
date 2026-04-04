package db

import (
	"database/sql"
	"fmt"
)

// migrations is append-only — never edit or reorder existing entries.
var migrations = []string{
	`CREATE TABLE IF NOT EXISTS schema_version (
		version INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS snapshots (
		id           INTEGER PRIMARY KEY AUTOINCREMENT,
		repo_path    TEXT    NOT NULL,
		timestamp    DATETIME NOT NULL,
		health_score REAL    NOT NULL,
		item_count   INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS debt_items (
		id           INTEGER PRIMARY KEY AUTOINCREMENT,
		snapshot_id  INTEGER NOT NULL REFERENCES snapshots(id) ON DELETE CASCADE,
		file         TEXT    NOT NULL,
		line         INTEGER NOT NULL,
		tag          TEXT    NOT NULL,
		comment      TEXT    NOT NULL,
		author       TEXT    NOT NULL,
		author_email TEXT    NOT NULL,
		date         DATETIME NOT NULL,
		churn        INTEGER NOT NULL,
		score        REAL    NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_debt_items_snapshot ON debt_items(snapshot_id);
	CREATE INDEX IF NOT EXISTS idx_snapshots_repo ON snapshots(repo_path, timestamp);`,
}

func migrate(db *sql.DB) error {
	current, err := schemaVersion(db)
	if err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}

	for i := current; i < len(migrations); i++ {
		if err := applyMigration(db, i+1, migrations[i]); err != nil {
			return fmt.Errorf("apply migration %d: %w", i+1, err)
		}
	}
	return nil
}

func schemaVersion(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_version'`).Scan(&count)
	if err != nil {
		return 0, err
	}
	if count == 0 {
		return 0, nil
	}

	var version int
	err = db.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_version`).Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

func applyMigration(db *sql.DB, version int, sql string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.Exec(sql); err != nil {
		return err
	}
	if _, err := tx.Exec(`INSERT INTO schema_version(version) VALUES (?)`, version); err != nil {
		return err
	}
	return tx.Commit()
}
