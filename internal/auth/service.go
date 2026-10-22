// Package auth holds the registration, login and session rules. It knows nothing about HTTP or SQL.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrEmailTaken is returned by Store.CreateUser when the email is already registered.
	ErrEmailTaken = errors.New("email already registered")
	// ErrNotFound is returned by Store lookups that match nothing.
	ErrNotFound = errors.New("not found")
	// ErrInvalidCredentials is returned for an unknown email or a wrong password (deliberately indistinguishable).
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrUnauthenticated is returned when a session token is missing, unknown or expired.
	ErrUnauthenticated = errors.New("not authenticated")
)

// ValidationError maps a field name to a human-readable problem.
type ValidationError map[string]string

func (v ValidationError) Error() string {
	return fmt.Sprintf("invalid input: %v", map[string]string(v))
}

// User is the public view of an account; the password hash never leaves the store.
type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

// Store is the persistence the Service needs. It is defined here, by the consumer.
type Store interface {
	CreateUser(ctx context.Context, email, name, passwordHash string) (User, error)
	UserByEmail(ctx context.Context, email string) (User, string, error)
	CreateSession(ctx context.Context, tokenHash []byte, userID int64, expiresAt time.Time) error
	UserBySession(ctx context.Context, tokenHash []byte, now time.Time) (User, error)
	DeleteSession(ctx context.Context, tokenHash []byte) error
}

// Service implements the auth use cases.
type Service struct {
	store     Store
	cost      int
	ttl       time.Duration
	now       func() time.Time
	dummyHash []byte // compared against when the email is unknown, so timing does not reveal who is registered
}

// NewService builds a Service. cost is the bcrypt cost, ttl the session lifetime.
func NewService(store Store, cost int, ttl time.Duration) (*Service, error) {
	dummy, err := bcrypt.GenerateFromPassword([]byte("dummy-password"), cost)
	if err != nil {
		return nil, fmt.Errorf("generate dummy hash: %w", err)
	}
	return &Service{store: store, cost: cost, ttl: ttl, now: time.Now, dummyHash: dummy}, nil
}

// Register validates the input and creates an account.
func (s *Service) Register(ctx context.Context, email, name, password string) (User, error) {
	email = normalizeEmail(email)
	name = strings.TrimSpace(name)
	if errs := validate(email, name, password); len(errs) > 0 {
		return User{}, errs
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	if err != nil {
		return User{}, fmt.Errorf("hash password: %w", err)
	}
	u, err := s.store.CreateUser(ctx, email, name, string(hash))
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

// Login checks the credentials and opens a session. It returns the raw token (for the cookie) and its expiry.
func (s *Service) Login(ctx context.Context, email, password string) (User, string, time.Time, error) {
	u, hash, err := s.store.UserByEmail(ctx, normalizeEmail(email))
	switch {
	case errors.Is(err, ErrNotFound):
		_ = bcrypt.CompareHashAndPassword(s.dummyHash, []byte(password))
		return User{}, "", time.Time{}, ErrInvalidCredentials
	case err != nil:
		return User{}, "", time.Time{}, fmt.Errorf("find user: %w", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return User{}, "", time.Time{}, ErrInvalidCredentials
	}

	token, err := newToken()
	if err != nil {
		return User{}, "", time.Time{}, err
	}
	expires := s.now().Add(s.ttl)
	if err := s.store.CreateSession(ctx, hashToken(token), u.ID, expires); err != nil {
		return User{}, "", time.Time{}, fmt.Errorf("create session: %w", err)
	}
	return u, token, expires, nil
}

// Authenticate resolves a session token to its user.
func (s *Service) Authenticate(ctx context.Context, token string) (User, error) {
	if token == "" {
		return User{}, ErrUnauthenticated
	}
	u, err := s.store.UserBySession(ctx, hashToken(token), s.now())
	if errors.Is(err, ErrNotFound) {
		return User{}, ErrUnauthenticated
	}
	if err != nil {
		return User{}, fmt.Errorf("find session: %w", err)
	}
	return u, nil
}

// Logout ends the session. An unknown token is not an error.
func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	if err := s.store.DeleteSession(ctx, hashToken(token)); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// TTL is the session lifetime, used by the HTTP layer for the cookie.
func (s *Service) TTL() time.Duration { return s.ttl }

func normalizeEmail(e string) string { return strings.ToLower(strings.TrimSpace(e)) }

func validate(email, name, password string) ValidationError {
	errs := ValidationError{}
	if addr, err := mail.ParseAddress(email); err != nil || addr.Address != email || len(email) > 254 {
		errs["email"] = "enter a valid email address"
	}
	if n := utf8.RuneCountInString(name); n < 1 || n > 100 {
		errs["name"] = "name must be 1 to 100 characters"
	}
	// bcrypt ignores everything after 72 bytes, so longer passwords would be silently truncated.
	if len(password) < 8 || len(password) > 72 {
		errs["password"] = "password must be 8 to 72 bytes"
	}
	return errs
}

// newToken returns 256 random bits, base64url encoded.
func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// hashToken is what the database stores, so a leaked table cannot be replayed as cookies.
func hashToken(token string) []byte {
	h := sha256.Sum256([]byte(token))
	return h[:]
}
