package store

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Store wraps a SQLite database for note storage.
type Store struct {
	db *sql.DB
}

// Note represents a single day's note.
type Note struct {
	Date      string `json:"date"`
	Content   string `json:"content"`
	UpdatedAt string `json:"updated_at"`
}

// New opens (or creates) the SQLite database and runs migrations.
func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return &Store{db: db}, nil
}

// Close closes the database.
func (s *Store) Close() error {
	return s.db.Close()
}

func migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS notes (
		date       TEXT PRIMARY KEY,
		content    TEXT NOT NULL DEFAULT '',
		created_at TEXT NOT NULL DEFAULT (datetime('now')),
		updated_at TEXT NOT NULL DEFAULT (datetime('now'))
	);

	CREATE INDEX IF NOT EXISTS idx_notes_updated ON notes(updated_at);
	`
	_, err := db.Exec(schema)
	return err
}

// Upsert creates or updates a note for a given date.
// Returns the updated note and whether a conflict was detected.
func (s *Store) Upsert(date, content string) (*Note, error) {
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := s.db.Exec(`
		INSERT INTO notes (date, content, created_at, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(date) DO UPDATE SET
			content = excluded.content,
			updated_at = excluded.updated_at
	`, date, content, now, now)
	if err != nil {
		return nil, fmt.Errorf("upsert: %w", err)
	}

	return &Note{Date: date, Content: content, UpdatedAt: now}, nil
}

// Get retrieves a note by date. Returns nil if not found.
func (s *Store) Get(date string) (*Note, error) {
	row := s.db.QueryRow(`SELECT date, content, updated_at FROM notes WHERE date = ?`, date)

	var n Note
	err := row.Scan(&n.Date, &n.Content, &n.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}

	return &n, nil
}

// ListDates returns all note dates, most recent first.
func (s *Store) ListDates() ([]string, error) {
	rows, err := s.db.Query(`SELECT date FROM notes ORDER BY date DESC`)
	if err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}
	defer rows.Close()

	var dates []string
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		dates = append(dates, d)
	}

	return dates, rows.Err()
}

// Search performs full-text search across all notes.
// Returns notes whose content matches the query, most recent first.
func (s *Store) Search(query string) ([]Note, error) {
	rows, err := s.db.Query(`
		SELECT date, content, updated_at FROM notes
		WHERE content LIKE '%' || ? || '%'
		ORDER BY date DESC
		LIMIT 50
	`, query)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	defer rows.Close()

	var results []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.Date, &n.Content, &n.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		results = append(results, n)
	}

	return results, rows.Err()
}

// GetUpdatedSince returns notes updated after the given timestamp.
// Useful for incremental sync.
func (s *Store) GetUpdatedSince(since string) ([]Note, error) {
	rows, err := s.db.Query(`
		SELECT date, content, updated_at FROM notes
		WHERE updated_at > ?
		ORDER BY date DESC
	`, since)
	if err != nil {
		return nil, fmt.Errorf("updated since: %w", err)
	}
	defer rows.Close()

	var results []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.Date, &n.Content, &n.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		results = append(results, n)
	}

	return results, rows.Err()
}
