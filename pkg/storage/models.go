package storage

import (
	"database/sql"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"time"
)

type Entry struct {
	ID        uuid.UUID      `db:"id"`
	UserId    string         `db:"user_id"`
	Title     string         `db:"title"`
	Username  string         `db:"username"`
	Password  []byte         `db:"password"`
	Notes     sql.NullString `db:"notes"`
	Tags      pq.StringArray `db:"tags"`
	Folder    sql.NullString `db:"folder"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
}

type User struct {
	ID        uuid.UUID `db:"id"`
	Email     string    `db:"email"`
	Username  string    `db:"username"`
	Password  []byte    `db:"password"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
