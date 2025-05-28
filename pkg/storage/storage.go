package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/AleksZelenchuk/vault-server/pkg/auth"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Store struct{ db *sqlx.DB }

func NewStore(db *sqlx.DB) *Store {
	return &Store{db: db}
}

var NoUserId = errors.New("no user id provided")
var PermissionDenied = errors.New("you dont have permission to do this")

func (s *Store) Create(ctx context.Context, e *Entry) (sql.Result, error) {
	userId, _ := auth.UserIDFromContext(ctx)
	if userId == "" {
		return nil, NoUserId
	}

	enc, err := Encrypt(e.Password)
	if err != nil {
		return nil, err
	}
	e.Password = enc
	query := "INSERT INTO vault_entries (id, title, username, password, notes, tags, folder, user_id, domain) VALUES (:id, :title, :username, :password, :notes, :tags, :folder, :user_id, :domain)"

	return s.db.NamedExecContext(ctx, query, e)
}

func (s *Store) Get(ctx context.Context, id uuid.UUID) (*Entry, error) {
	userId, _ := auth.UserIDFromContext(ctx)
	if userId == "" {
		return nil, NoUserId
	}

	var e Entry
	err := s.db.GetContext(ctx, &e, `SELECT * FROM vault_entries WHERE id=$1 AND user_id=$2`, id, userId)
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
	permErr := s.validateUserPermission(ctx, id)
	if permErr != nil {
		return false, permErr
	}

	res, err := s.db.ExecContext(ctx, `DELETE FROM vault_entries WHERE id=$1`, id)
	if err != nil {
		return false, err
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	if ra == int64(0) {
		return false, sql.ErrNoRows
	}

	return true, nil
}

func (s *Store) List(ctx context.Context, domain string, folder string, tags []string) ([]Entry, error) {
	query := `SELECT * FROM vault_entries WHERE 1=1`

	userId, _ := auth.UserIDFromContext(ctx)
	if userId == "" {
		return nil, NoUserId
	}
	var args []interface{}

	query += ` AND user_id=$1`
	args = append(args, userId)

	if len(domain) > 0 {
		query += ` AND "domain" LIKE $2`
		args = append(args, "%"+domain+"%")
	}

	if folder != "" {
		query += ` AND folder=$3`
		args = append(args, folder)
	}
	if len(tags) > 0 {
		query += ` AND tags @> $4`
		args = append(args, pq.Array(tags))
	}

	var entries []Entry
	err := s.db.SelectContext(ctx, &entries, query, args...)

	for _, entry := range entries {
		entry.Password, _ = Decrypt(entry.Password)
	}
	return entries, err
}

// validateUserPermission we need to check if given used have permission to perform action with the requested entry
// before proceeding
func (s *Store) validateUserPermission(ctx context.Context, id uuid.UUID) error {
	userId, _ := auth.UserIDFromContext(ctx)
	if userId == "" {
		return NoUserId
	}
	row := s.db.QueryRowContext(ctx, `SELECT user_id FROM vault_entries WHERE id=$1`, id)

	if row == nil {
		return errors.New("no row found")
	}

	var dbUserId sql.NullString
	if err := row.Scan(&dbUserId); err != nil {
		return err
	}

	if !dbUserId.Valid || dbUserId.String != userId {
		return PermissionDenied
	}

	return nil
}
