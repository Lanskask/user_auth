// Package httpapi exposes the auth service over HTTP: JSON API under /api/v1, health probes and the SPA.
package httpapi

import (
	"context"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"github.com/Lanskask/user_auth/internal/auth"
)

// Authenticator is the part of auth.Service the handlers use.
type Authenticator interface {
	Register(ctx context.Context, email, name, password string) (auth.User, error)
	Login(ctx context.Context, email, password string) (auth.User, string, time.Time, error)
	Authenticate(ctx context.Context, token string) (auth.User, error)
	Logout(ctx context.Context, token string) error
}

// Pinger reports database reachability for the readiness probe.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Options configures the HTTP handler.
type Options struct {
	Auth         Authenticator
	DB           Pinger
	Logger       *slog.Logger
	CookieSecure bool
	// Static is the built frontend; nil disables static serving (API only).
	Static fs.FS
}

type server struct {
	opts Options
	log  *slog.Logger
}

// New builds the complete handler tree with all middleware applied.
func New(opts Options) http.Handler {
	s := &server{opts: opts, log: opts.Logger}

	authLimit := newRateLimiter(10, 6*time.Second) // burst of 10, then one request per 6 s per client IP

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.HandleFunc("GET /readyz", s.readyz)
	mux.Handle("POST /api/v1/auth/register", authLimit.middleware(http.HandlerFunc(s.register)))
	mux.Handle("POST /api/v1/auth/login", authLimit.middleware(http.HandlerFunc(s.login)))
	mux.HandleFunc("POST /api/v1/auth/logout", s.logout)
	mux.Handle("GET /api/v1/me", s.requireAuth(http.HandlerFunc(s.me)))
	mux.HandleFunc("/api/", func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusNotFound, "not_found", "no such endpoint", nil)
	})
	if opts.Static != nil {
		mux.Handle("/", spaHandler(opts.Static))
	}

	return chain(mux,
		s.recoverer,
		requestID,
		s.accessLog,
		securityHeaders(opts.CookieSecure),
		originCheck,
		bodyLimit(16<<10),
	)
}

func (s *server) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := s.opts.DB.Ping(ctx); err != nil {
		s.log.WarnContext(r.Context(), "readiness check failed", "err", err)
		writeError(w, http.StatusServiceUnavailable, "not_ready", "database unavailable", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
