package storage

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserStore struct{ db *sqlx.DB }

func NewUserStore(db *sqlx.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) CreateUser(ctx context.Context, e *User) (sql.Result, error) {
	enc, err := Encrypt(e.Password)
	if err != nil {
		return nil, err
	}
	e.Password = enc
	query := "INSERT INTO vault_users (id, email, username, password) VALUES (:id, :email, :username, :password)"

	return s.db.NamedExecContext(ctx, query, e)
}

// Deprecated: GetByUsername: should be reworked to use user id instead
func (s *UserStore) GetByUsername(ctx context.Context, username string) (*User, error) {
	var e User
	err := s.db.GetContext(ctx, &e, `SELECT * FROM vault_users WHERE username=$1`, username)
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

func (s *UserStore) DeleteUser(ctx context.Context, id uuid.UUID) (bool, error) {
	_, err := s.db.ExecContext(ctx, `DELETE FROM vault_users WHERE id=$1`, id)
	if err != nil {
		return false, err
	}
	return true, nil
}
