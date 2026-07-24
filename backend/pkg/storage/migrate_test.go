package storage_test

import (
	"path/filepath"
	"testing"

	"github.com/shadow-dom/solid-fiber/pkg/storage"
)

func TestMigrate_AppliesAndIsIdempotent(t *testing.T) {
	db, err := storage.OpenSQLite(filepath.Join(t.TempDir(), "m.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	migs := []storage.Migration{
		{Version: 2, Name: "second", SQL: "CREATE TABLE t2(id INTEGER);"},
		{Version: 1, Name: "first", SQL: "CREATE TABLE t1(id INTEGER);"}, // intentionally out of order
	}

	if err := storage.Migrate(db, migs); err != nil {
		t.Fatalf("first migrate: %v", err)
	}
	// Running again must be a no-op, not an error.
	if err := storage.Migrate(db, migs); err != nil {
		t.Fatalf("second migrate (idempotent): %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 recorded migrations, got %d", count)
	}

	// Both tables should exist (ascending order applied despite input order).
	for _, table := range []string{"t1", "t2"} {
		var name string
		err := db.QueryRow(
			`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table,
		).Scan(&name)
		if err != nil {
			t.Fatalf("expected table %q to exist: %v", table, err)
		}
	}
}

func TestMigrate_AppliesOnlyNewVersions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "m.db")
	db, err := storage.OpenSQLite(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	if err := storage.Migrate(db, []storage.Migration{
		{Version: 1, Name: "first", SQL: "CREATE TABLE t1(id INTEGER);"},
	}); err != nil {
		t.Fatalf("migrate v1: %v", err)
	}
	// Add a v2 on top; only v2 should apply.
	if err := storage.Migrate(db, []storage.Migration{
		{Version: 1, Name: "first", SQL: "CREATE TABLE t1(id INTEGER);"},
		{Version: 2, Name: "second", SQL: "CREATE TABLE t2(id INTEGER);"},
	}); err != nil {
		t.Fatalf("migrate v2: %v", err)
	}

	var max int
	if err := db.QueryRow(`SELECT MAX(version) FROM schema_migrations`).Scan(&max); err != nil {
		t.Fatalf("max version: %v", err)
	}
	if max != 2 {
		t.Fatalf("expected schema at version 2, got %d", max)
	}
}
