package httpapi

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/Lanskask/user_auth/internal/auth"
)

type stubAuth struct {
	registerErr error
	loginErr    error
	authErr     error
}

var testUser = auth.User{ID: 1, Email: "a@example.com", Name: "Ann"}

func (s stubAuth) Register(context.Context, string, string, string) (auth.User, error) {
	return testUser, s.registerErr
}

func (s stubAuth) Login(context.Context, string, string) (auth.User, string, time.Time, error) {
	return testUser, "tok", time.Now().Add(time.Hour), s.loginErr
}

func (s stubAuth) Authenticate(_ context.Context, token string) (auth.User, error) {
	if s.authErr != nil || token != "tok" {
		return auth.User{}, auth.ErrUnauthenticated
	}
	return testUser, nil
}

func (stubAuth) Logout(context.Context, string) error { return nil }

type stubDB struct{ err error }

func (d stubDB) Ping(context.Context) error { return d.err }

func newTestServer(a stubAuth, db stubDB) http.Handler {
	return New(Options{
		Auth: a, DB: db, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Static: fstest.MapFS{"index.html": {Data: []byte("<html>spa</html>")}},
	})
}

func do(h http.Handler, method, path, body string, hdr map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Host = "example.com"
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name string
		auth stubAuth
		body string
		want int
	}{
		{"created", stubAuth{}, `{"email":"a@example.com","name":"Ann","password":"password1"}`, http.StatusCreated},
		{"validation", stubAuth{registerErr: auth.ValidationError{"email": "bad"}}, `{}`, http.StatusUnprocessableEntity},
		{"duplicate", stubAuth{registerErr: auth.ErrEmailTaken}, `{}`, http.StatusConflict},
		{"internal error hides details", stubAuth{registerErr: errors.New("pq: secret detail")}, `{}`, http.StatusInternalServerError},
		{"malformed json", stubAuth{}, `{`, http.StatusBadRequest},
		{"unknown field", stubAuth{}, `{"admin":true}`, http.StatusBadRequest},
		{"trailing data", stubAuth{}, `{} {}`, http.StatusBadRequest},
		{"too large", stubAuth{}, `{"name":"` + strings.Repeat("x", 20<<10) + `"}`, http.StatusRequestEntityTooLarge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(newTestServer(tt.auth, stubDB{}), "POST", "/api/v1/auth/register", tt.body, nil)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tt.want, rec.Body)
			}
			if strings.Contains(rec.Body.String(), "secret detail") {
				t.Fatal("internal error leaked to client")
			}
		})
	}
}

func TestLoginSetsHardenedCookie(t *testing.T) {
	rec := do(newTestServer(stubAuth{}, stubDB{}), "POST", "/api/v1/auth/login", `{"email":"a@example.com","password":"x"}`, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	c := rec.Result().Cookies()
	if len(c) != 1 || c[0].Value != "tok" || !c[0].HttpOnly || c[0].SameSite != http.SameSiteLaxMode || c[0].Path != "/" {
		t.Fatalf("cookie = %+v", c)
	}
	if strings.Contains(rec.Body.String(), "password") {
		t.Fatal("response must not contain password fields")
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	rec := do(newTestServer(stubAuth{loginErr: auth.ErrInvalidCredentials}, stubDB{}), "POST", "/api/v1/auth/login", `{}`, nil)
	if rec.Code != http.StatusUnauthorized || len(rec.Result().Cookies()) != 0 {
		t.Fatalf("status = %d cookies = %v", rec.Code, rec.Result().Cookies())
	}
}

func TestMeRequiresSession(t *testing.T) {
	h := newTestServer(stubAuth{}, stubDB{})
	if rec := do(h, "GET", "/api/v1/me", "", nil); rec.Code != http.StatusUnauthorized {
		t.Fatalf("no cookie: %d", rec.Code)
	}
	if rec := do(h, "GET", "/api/v1/me", "", map[string]string{"Cookie": "session=wrong"}); rec.Code != http.StatusUnauthorized {
		t.Fatalf("bad cookie: %d", rec.Code)
	}
	rec := do(h, "GET", "/api/v1/me", "", map[string]string{"Cookie": "session=tok"})
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "a@example.com") {
		t.Fatalf("good cookie: %d %s", rec.Code, rec.Body)
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("API responses must not be cached")
	}
}

func TestLogoutClearsCookie(t *testing.T) {
	hdr := map[string]string{"Cookie": "session=tok", "Origin": "http://example.com"}
	rec := do(newTestServer(stubAuth{}, stubDB{}), "POST", "/api/v1/auth/logout", "", hdr)
	c := rec.Result().Cookies()
	if rec.Code != http.StatusNoContent || len(c) != 1 || c[0].MaxAge >= 0 {
		t.Fatalf("status = %d cookies = %+v", rec.Code, c)
	}
}

func TestCrossOriginPostRefused(t *testing.T) {
	h := newTestServer(stubAuth{}, stubDB{})
	hdr := map[string]string{"Origin": "https://evil.example"}
	if rec := do(h, "POST", "/api/v1/auth/logout", "", hdr); rec.Code != http.StatusForbidden {
		t.Fatalf("cross-origin: %d", rec.Code)
	}
	hdr["Origin"] = "http://example.com"
	if rec := do(h, "POST", "/api/v1/auth/logout", "", hdr); rec.Code != http.StatusNoContent {
		t.Fatalf("same-origin: %d", rec.Code)
	}
	hdr["Origin"] = "https://example.com" // same host, wrong scheme for a non-Secure deployment
	if rec := do(h, "POST", "/api/v1/auth/logout", "", hdr); rec.Code != http.StatusForbidden {
		t.Fatalf("scheme mismatch: %d", rec.Code)
	}
}

func TestPostWithSessionCookieButNoOriginRefused(t *testing.T) {
	h := newTestServer(stubAuth{}, stubDB{})
	if rec := do(h, "POST", "/api/v1/auth/logout", "", map[string]string{"Cookie": "session=tok"}); rec.Code != http.StatusForbidden {
		t.Fatalf("cookie without Origin: %d", rec.Code)
	}
	if rec := do(h, "POST", "/api/v1/auth/logout", "", nil); rec.Code != http.StatusNoContent {
		t.Fatalf("no cookie, no Origin (curl): %d", rec.Code)
	}
}

func TestRateLimitOnLogin(t *testing.T) {
	h := newTestServer(stubAuth{loginErr: auth.ErrInvalidCredentials}, stubDB{})
	var last int
	for range 11 {
		last = do(h, "POST", "/api/v1/auth/login", `{}`, nil).Code
	}
	if last != http.StatusTooManyRequests {
		t.Fatalf("11th attempt status = %d, want 429", last)
	}
}

func TestProbes(t *testing.T) {
	if rec := do(newTestServer(stubAuth{}, stubDB{}), "GET", "/healthz", "", nil); rec.Code != 200 {
		t.Fatalf("healthz %d", rec.Code)
	}
	if rec := do(newTestServer(stubAuth{}, stubDB{}), "GET", "/readyz", "", nil); rec.Code != 200 {
		t.Fatalf("readyz %d", rec.Code)
	}
	if rec := do(newTestServer(stubAuth{}, stubDB{err: errors.New("down")}), "GET", "/readyz", "", nil); rec.Code != 503 {
		t.Fatalf("readyz down %d", rec.Code)
	}
}

func TestSPAFallbackAndAPI404(t *testing.T) {
	h := newTestServer(stubAuth{}, stubDB{})
	rec := do(h, "GET", "/dashboard", "", nil)
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "spa") {
		t.Fatalf("spa fallback: %d %s", rec.Code, rec.Body)
	}
	if rec := do(h, "GET", "/api/v1/nope", "", nil); rec.Code != 404 || !strings.Contains(rec.Body.String(), "not_found") {
		t.Fatalf("api 404: %d %s", rec.Code, rec.Body)
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" || rec.Header().Get("Content-Security-Policy") == "" {
		t.Fatal("security headers missing")
	}
}

func TestRecoverer(t *testing.T) {
	s := &server{log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	h := s.recoverer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") }))
	if rec := do(h, "GET", "/", "", nil); rec.Code != 500 {
		t.Fatalf("status = %d", rec.Code)
	}
}
