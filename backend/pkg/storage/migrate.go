package storage

import (
	"database/sql"
	"fmt"
	"sort"
	"time"
)

// Migration is a single, forward-only schema change within a namespace.
type Migration struct {
	Version int
	Name    string
	SQL     string
}

// Migrate applies the migrations for a namespace whose Version is greater than
// the highest already recorded for that namespace. Versions are scoped per
// namespace, so each resource numbers its migrations independently from 1. It
// is idempotent, and each migration runs in its own transaction.
func Migrate(db *sql.DB, namespace string, migrations []Migration) error {
	if err := ensureMigrationsTable(db); err != nil {
		return err
	}

	var current int
	if err := db.QueryRow(
		`SELECT COALESCE(MAX(version), 0) FROM schema_migrations WHERE namespace = ?`, namespace,
	).Scan(&current); err != nil {
		return fmt.Errorf("read schema version for %q: %w", namespace, err)
	}

	ordered := make([]Migration, len(migrations))
	copy(ordered, migrations)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Version < ordered[j].Version })

	for _, m := range ordered {
		if m.Version <= current {
			continue
		}
		if err := applyMigration(db, namespace, m); err != nil {
			return err
		}
	}
	return nil
}

func applyMigration(db *sql.DB, namespace string, m Migration) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin migration %s/%d (%s): %w", namespace, m.Version, m.Name, err)
	}
	if _, err := tx.Exec(m.SQL); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("apply migration %s/%d (%s): %w", namespace, m.Version, m.Name, err)
	}
	if _, err := tx.Exec(
		`INSERT INTO schema_migrations (namespace, version, name, applied_at) VALUES (?, ?, ?, ?)`,
		namespace, m.Version, m.Name, time.Now().Unix(),
	); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("record migration %s/%d (%s): %w", namespace, m.Version, m.Name, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s/%d (%s): %w", namespace, m.Version, m.Name, err)
	}
	return nil
}

func ensureMigrationsTable(db *sql.DB) error {
	var hasNamespace int
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM pragma_table_info('schema_migrations') WHERE name = 'namespace'`,
	).Scan(&hasNamespace); err != nil {
		return fmt.Errorf("inspect schema_migrations: %w", err)
	}
	// A pre-namespacing table only tracked bookkeeping; drop it so the namespaced
	// schema can be created. Resource tables persist and their CREATE ... IF NOT
	// EXISTS migrations simply re-run.
	if hasNamespace == 0 {
		if _, err := db.Exec(`DROP TABLE IF EXISTS schema_migrations`); err != nil {
			return fmt.Errorf("drop legacy schema_migrations: %w", err)
		}
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		namespace  TEXT    NOT NULL,
		version    INTEGER NOT NULL,
		name       TEXT    NOT NULL,
		applied_at INTEGER NOT NULL,
		PRIMARY KEY (namespace, version)
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}
	return nil
}
