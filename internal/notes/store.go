package notes

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgxpool"
)

const MaxTextCharacters = 2000

var ErrInvalidText = errors.New("text must contain 1 to 2000 characters")

type Note struct {
	ID        int64     `json:"id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

func CleanText(text string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" || strings.ContainsRune(text, 0) || !utf8.ValidString(text) || utf8.RuneCountInString(text) > MaxTextCharacters {
		return "", ErrInvalidText
	}
	return text, nil
}

type Store struct {
	pool *pgxpool.Pool
}

func Open(ctx context.Context, url string) (*Store, error) {
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, errors.New("invalid database configuration")
	}
	config.MaxConns = 4
	config.MinConns = 0
	config.ConnConfig.ConnectTimeout = 3 * time.Second
	config.ConnConfig.RuntimeParams["statement_timeout"] = "2000"
	config.ConnConfig.RuntimeParams["timezone"] = "UTC"
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, errors.New("database initialization failed")
	}
	fail := func() (*Store, error) {
		pool.Close()
		return nil, errors.New("database initialization failed")
	}
	if err = pool.Ping(ctx); err != nil {
		return fail()
	}
	_, err = pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS app.notes (
        id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
        text text NOT NULL CHECK (char_length(text) BETWEEN 1 AND 2000),
        created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
    )`)
	if err != nil {
		return fail()
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() { s.pool.Close() }

func (s *Store) Ready(ctx context.Context) error { return s.pool.Ping(ctx) }

func (s *Store) Create(ctx context.Context, text string) (Note, error) {
	text, err := CleanText(text)
	if err != nil {
		return Note{}, err
	}
	var note Note
	err = s.pool.QueryRow(ctx, `INSERT INTO app.notes (text) VALUES ($1) RETURNING id, text, created_at`, text).
		Scan(&note.ID, &note.Text, &note.CreatedAt)
	return note, err
}

func (s *Store) List(ctx context.Context) ([]Note, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, text, created_at FROM app.notes ORDER BY id DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Note, 0)
	for rows.Next() {
		var note Note
		if err := rows.Scan(&note.ID, &note.Text, &note.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, note)
	}
	return result, rows.Err()
}
