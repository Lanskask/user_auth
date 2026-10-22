import { defineConfig, devices } from "@playwright/test";

const PORT = 8081; // not 8080, so the tests never collide with a developer's running server
const DB_URL =
  process.env.DB_URL ?? "postgres://postgres:postgres@localhost:5432/user_auth?sslmode=disable";

export default defineConfig({
  testDir: "e2e",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? [["github"], ["html", { open: "never" }]] : "list",
  use: { baseURL: `http://localhost:${PORT}`, trace: "retain-on-failure" },
  projects: [{ name: "chromium", use: { ...devices["Desktop Chrome"] } }],
  // Tests run against the real single binary: Go server + embedded UI (run `npm run build` first) + Postgres.
  webServer: {
    command: "go run ./cmd/server",
    cwd: "..",
    url: `http://localhost:${PORT}/readyz`,
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
    env: {
      AUTH_ADDR: `:${PORT}`,
      AUTH_DATABASE_URL: DB_URL,
      AUTH_COOKIE_SECURE: "false",
      AUTH_BCRYPT_COST: "10",
      AUTH_RATE_LIMIT_BURST: "1000", // parallel tests share one client IP; the limiter itself is unit-tested in Go
    },
  },
});
