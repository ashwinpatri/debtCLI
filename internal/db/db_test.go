package db

import (
	"testing"
	"time"

	"github.com/ashwinpatri/debtCLI/internal/models"
)

// openMemory opens an in-memory SQLite database with migrations applied.
// The cache=shared parameter allows multiple connections to share the same
// in-memory store, which is necessary for some test patterns.
func openMemory(t *testing.T) interface {
	Exec(query string, args ...any) (interface{ LastInsertId() (int64, error) }, error)
} {
	t.Helper()
	return nil
}

func testDB(t *testing.T) *testDatabase {
	t.Helper()
	db, err := Open("file::memory:?cache=shared&_fk=1")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return &testDatabase{db: db, t: t}
}

type testDatabase struct {
	db interface {
		Close() error
	}
	t *testing.T
}

func TestOpen_InMemory(t *testing.T) {
	db, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()
}

func TestMigrate_Idempotent(t *testing.T) {
	db, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	// Running migrate a second time must not fail.
	if err := migrate(db); err != nil {
		t.Errorf("second migrate: %v", err)
	}
}

func TestWriteAndLoadSnapshot(t *testing.T) {
	db, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC().Truncate(time.Second)
	snap := &models.Snapshot{
		RepoPath:    "/repo",
		Timestamp:   now,
		HealthScore: 72.5,
		ItemCount:   2,
		Items: []models.DebtItem{
			{
				File:        "main.go",
				Line:        10,
				Tag:         "TODO",
				Comment:     "fix this",
				Author:      "ashwin",
				AuthorEmail: "a@example.com",
				Date:        now,
				Churn:       3,
				Score:       4.2,
			},
			{
				File:    "api.go",
				Line:    55,
				Tag:     "FIXME",
				Comment: "off by one",
				Author:  "dan",
				Date:    now,
				Churn:   1,
				Score:   2.1,
			},
		},
	}

	if err := WriteSnapshot(db, snap); err != nil {
		t.Fatalf("WriteSnapshot: %v", err)
	}
	if snap.ID == 0 {
		t.Error("expected snap.ID to be set after write")
	}

	loaded, err := LoadLastSnapshot(db, "/repo")
	if err != nil {
		t.Fatalf("LoadLastSnapshot: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected snapshot, got nil")
	}
	if loaded.HealthScore != 72.5 {
		t.Errorf("health score: got %.1f", loaded.HealthScore)
	}
	if len(loaded.Items) != 2 {
		t.Errorf("items: got %d, want 2", len(loaded.Items))
	}
	if loaded.Items[0].Tag != "TODO" {
		t.Errorf("first item tag: got %s", loaded.Items[0].Tag)
	}
}

func TestLoadLastSnapshot_NoRows(t *testing.T) {
	db, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	snap, err := LoadLastSnapshot(db, "/nonexistent")
	if err != nil {
		t.Fatalf("LoadLastSnapshot: %v", err)
	}
	if snap != nil {
		t.Error("expected nil for unknown repo")
	}
}

func TestLoadHistory(t *testing.T) {
	db, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	base := time.Now().UTC().Truncate(time.Second)
	for i, score := range []float64{80.0, 75.0, 60.0} {
		snap := &models.Snapshot{
			RepoPath:    "/repo",
			Timestamp:   base.Add(time.Duration(i) * time.Minute),
			HealthScore: score,
		}
		if err := WriteSnapshot(db, snap); err != nil {
			t.Fatalf("WriteSnapshot %d: %v", i, err)
		}
	}

	history, err := LoadHistory(db, "/repo")
	if err != nil {
		t.Fatalf("LoadHistory: %v", err)
	}
	if len(history) != 3 {
		t.Fatalf("expected 3 snapshots, got %d", len(history))
	}
	if history[0].HealthScore != 80.0 {
		t.Errorf("first score: got %.1f, want 80.0", history[0].HealthScore)
	}
	if history[2].HealthScore != 60.0 {
		t.Errorf("last score: got %.1f, want 60.0", history[2].HealthScore)
	}
}
