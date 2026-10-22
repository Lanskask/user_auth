// Package store implements auth.Store on PostgreSQL.
package store

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/Lanskask/user_auth/internal/auth"
)

//go:embed migrations/*.sql
var migrations embed.FS

const uniqueViolation = "23505"

// Store is a PostgreSQL-backed auth.Store.
type Store struct {
	pool *pgxpool.Pool
}

var _ auth.Store = (*Store)(nil)

// Open connects to Postgres, applies pending migrations and returns the Store.
func Open(ctx context.Context, url string) (*Store, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	if err := migrate(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}
	return &Store{pool: pool}, nil
}

func migrate(ctx context.Context, pool *pgxpool.Pool) error {
	db := stdlib.OpenDBFromPool(pool)
	defer func() { _ = db.Close() }() // a *sql.DB view of the pool; the pool itself is closed by Store.Close
	return runMigrations(ctx, db)
}

func runMigrations(ctx context.Context, db *sql.DB) error {
	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose dialect: %w", err)
	}
	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}

// Close releases the connection pool.
func (s *Store) Close() { s.pool.Close() }

// Ping reports whether the database is reachable (used by the readiness probe).
func (s *Store) Ping(ctx context.Context) error { return s.pool.Ping(ctx) }

// CreateUser inserts a user; a duplicate email yields auth.ErrEmailTaken.
func (s *Store) CreateUser(ctx context.Context, email, name, passwordHash string) (auth.User, error) {
	var u auth.User
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (email, name, password_hash) VALUES ($1, $2, $3)
		 RETURNING id, email, name, created_at`,
		email, name, passwordHash).Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
		return auth.User{}, auth.ErrEmailTaken
	}
	if err != nil {
		return auth.User{}, fmt.Errorf("insert user: %w", err)
	}
	return u, nil
}

// UserByEmail returns the user and the stored password hash.
func (s *Store) UserByEmail(ctx context.Context, email string) (auth.User, string, error) {
	var (
		u    auth.User
		hash string
	)
	err := s.pool.QueryRow(ctx,
		`SELECT id, email, name, created_at, password_hash FROM users WHERE lower(email) = lower($1)`,
		email).Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.User{}, "", auth.ErrNotFound
	}
	if err != nil {
		return auth.User{}, "", fmt.Errorf("select user: %w", err)
	}
	return u, hash, nil
}

// CreateSession stores a session keyed by the hash of its token.
func (s *Store) CreateSession(ctx context.Context, tokenHash []byte, userID int64, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO sessions (token_hash, user_id, expires_at) VALUES ($1, $2, $3)`,
		tokenHash, userID, expiresAt)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

// UserBySession returns the owner of an unexpired session.
func (s *Store) UserBySession(ctx context.Context, tokenHash []byte, now time.Time) (auth.User, error) {
	var u auth.User
	err := s.pool.QueryRow(ctx,
		`SELECT u.id, u.email, u.name, u.created_at
		 FROM sessions s JOIN users u ON u.id = s.user_id
		 WHERE s.token_hash = $1 AND s.expires_at > $2`,
		tokenHash, now).Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.User{}, auth.ErrNotFound
	}
	if err != nil {
		return auth.User{}, fmt.Errorf("select session: %w", err)
	}
	return u, nil
}

// DeleteSession removes a session; missing rows are fine.
func (s *Store) DeleteSession(ctx context.Context, tokenHash []byte) error {
	if _, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE token_hash = $1`, tokenHash); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// DeleteExpiredSessions removes sessions that expired before now and returns how many.
func (s *Store) DeleteExpiredSessions(ctx context.Context, now time.Time) (int64, error) {
	tag, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE expires_at <= $1`, now)
	if err != nil {
		return 0, fmt.Errorf("purge sessions: %w", err)
	}
	return tag.RowsAffected(), nil
}
