package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type fakeStore struct {
	users    map[string]fakeUser // by email
	sessions map[string]fakeSession
	nextID   int64
}

type fakeUser struct {
	User
	hash string
}

type fakeSession struct {
	userID  int64
	expires time.Time
}

func newFakeStore() *fakeStore {
	return &fakeStore{users: map[string]fakeUser{}, sessions: map[string]fakeSession{}}
}

func (f *fakeStore) CreateUser(_ context.Context, email, name, hash string) (User, error) {
	if _, ok := f.users[email]; ok {
		return User{}, ErrEmailTaken
	}
	f.nextID++
	u := User{ID: f.nextID, Email: email, Name: name}
	f.users[email] = fakeUser{u, hash}
	return u, nil
}

func (f *fakeStore) UserByEmail(_ context.Context, email string) (User, string, error) {
	u, ok := f.users[email]
	if !ok {
		return User{}, "", ErrNotFound
	}
	return u.User, u.hash, nil
}

func (f *fakeStore) CreateSession(_ context.Context, h []byte, uid int64, exp time.Time) error {
	f.sessions[string(h)] = fakeSession{uid, exp}
	return nil
}

func (f *fakeStore) UserBySession(_ context.Context, h []byte, now time.Time) (User, error) {
	s, ok := f.sessions[string(h)]
	if !ok || !s.expires.After(now) {
		return User{}, ErrNotFound
	}
	for _, u := range f.users {
		if u.ID == s.userID {
			return u.User, nil
		}
	}
	return User{}, ErrNotFound
}

func (f *fakeStore) DeleteSession(_ context.Context, h []byte) error {
	delete(f.sessions, string(h))
	return nil
}

func newTestService(t *testing.T) (*Service, *fakeStore) {
	t.Helper()
	fs := newFakeStore()
	svc, err := NewService(fs, bcrypt.MinCost, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return svc, fs
}

func TestRegisterValidation(t *testing.T) {
	tests := []struct {
		name, email, userName, password string
		wantFields                      []string
	}{
		{"valid", "a@example.com", "Ann", "password1", nil},
		{"bad email", "not-an-email", "Ann", "password1", []string{"email"}},
		{"display-name email rejected", "Ann <a@example.com>", "Ann", "password1", []string{"email"}},
		{"empty name", "a@example.com", "  ", "password1", []string{"name"}},
		{"short password", "a@example.com", "Ann", "short", []string{"password"}},
		{"password over 72 bytes", "a@example.com", "Ann", strings.Repeat("x", 73), []string{"password"}},
		{"everything wrong", "", "", "", []string{"email", "name", "password"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _ := newTestService(t)
			_, err := svc.Register(context.Background(), tt.email, tt.userName, tt.password)
			if tt.wantFields == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			var ve ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("want ValidationError, got %v", err)
			}
			for _, f := range tt.wantFields {
				if _, ok := ve[f]; !ok {
					t.Errorf("missing error for field %q: %v", f, ve)
				}
			}
			if len(ve) != len(tt.wantFields) {
				t.Errorf("got fields %v, want %v", ve, tt.wantFields)
			}
		})
	}
}

func TestRegisterDuplicateEmailIsCaseInsensitive(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	if _, err := svc.Register(ctx, "a@example.com", "Ann", "password1"); err != nil {
		t.Fatal(err)
	}
	_, err := svc.Register(ctx, " A@Example.com ", "Ann2", "password1")
	if !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("got %v, want ErrEmailTaken", err)
	}
}

func TestRegisterStoresHashNotPassword(t *testing.T) {
	svc, fs := newTestService(t)
	if _, err := svc.Register(context.Background(), "a@example.com", "Ann", "password1"); err != nil {
		t.Fatal(err)
	}
	if h := fs.users["a@example.com"].hash; h == "password1" || bcrypt.CompareHashAndPassword([]byte(h), []byte("password1")) != nil {
		t.Fatalf("password not stored as a valid bcrypt hash: %q", h)
	}
}

func TestLoginAuthenticateLogout(t *testing.T) {
	svc, fs := newTestService(t)
	ctx := context.Background()
	if _, err := svc.Register(ctx, "a@example.com", "Ann", "password1"); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name, email, password string
	}{
		{"wrong password", "a@example.com", "nope-nope"},
		{"unknown email", "ghost@example.com", "password1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, err := svc.Login(ctx, tt.email, tt.password)
			if !errors.Is(err, ErrInvalidCredentials) {
				t.Fatalf("got %v, want ErrInvalidCredentials", err)
			}
		})
	}

	u, token, expires, err := svc.Login(ctx, "A@example.com", "password1")
	if err != nil {
		t.Fatal(err)
	}
	if u.Email != "a@example.com" || token == "" || !expires.After(time.Now()) {
		t.Fatalf("bad login result: %+v %q %v", u, token, expires)
	}
	if _, ok := fs.sessions[token]; ok {
		t.Fatal("raw token must not be used as the storage key")
	}

	got, err := svc.Authenticate(ctx, token)
	if err != nil || got.ID != u.ID {
		t.Fatalf("Authenticate = %+v, %v", got, err)
	}

	if err := svc.Logout(ctx, token); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Authenticate(ctx, token); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("after logout got %v, want ErrUnauthenticated", err)
	}
}

func TestAuthenticateRejectsExpiredAndEmpty(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	if _, err := svc.Register(ctx, "a@example.com", "Ann", "password1"); err != nil {
		t.Fatal(err)
	}
	_, token, _, err := svc.Login(ctx, "a@example.com", "password1")
	if err != nil {
		t.Fatal(err)
	}

	svc.now = func() time.Time { return time.Now().Add(2 * time.Hour) }
	if _, err := svc.Authenticate(ctx, token); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("expired: got %v", err)
	}
	if _, err := svc.Authenticate(ctx, ""); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("empty: got %v", err)
	}
}

func TestTokensAreUnique(t *testing.T) {
	a, err1 := newToken()
	b, err2 := newToken()
	if err1 != nil || err2 != nil || a == b || len(a) < 43 {
		t.Fatalf("tokens %q %q (%v %v)", a, b, err1, err2)
	}
}
