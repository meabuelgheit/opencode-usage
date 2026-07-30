package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// Open opens the opencode SQLite database in read-only mode.
// The dbPath can be a path or "~" shortcut. Returns the *sql.DB handle.
func Open(dbPath string) (*sql.DB, error) {
	// Expand ~ to home directory
	if strings.HasPrefix(dbPath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("cannot find home directory: %w", err)
		}
		dbPath = filepath.Join(home, dbPath[2:])
	}

	// Check if file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("database not found at %s", dbPath)
	}

	// Open in read-only mode with WAL support
	// Use mode=ro to open read-only, and add _query_only to avoid WAL issues
	dsn := dbPath + "?mode=ro&_journal_mode=WAL&_query_only=true"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}
