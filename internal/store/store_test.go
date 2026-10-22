package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Lanskask/user_auth/internal/auth"
)

// newTestStore creates a throwaway database on the server named by TEST_DATABASE_URL, so the tests never
// touch (or truncate) real data. The test is skipped when the variable is unset.
func newTestStore(t *testing.T) *Store {
	t.Helper()
	adminURL := os.Getenv("TEST_DATABASE_URL")
	if adminURL == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping Postgres integration test")
	}
	ctx := context.Background()

	suffix := make([]byte, 6)
	if _, err := rand.Read(suffix); err != nil {
		t.Fatal(err)
	}
	dbName := "test_" + hex.EncodeToString(suffix)

	admin, err := pgx.Connect(ctx, adminURL)
	if err != nil {
		t.Fatalf("connect admin: %v", err)
	}
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+dbName); err != nil {
		t.Fatalf("create database: %v", err)
	}
	t.Cleanup(func() {
		_, _ = admin.Exec(ctx, "DROP DATABASE "+dbName+" WITH (FORCE)")
		_ = admin.Close(ctx)
	})

	u, err := url.Parse(adminURL)
	if err != nil {
		t.Fatal(err)
	}
	u.Path = "/" + dbName
	s, err := Open(ctx, u.String())
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(s.Close)
	return s
}

func TestUsers(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	u, err := s.CreateUser(ctx, "a@example.com", "Ann", "hash")
	if err != nil || u.ID == 0 {
		t.Fatalf("CreateUser = %+v, %v", u, err)
	}
	if _, err := s.CreateUser(ctx, "A@EXAMPLE.com", "Dup", "hash"); !errors.Is(err, auth.ErrEmailTaken) {
		t.Fatalf("duplicate: got %v, want ErrEmailTaken", err)
	}

	got, hash, err := s.UserByEmail(ctx, "A@example.com")
	if err != nil || got.ID != u.ID || hash != "hash" {
		t.Fatalf("UserByEmail = %+v, %q, %v", got, hash, err)
	}
	if _, _, err := s.UserByEmail(ctx, "ghost@example.com"); !errors.Is(err, auth.ErrNotFound) {
		t.Fatalf("missing: got %v, want ErrNotFound", err)
	}
}

func TestSessions(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	u, err := s.CreateUser(ctx, "a@example.com", "Ann", "hash")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()

	if err := s.CreateSession(ctx, []byte("live"), u.ID, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateSession(ctx, []byte("dead"), u.ID, now.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}

	if got, err := s.UserBySession(ctx, []byte("live"), now); err != nil || got.ID != u.ID {
		t.Fatalf("live session: %+v, %v", got, err)
	}
	if _, err := s.UserBySession(ctx, []byte("dead"), now); !errors.Is(err, auth.ErrNotFound) {
		t.Fatalf("expired session: got %v", err)
	}

	if n, err := s.DeleteExpiredSessions(ctx, now); err != nil || n != 1 {
		t.Fatalf("DeleteExpiredSessions = %d, %v", n, err)
	}
	if err := s.DeleteSession(ctx, []byte("live")); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UserBySession(ctx, []byte("live"), now); !errors.Is(err, auth.ErrNotFound) {
		t.Fatalf("deleted session: got %v", err)
	}
}
