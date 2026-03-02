package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var singleton *sql.DB

// Init opens (or creates) the SQLite database at dataDir/vibecodepc.db,
// enables WAL mode, and runs schema migrations.
func Init(dataDir string) error {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return fmt.Errorf("db: create data dir: %w", err)
	}

	dbPath := filepath.Join(dataDir, "vibecodepc.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("db: open: %w", err)
	}

	// SQLite performs best with a single writer connection.
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		return fmt.Errorf("db: ping: %w", err)
	}

	// Enable WAL mode and set busy timeout.
	pragmas := []string{
		`PRAGMA journal_mode=WAL`,
		`PRAGMA busy_timeout=5000`,
		`PRAGMA foreign_keys=ON`,
		`PRAGMA synchronous=NORMAL`,
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return fmt.Errorf("db: pragma %q: %w", p, err)
		}
	}

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("db: migrate schema: %w", err)
	}

	if _, err := db.Exec(seedSQL); err != nil {
		return fmt.Errorf("db: seed: %w", err)
	}

	singleton = db
	return nil
}

// DB returns the singleton *sql.DB. Panics if Init has not been called.
func DB() *sql.DB {
	if singleton == nil {
		panic("db: not initialised — call db.Init() first")
	}
	return singleton
}
