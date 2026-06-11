package postgres

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/smallestfisher/relaydeck/backend/internal/auth"
	"github.com/smallestfisher/relaydeck/backend/internal/domain"
)

type UserStore struct {
	db *sql.DB
}

func Open(ctx context.Context, databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) EnsureSchema(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  role TEXT NOT NULL,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`)
	return err
}

func (s *UserStore) BootstrapOwner(ctx context.Context, email string, password string) error {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM users`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	id, err := newUUID()
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
INSERT INTO users (id, email, password_hash, role, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, now(), now())`, id, email, hash, domain.UserRoleOwner, domain.UserStatusActive)
	return err
}

func (s *UserStore) Users() []domain.User {
	rows, err := s.db.QueryContext(context.Background(), `
SELECT id::text, email, password_hash, role, status, created_at, updated_at
FROM users
ORDER BY created_at ASC`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	users := []domain.User{}
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil
		}
		users = append(users, user)
	}
	return users
}

func (s *UserStore) UserByEmail(email string) (domain.User, bool) {
	return s.queryUser(`
SELECT id::text, email, password_hash, role, status, created_at, updated_at
FROM users
WHERE email = $1`, email)
}

func (s *UserStore) UserByID(id string) (domain.User, bool) {
	return s.queryUser(`
SELECT id::text, email, password_hash, role, status, created_at, updated_at
FROM users
WHERE id = $1::uuid`, id)
}

func (s *UserStore) queryUser(query string, arg string) (domain.User, bool) {
	row := s.db.QueryRowContext(context.Background(), query, arg)
	user, err := scanUser(row)
	return user, err == nil
}

type userScanner interface {
	Scan(dest ...any) error
}

func scanUser(scanner userScanner) (domain.User, error) {
	var user domain.User
	var role string
	var status string
	if err := scanner.Scan(&user.ID, &user.Email, &user.PasswordHash, &role, &status, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return domain.User{}, err
	}
	user.Role = domain.UserRole(role)
	user.Status = domain.UserStatus(status)
	return user, nil
}

func newUUID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	return fmt.Sprintf("%s-%s-%s-%s-%s", hex.EncodeToString(buf[0:4]), hex.EncodeToString(buf[4:6]), hex.EncodeToString(buf[6:8]), hex.EncodeToString(buf[8:10]), hex.EncodeToString(buf[10:16])), nil
}
