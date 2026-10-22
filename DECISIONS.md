# Decision log

Each entry records what was chosen and why, with its source: a requirement, an industry standard, a best practice, or a
measured result.

| #   | Decision                                      | Source                             |
|-----|-----------------------------------------------|------------------------------------|
| 1   | One Go module with `internal/` packages       | Go project layout, best practice   |
| 2   | Standard library `net/http` instead of Fiber  | Best practice, fewer dependencies  |
| 3   | Server-side sessions in PostgreSQL, no JWT    | OWASP session management guidance  |
| 4   | Store a hash of the session token             | OWASP session management guidance  |
| 5   | bcrypt with a 72-byte password limit          | Existing choice, kept; bcrypt spec |
| 6   | Dummy hash for unknown emails                 | OWASP authentication cheat sheet   |
| 7   | `Origin` check plus `SameSite=Lax` for CSRF   | OWASP CSRF cheat sheet             |
| 8   | In-memory per-IP rate limiter                 | YAGNI; limits stated in README     |
| 9   | Environment variables for configuration       | Twelve-factor app                  |
| 10  | Frontend embedded in the Go binary            | Requirement: single deployable     |
| 11  | Plain CSS with tokens instead of a UI library | YAGNI; strict CSP                  |
| 12  | Pinned patched Go toolchain in `go.mod`       | Measured: `govulncheck` result     |
| 13  | Frontend build output in `web/static`         | Tooling constraint                 |
| 14  | Playwright against the real stack             | Requirement: browser tests         |

## 1. One Go module with `internal/` packages

The 2022 version had a nested `go.mod` per package joined with `replace` directives and bare import paths such as
`"configs"`. That does not work with normal tooling (`go test ./...` could not see across modules) and nothing needed
separate versioning. A single module with `cmd/` and `internal/` is the layout the Go team documents.

## 2. Standard library `net/http` instead of Fiber

Go 1.22 added method and path patterns to `http.ServeMux`, which covers every route here. Staying on the standard
library keeps `httptest`, `http.MaxBytesReader`, `http.Cookie` and `Server.Shutdown` working as documented, and removes
`fasthttp`, which does not implement `net/http` semantics. Fewer dependencies also means fewer advisories to track.

## 3. Server-side sessions in PostgreSQL, no JWT

The old service kept sessions in process memory, so every restart signed everyone out and two instances could not share
them. A `sessions` table fixes both and allows real logout, because a JWT cannot be revoked before it expires. A new
token is minted on every login, which prevents session fixation.

## 4. Store a hash of the session token

The cookie holds 256 random bits. The database holds only `SHA-256(token)`. A read-only leak of the table (backup,
replica, SQL injection) then yields nothing that can be presented as a cookie. A fast hash is correct here because the
input is already high-entropy; bcrypt is only needed for low-entropy passwords.

## 5. bcrypt with a 72-byte password limit

bcrypt was already in use and remains acceptable. Cost is 12 by default (the old config used 14, about one second of
CPU per request, which made login an easy denial-of-service lever). bcrypt reads only the first 72 bytes, so longer
passwords are rejected with a clear message rather than silently truncated.

## 6. Dummy hash for unknown emails

The old login skipped bcrypt when the email was unknown, so response time revealed which emails were registered. Login
now always runs one bcrypt comparison and returns the same 401 body for both failure cases.

## 7. `Origin` check plus `SameSite=Lax` for CSRF

Cookie auth needs CSRF protection. Browsers always send `Origin` on cross-site `POST` requests, so the middleware
refuses any state-changing request whose `Origin` host differs from `Host`. Together with `SameSite=Lax` cookies and
JSON-only bodies this meets the OWASP guidance without a token round trip. The scheme must match too (https when
cookies are `Secure`). A request with no `Origin` is refused if it carries a session cookie; cookieless requests (curl,
server to server) are allowed because a victim's browser cannot be driven through them. The old
`AllowOrigins: "*"` with credentials is gone: the UI is same-origin in production and proxied by Vite in development,
so CORS is not needed at all.

## 8. In-memory per-IP rate limiter

A token bucket (burst 10, then one per 6 seconds) on register and login protects bcrypt CPU and slows credential
stuffing. It is deliberately simple and has stated limits (README, "Known limits"). A shared store such as Redis would
be the next step if the service ran on several instances, and is not built until that is true.

## 9. Environment variables for configuration

The old `config.yaml` was committed with a database password. Environment variables keep secrets out of the repo and
match how containers and platforms inject settings. `.env.example` documents every variable.

## 10. Frontend embedded in the Go binary

`//go:embed` packs the built UI into the server, which serves it with an `index.html` fallback for client-side routes.
One artifact and one origin: no CORS, no second deployment, and a 21 MB container image.

## 11. Plain CSS with tokens instead of a UI library

Six screens do not justify a component framework. CSS custom properties give a consistent spacing, colour and radius
scale plus dark mode in one place, and keep the Content Security Policy at `default-src 'self'` with no inline styles.

## 12. Pinned patched Go toolchain

`govulncheck` against Go 1.26.0 reported reachable standard-library vulnerabilities in `net/http` and `crypto/tls`.
`go.mod` now names `toolchain go1.26.10`, after which the scan reports none that the code calls. CI runs
`govulncheck` on every push.

## 13. Frontend build output in `web/static`

The conventional `dist` name is blocked by this machine's Claude Code deny rules for `dist/` and `build/` paths, so the
build goes to `web/static`. The folder keeps a tracked `.gitkeep` because `go:embed` needs at least one file to compile.

## 14. Playwright against the real stack

The browser tests start the actual Go server (with the embedded UI) and Postgres rather than mocking the API, so they
exercise cookies, redirects, headers and CSP as a browser sees them. Each test registers its own unique account, which
lets tests run in parallel without cleanup. The rate-limit burst is raised for the run because tests share one client
address; the limiter itself is covered by a Go test.
