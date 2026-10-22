package httpapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Lanskask/user_auth/internal/auth"
)

type ctxKey struct{}

type credentials struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type userResponse struct {
	User auth.User `json:"user"`
}

func (s *server) register(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if !decodeJSON(w, r, &in) {
		return
	}
	u, err := s.opts.Auth.Register(r.Context(), in.Email, in.Name, in.Password)
	var ve auth.ValidationError
	switch {
	case errors.As(err, &ve):
		writeError(w, http.StatusUnprocessableEntity, "validation_failed", "check the highlighted fields", ve)
	case errors.Is(err, auth.ErrEmailTaken):
		writeError(w, http.StatusConflict, "email_taken", "an account with this email already exists",
			map[string]string{"email": "already registered"})
	case err != nil:
		s.internalError(w, r, err)
	default:
		writeJSON(w, http.StatusCreated, userResponse{u})
	}
}

func (s *server) login(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if !decodeJSON(w, r, &in) {
		return
	}
	u, token, expires, err := s.opts.Auth.Login(r.Context(), in.Email, in.Password)
	switch {
	case errors.Is(err, auth.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password", nil)
	case err != nil:
		s.internalError(w, r, err)
	default:
		http.SetCookie(w, s.sessionCookie(token, expires))
		writeJSON(w, http.StatusOK, userResponse{u})
	}
}

func (s *server) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(s.cookieName()); err == nil {
		if err := s.opts.Auth.Logout(r.Context(), c.Value); err != nil {
			s.internalError(w, r, err)
			return
		}
	}
	http.SetCookie(w, s.sessionCookie("", time.Unix(0, 0)))
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) me(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, userResponse{userFrom(r.Context())})
}

// requireAuth resolves the session cookie and puts the user in the request context.
func (s *server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(s.cookieName())
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthenticated", "sign in to continue", nil)
			return
		}
		u, err := s.opts.Auth.Authenticate(r.Context(), c.Value)
		switch {
		case errors.Is(err, auth.ErrUnauthenticated):
			http.SetCookie(w, s.sessionCookie("", time.Unix(0, 0)))
			writeError(w, http.StatusUnauthorized, "unauthenticated", "sign in to continue", nil)
		case err != nil:
			s.internalError(w, r, err)
		default:
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, u)))
		}
	})
}

func userFrom(ctx context.Context) auth.User {
	u, _ := ctx.Value(ctxKey{}).(auth.User)
	return u
}

// cookieName uses the __Host- prefix when Secure, which makes browsers refuse a cookie set by a sibling subdomain.
func (s *server) cookieName() string {
	if s.opts.CookieSecure {
		return "__Host-session"
	}
	return "session"
}

func (s *server) sessionCookie(value string, expires time.Time) *http.Cookie {
	c := &http.Cookie{ //nolint:gosec // Secure is configurable so plain-http local development works
		Name:     s.cookieName(),
		Value:    value,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		Secure:   s.opts.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	}
	if value == "" {
		c.MaxAge = -1
	}
	return c
}

func (s *server) internalError(w http.ResponseWriter, r *http.Request, err error) {
	s.log.ErrorContext(r.Context(), "request failed", "err", err, "method", r.Method, "path", r.URL.Path)
	writeError(w, http.StatusInternalServerError, "internal", "something went wrong", nil)
}
