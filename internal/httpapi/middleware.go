package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type middleware func(http.Handler) http.Handler

// chain wraps h so the first middleware is the outermost.
func chain(h http.Handler, mws ...middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (s *server) recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				s.log.ErrorContext(r.Context(), "panic in handler", "panic", v, "path", r.URL.Path)
				writeError(w, http.StatusInternalServerError, "internal", "something went wrong", nil)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, 8)
		_, _ = rand.Read(b)
		id := hex.EncodeToString(b)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r)
	})
}

func (s *server) accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		s.log.InfoContext(r.Context(), "request",
			"method", r.Method, "path", r.URL.Path, "status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(), "request_id", w.Header().Get("X-Request-ID"))
	})
}

func securityHeaders(secure bool) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("Referrer-Policy", "no-referrer")
			h.Set("Cross-Origin-Opener-Policy", "same-origin")
			if secure {
				h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
			}
			if len(r.URL.Path) >= 5 && r.URL.Path[:5] == "/api/" {
				h.Set("Cache-Control", "no-store")
			}
			next.ServeHTTP(w, r)
		})
	}
}

// originCheck is the CSRF defence for cookie auth: a state-changing request that a browser marks with an Origin of
// another site is refused. Together with SameSite=Lax cookies and JSON-only bodies this blocks cross-site forgery.
// Requests without Origin (curl, server-to-server) are allowed because browsers always send it on cross-site POSTs.
func originCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
		default:
			if o := r.Header.Get("Origin"); o != "" {
				u, err := url.Parse(o)
				if err != nil || u.Host != r.Host {
					writeError(w, http.StatusForbidden, "bad_origin", "cross-origin request refused", nil)
					return
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func bodyLimit(n int64) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, n)
			next.ServeHTTP(w, r)
		})
	}
}

// rateLimiter is a per-client-IP token bucket kept in memory. It protects bcrypt-heavy endpoints from brute force
// and CPU exhaustion. It keys on the TCP peer address; behind a reverse proxy configure the proxy to limit instead.
type rateLimiter struct {
	burst     int
	every     time.Duration
	mu        sync.Mutex
	clients   map[string]*client
	lastPrune time.Time
}

type client struct {
	lim  *rate.Limiter
	seen time.Time
}

func newRateLimiter(burst int, every time.Duration) *rateLimiter {
	return &rateLimiter{burst: burst, every: every, clients: map[string]*client{}}
}

func (rl *rateLimiter) allow(key string, now time.Time) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if now.Sub(rl.lastPrune) > time.Minute {
		for k, c := range rl.clients {
			if now.Sub(c.seen) > 10*time.Minute {
				delete(rl.clients, k)
			}
		}
		rl.lastPrune = now
	}
	c, ok := rl.clients[key]
	if !ok {
		c = &client{lim: rate.NewLimiter(rate.Every(rl.every), rl.burst)}
		rl.clients[key] = c
	}
	c.seen = now
	return c.lim.AllowN(now, 1)
}

func (rl *rateLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}
		if !rl.allow(host, time.Now()) {
			w.Header().Set("Retry-After", "6")
			writeError(w, http.StatusTooManyRequests, "rate_limited", "too many attempts, try again shortly", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}
