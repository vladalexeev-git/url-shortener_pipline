package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"

	"url-shortener/internal/domain"
	"url-shortener/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmtEvent, err := db.Prepare(`

CREATE TABLE IF NOT EXISTS events(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		event_type TEXT NOT NULL,
		payload TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'new' CHECK(status IN ('new', 'done')),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: stmt event %w", op, err)
	}

	_, err = stmtEvent.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: stmt event exec %w", op, err)
	}

	stmtUrl, err := db.Prepare(`

	CREATE TABLE IF NOT EXISTS url(
		id INTEGER PRIMARY KEY,
		alias TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: stmt url %w", op, err)
	}

	_, err = stmtUrl.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: stmt url exec %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) (id int64, err error) {
	const op = "storage.sqlite.SaveURL"

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}

		commitErr := tx.Commit()
		if commitErr != nil {
			err = fmt.Errorf("%s: %w", op, commitErr)
		}
	}()

	stmt, err := tx.Prepare("INSERT INTO url(url, alias) VALUES(?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(urlToSave, alias)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err = res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}

	eventPayload := fmt.Sprintf(
		`{"id": %d, "url": "%s", "alias": "%s"}`,
		id,
		urlToSave,
		alias,
	)

	if err := s.saveEvent(tx, "URLCreated", eventPayload); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) saveEvent(tx *sql.Tx, eventType string, payload string) error {
	const op = "storage.sqlite.saveEvent"

	stmt, err := tx.Prepare("INSERT INTO events(event_type, payload) VALUES(?, ?)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(eventType, payload)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

type event struct {
	ID      int    `db:"id"`
	Type    string `db:"event_type"`
	Payload string `db:"payload"`
}

// GetNewEvent returns event with status 'new' from the database.
func (s *Storage) GetNewEvent() (domain.Event, error) {
	const op = "storage.sqlite.GetNewEvent"

	// TODO: add field `reserved_to` for locking events for processing
	row := s.db.QueryRow("SELECT id, event_type, payload FROM events WHERE status = 'new' LIMIT 1")

	var evt event

	err := row.Scan(&evt.ID, &evt.Type, &evt.Payload)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Event{}, nil // No new events found
		}

		return domain.Event{}, fmt.Errorf("%s: %w", op, err)
	}

	return domain.Event{
		ID:      evt.ID,
		Type:    evt.Type,
		Payload: evt.Payload,
	}, nil
}

func (s *Storage) SetDone(id int) error {
	const op = "storage.sqlite.MarkEventAsDone"

	stmt, err := s.db.Prepare("UPDATE events SET status = 'done' WHERE id = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.sqlite.GetURL"

	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	var resURL string

	err = stmt.QueryRow(alias).Scan(&resURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}

		return "", fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return resURL, nil
}

// TODO: implement method
// func (s *Storage) DeleteURL(alias string) error
