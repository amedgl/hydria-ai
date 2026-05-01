// internal/tracker/tracker.go
// SQLite-based attempt tracker.
// No password is ever tried twice — not in the same session, not across sessions.
package tracker

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

// Init opens (or creates) the SQLite database and creates tables.
func Init(dbPath string) error {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("open sqlite: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS attempts (
			id        INTEGER PRIMARY KEY AUTOINCREMENT,
			target    TEXT    NOT NULL,
			username  TEXT    NOT NULL,
			password  TEXT    NOT NULL,
			tried_at  TEXT    NOT NULL,
			result    TEXT    NOT NULL DEFAULT 'failed'
		);

		CREATE TABLE IF NOT EXISTS sessions (
			session_id      TEXT PRIMARY KEY,
			target          TEXT NOT NULL,
			service         TEXT NOT NULL,
			username        TEXT NOT NULL,
			image_path      TEXT,
			wordlist_path   TEXT,
			created_at      TEXT NOT NULL,
			updated_at      TEXT NOT NULL,
			status          TEXT NOT NULL DEFAULT 'running',
			found_password  TEXT,
			tried_count     INTEGER DEFAULT 0,
			skipped_count   INTEGER DEFAULT 0
		);

		CREATE INDEX IF NOT EXISTS idx_attempts
			ON attempts (target, username, password);
	`)
	return err
}

// DB returns the underlying *sql.DB for use by other packages.
func DB() *sql.DB { return db }

// FilterUntried returns passwords not yet tried for target/username,
// and the number of passwords that were skipped.
func FilterUntried(target, username string, passwords []string) ([]string, int) {
	if len(passwords) == 0 {
		return nil, 0
	}

	placeholders := strings.Repeat("?,", len(passwords))
	placeholders = placeholders[:len(placeholders)-1]

	query := fmt.Sprintf(
		`SELECT password FROM attempts WHERE target=? AND username=? AND password IN (%s)`,
		placeholders,
	)

	args := make([]any, 0, len(passwords)+2)
	args = append(args, target, username)
	for _, p := range passwords {
		args = append(args, p)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return passwords, 0
	}
	defer rows.Close()

	tried := make(map[string]struct{})
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err == nil {
			tried[p] = struct{}{}
		}
	}

	var untried []string
	for _, p := range passwords {
		if _, found := tried[p]; !found {
			untried = append(untried, p)
		}
	}
	return untried, len(passwords) - len(untried)
}

// RecordAttempt records a single attempt.
func RecordAttempt(target, username, password, result string) {
	db.Exec(
		`INSERT INTO attempts (target, username, password, tried_at, result) VALUES (?,?,?,?,?)`,
		target, username, password, time.Now().Format(time.RFC3339), result,
	)
}

// RecordAttemptsBulk records many attempts in one transaction.
func RecordAttemptsBulk(target, username string, passwords []string, result string) {
	tx, err := db.Begin()
	if err != nil {
		return
	}
	stmt, err := tx.Prepare(
		`INSERT OR IGNORE INTO attempts (target, username, password, tried_at, result) VALUES (?,?,?,?,?)`,
	)
	if err != nil {
		tx.Rollback()
		return
	}
	defer stmt.Close()

	now := time.Now().Format(time.RFC3339)
	for _, p := range passwords {
		stmt.Exec(target, username, p, now, result)
	}
	tx.Commit()
}

// GetTriedCount returns how many attempts have been made against target/username.
func GetTriedCount(target, username string) int {
	var count int
	db.QueryRow(
		`SELECT COUNT(*) FROM attempts WHERE target=? AND username=?`,
		target, username,
	).Scan(&count)
	return count
}
