package storage

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Store struct{ db *sqlx.DB }

func NewStore(db *sqlx.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Create(ctx context.Context, e *Entry) (sql.Result, error) {
	enc, err := Encrypt(e.Password)
	if err != nil {
		return nil, err
	}
	e.Password = enc
	query := "INSERT INTO vault_entries (id, title, username, password, notes, tags, folder) VALUES (:id, :title, :username, :password, :notes, :tags, :folder)"

	return s.db.NamedExecContext(ctx, query, e)
}

func (s *Store) Get(ctx context.Context, id uuid.UUID) (*Entry, error) {
	var e Entry
	err := s.db.GetContext(ctx, &e, `SELECT * FROM vault_entries WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	dec, err := Decrypt(e.Password)
	if err != nil {
		return nil, err
	}
	e.Password = dec
	return &e, nil
}

func (s *Store) Delete(ctx context.Context, id uuid.UUID) (bool, error) {
	_, err := s.db.ExecContext(ctx, `DELETE FROM vault_entries WHERE id=$1`, id)
	if err != nil {
		return false, err
	}
	return true, nil
}

// Similar Update and Delete implementations...

func (s *Store) List(ctx context.Context, folder string, tags []string) ([]Entry, error) {
	query := `SELECT * FROM vault_entries WHERE 1=1`
	args := []interface{}{}
	if folder != "" {
		query += ` AND folder=$1`
		args = append(args, folder)
	}
	if len(tags) > 0 {
		query += ` AND tags @> $2`
		args = append(args, pq.Array(tags))
	}
	var entries []Entry
	err := s.db.SelectContext(ctx, &entries, query, args...)
	for _, entry := range entries {
		entry.Password, _ = Decrypt(entry.Password)
	}
	return entries, err
}
