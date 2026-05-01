// internal/session/session.go
// Session management — every attack is tied to a session.
// Sessions can be paused and resumed from where they left off.
package session

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/hydria-ai/hydria/internal/tracker"
)

// Session represents one attack session.
type Session struct {
	SessionID     string
	Target        string
	Service       string
	Username      string
	ImagePath     string
	WordlistPath  string
	CreatedAt     string
	UpdatedAt     string
	Status        string
	FoundPassword string
	TriedCount    int
	SkippedCount  int
}

func db() *sql.DB { return tracker.DB() }

// Create inserts a new session and returns its ID.
func Create(target, service, username, imagePath string) (string, error) {
	sessionID := fmt.Sprintf("sess_%s_%s",
		time.Now().Format("20060102_150405"),
		randomHex(6),
	)
	now := time.Now().Format(time.RFC3339)
	_, err := db().Exec(`
		INSERT INTO sessions
		(session_id, target, service, username, image_path, created_at, updated_at, status)
		VALUES (?,?,?,?,?,?,?,?)`,
		sessionID, target, service, username, imagePath, now, now, "running",
	)
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	return sessionID, nil
}

// Get fetches a session by ID. Returns nil if not found.
func Get(sessionID string) (*Session, error) {
	row := db().QueryRow(`SELECT * FROM sessions WHERE session_id=?`, sessionID)
	return scanSession(row)
}

// List returns all sessions ordered by most recent.
func List() ([]Session, error) {
	rows, err := db().Query(`SELECT * FROM sessions ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []Session
	for rows.Next() {
		s, err := scanSession(rows)
		if err != nil {
			continue
		}
		sessions = append(sessions, *s)
	}
	return sessions, nil
}

// UpdateStatus sets the session status (and optionally the found password).
func UpdateStatus(sessionID, status, foundPassword string) {
	now := time.Now().Format(time.RFC3339)
	if foundPassword != "" {
		db().Exec(
			`UPDATE sessions SET status=?, found_password=?, updated_at=? WHERE session_id=?`,
			status, foundPassword, now, sessionID,
		)
	} else {
		db().Exec(
			`UPDATE sessions SET status=?, updated_at=? WHERE session_id=?`,
			status, now, sessionID,
		)
	}
}

// UpdateCounts updates the tried/skipped counters.
func UpdateCounts(sessionID string, tried, skipped int) {
	db().Exec(
		`UPDATE sessions SET tried_count=?, skipped_count=?, updated_at=? WHERE session_id=?`,
		tried, skipped, time.Now().Format(time.RFC3339), sessionID,
	)
}

// UpdateWordlist saves the wordlist path to the session.
func UpdateWordlist(sessionID, wordlistPath string) {
	db().Exec(
		`UPDATE sessions SET wordlist_path=?, updated_at=? WHERE session_id=?`,
		wordlistPath, time.Now().Format(time.RFC3339), sessionID,
	)
}

// Pause marks a session as paused.
func Pause(sessionID string) {
	UpdateStatus(sessionID, "paused", "")
}

// FindResumable returns the most recent paused session for target/username, or nil.
func FindResumable(target, username string) (*Session, error) {
	row := db().QueryRow(`
		SELECT * FROM sessions
		WHERE target=? AND username=? AND status='paused'
		ORDER BY updated_at DESC LIMIT 1`,
		target, username,
	)
	return scanSession(row)
}

// ── helpers ───────────────────────────────────────────────────────────────

type scanner interface {
	Scan(dest ...any) error
}

func scanSession(row scanner) (*Session, error) {
	var s Session
	var foundPass sql.NullString
	var imagePath, wordlistPath sql.NullString
	err := row.Scan(
		&s.SessionID, &s.Target, &s.Service, &s.Username,
		&imagePath, &wordlistPath,
		&s.CreatedAt, &s.UpdatedAt, &s.Status,
		&foundPass, &s.TriedCount, &s.SkippedCount,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if foundPass.Valid {
		s.FoundPassword = foundPass.String
	}
	if imagePath.Valid {
		s.ImagePath = imagePath.String
	}
	if wordlistPath.Valid {
		s.WordlistPath = wordlistPath.String
	}
	return &s, nil
}

func randomHex(n int) string {
	const chars = "0123456789abcdef"
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[time.Now().UnixNano()%int64(len(chars))]
		time.Sleep(1)
	}
	return string(b)
}
