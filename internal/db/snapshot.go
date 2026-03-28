package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ashwinpatri/debtCLI/internal/models"
)

// WriteSnapshot persists a complete snapshot and all its debt items in a
// single transaction. If anything fails the entire write is rolled back —
// a half-written snapshot is worse than no snapshot.
func WriteSnapshot(db *sql.DB, snap *models.Snapshot) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	result, err := tx.Exec(
		`INSERT INTO snapshots (repo_path, timestamp, health_score, item_count)
		 VALUES (?, ?, ?, ?)`,
		snap.RepoPath,
		snap.Timestamp.UTC().Format(time.RFC3339),
		snap.HealthScore,
		snap.ItemCount,
	)
	if err != nil {
		return fmt.Errorf("insert snapshot: %w", err)
	}

	snapID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}
	snap.ID = snapID

	stmt, err := tx.Prepare(
		`INSERT INTO debt_items
		   (snapshot_id, file, line, tag, comment, author, author_email, date, churn, score)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
	)
	if err != nil {
		return fmt.Errorf("prepare item insert: %w", err)
	}
	defer stmt.Close()

	for _, item := range snap.Items {
		_, err := stmt.Exec(
			snapID,
			item.File,
			item.Line,
			item.Tag,
			item.Comment,
			item.Author,
			item.AuthorEmail,
			item.Date.UTC().Format(time.RFC3339),
			item.Churn,
			item.Score,
		)
		if err != nil {
			return fmt.Errorf("insert debt item: %w", err)
		}
	}

	return tx.Commit()
}

// LoadLastSnapshot retrieves the most recent snapshot for repoPath, including
// all its debt items. Returns nil, nil if no snapshot exists for this repo.
func LoadLastSnapshot(db *sql.DB, repoPath string) (*models.Snapshot, error) {
	row := db.QueryRow(
		`SELECT id, repo_path, timestamp, health_score, item_count
		 FROM snapshots
		 WHERE repo_path = ?
		 ORDER BY timestamp DESC
		 LIMIT 1`,
		repoPath,
	)

	snap, err := scanSnapshot(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("load last snapshot: %w", err)
	}

	snap.Items, err = loadItems(db, snap.ID)
	if err != nil {
		return nil, err
	}

	return snap, nil
}

// LoadHistory returns all snapshots for repoPath ordered oldest-first,
// without loading individual debt items (used for the history chart).
func LoadHistory(db *sql.DB, repoPath string) ([]*models.Snapshot, error) {
	rows, err := db.Query(
		`SELECT id, repo_path, timestamp, health_score, item_count
		 FROM snapshots
		 WHERE repo_path = ?
		 ORDER BY timestamp ASC`,
		repoPath,
	)
	if err != nil {
		return nil, fmt.Errorf("query history: %w", err)
	}
	defer rows.Close()

	var snaps []*models.Snapshot
	for rows.Next() {
		snap, err := scanSnapshot(rows)
		if err != nil {
			return nil, fmt.Errorf("scan snapshot row: %w", err)
		}
		snaps = append(snaps, snap)
	}
	return snaps, rows.Err()
}

// scanSnapshot reads a snapshot row from either a *sql.Row or *sql.Rows.
// It accepts the common interface to avoid duplicating scan logic.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanSnapshot(r rowScanner) (*models.Snapshot, error) {
	var snap models.Snapshot
	var ts string
	if err := r.Scan(&snap.ID, &snap.RepoPath, &ts, &snap.HealthScore, &snap.ItemCount); err != nil {
		return nil, err
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return nil, fmt.Errorf("parse snapshot timestamp: %w", err)
	}
	snap.Timestamp = t
	return &snap, nil
}

// loadItems fetches all debt items belonging to snapshotID.
func loadItems(db *sql.DB, snapshotID int64) ([]models.DebtItem, error) {
	rows, err := db.Query(
		`SELECT file, line, tag, comment, author, author_email, date, churn, score
		 FROM debt_items
		 WHERE snapshot_id = ?
		 ORDER BY score DESC`,
		snapshotID,
	)
	if err != nil {
		return nil, fmt.Errorf("query debt items: %w", err)
	}
	defer rows.Close()

	var items []models.DebtItem
	for rows.Next() {
		var item models.DebtItem
		var dateStr string
		if err := rows.Scan(
			&item.File, &item.Line, &item.Tag, &item.Comment,
			&item.Author, &item.AuthorEmail, &dateStr,
			&item.Churn, &item.Score,
		); err != nil {
			return nil, fmt.Errorf("scan debt item: %w", err)
		}
		t, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			return nil, fmt.Errorf("parse item date: %w", err)
		}
		item.Date = t
		items = append(items, item)
	}
	return items, rows.Err()
}
