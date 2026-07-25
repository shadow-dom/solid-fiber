package storage_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/shadow-dom/solid-fiber/pkg/storage"
)

func openDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := storage.OpenSQLite(filepath.Join(t.TempDir(), "m.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func hasTable(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()
	var found string
	err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, name).Scan(&found)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		t.Fatalf("table lookup %q: %v", name, err)
	}
	return true
}

func TestMigrate_AppliesAndIsIdempotent(t *testing.T) {
	db := openDB(t)
	migs := []storage.Migration{
		{Version: 2, Name: "second", SQL: "CREATE TABLE t2(id INTEGER);"},
		{Version: 1, Name: "first", SQL: "CREATE TABLE t1(id INTEGER);"},
	}

	if err := storage.Migrate(db, "core", migs); err != nil {
		t.Fatalf("first migrate: %v", err)
	}
	if err := storage.Migrate(db, "core", migs); err != nil {
		t.Fatalf("second migrate (idempotent): %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE namespace = ?`, "core").Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 recorded migrations, got %d", count)
	}
	for _, table := range []string{"t1", "t2"} {
		if !hasTable(t, db, table) {
			t.Fatalf("expected table %q to exist", table)
		}
	}
}

func TestMigrate_AppliesOnlyNewVersions(t *testing.T) {
	db := openDB(t)
	if err := storage.Migrate(db, "core", []storage.Migration{
		{Version: 1, Name: "first", SQL: "CREATE TABLE t1(id INTEGER);"},
	}); err != nil {
		t.Fatalf("migrate v1: %v", err)
	}
	if err := storage.Migrate(db, "core", []storage.Migration{
		{Version: 1, Name: "first", SQL: "CREATE TABLE t1(id INTEGER);"},
		{Version: 2, Name: "second", SQL: "CREATE TABLE t2(id INTEGER);"},
	}); err != nil {
		t.Fatalf("migrate v2: %v", err)
	}

	var max int
	if err := db.QueryRow(`SELECT MAX(version) FROM schema_migrations WHERE namespace = ?`, "core").Scan(&max); err != nil {
		t.Fatalf("max version: %v", err)
	}
	if max != 2 {
		t.Fatalf("expected schema at version 2, got %d", max)
	}
}

// A database created before namespacing has a schema_migrations table without a
// namespace column; Migrate must upgrade it without losing the resource tables.
func TestMigrate_UpgradesLegacyTrackingTable(t *testing.T) {
	db := openDB(t)
	if _, err := db.Exec(`CREATE TABLE schema_migrations (
		version INTEGER PRIMARY KEY, name TEXT NOT NULL, applied_at INTEGER NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO schema_migrations (version, name, applied_at) VALUES (1, 'legacy', 0)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE work_items (id TEXT PRIMARY KEY)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO work_items (id) VALUES ('keep-me')`); err != nil {
		t.Fatal(err)
	}

	if err := storage.Migrate(db, "work_items", []storage.Migration{
		{Version: 1, Name: "create_work_items", SQL: "CREATE TABLE IF NOT EXISTS work_items (id TEXT PRIMARY KEY);"},
	}); err != nil {
		t.Fatalf("migrate legacy db: %v", err)
	}

	var id string
	if err := db.QueryRow(`SELECT id FROM work_items`).Scan(&id); err != nil || id != "keep-me" {
		t.Fatalf("legacy data lost: id=%q err=%v", id, err)
	}
	var version int
	if err := db.QueryRow(`SELECT version FROM schema_migrations WHERE namespace = ?`, "work_items").Scan(&version); err != nil {
		t.Fatalf("expected namespaced record: %v", err)
	}
}

// Two resources each numbering their migrations from 1 must not collide: both
// version-1 migrations apply because versions are scoped per namespace.
func TestMigrate_NamespacesAreIndependent(t *testing.T) {
	db := openDB(t)
	if err := storage.Migrate(db, "work_items", []storage.Migration{
		{Version: 1, Name: "create_work_items", SQL: "CREATE TABLE work_items(id TEXT);"},
	}); err != nil {
		t.Fatalf("work_items migrate: %v", err)
	}
	if err := storage.Migrate(db, "tags", []storage.Migration{
		{Version: 1, Name: "create_tags", SQL: "CREATE TABLE tags(id TEXT);"},
	}); err != nil {
		t.Fatalf("tags migrate: %v", err)
	}

	for _, table := range []string{"work_items", "tags"} {
		if !hasTable(t, db, table) {
			t.Fatalf("expected table %q to exist (namespace collision?)", table)
		}
	}
}
