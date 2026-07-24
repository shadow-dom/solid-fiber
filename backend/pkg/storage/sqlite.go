// Package storage provides datastore connection helpers.
package storage

import (
	"database/sql"
	"fmt"

	// Pure-Go SQLite driver (no cgo); registers the "sqlite" driver.
	_ "modernc.org/sqlite"
)

// OpenSQLite opens (creating if necessary) a SQLite database at path and
// verifies the connection. WAL journaling and a busy timeout are enabled so
// concurrent readers don't block on a writer, and a short lock wait avoids
// spurious "database is locked" errors under light contention.
func OpenSQLite(path string) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)",
		path,
	)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %q: %w", path, err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite %q: %w", path, err)
	}
	return db, nil
}
