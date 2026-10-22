.PHONY: setup run run-api test test-web test-e2e screenshots lint build web db-up db-down docker-up docker-down

COMPOSE := $(shell docker compose version >/dev/null 2>&1 && echo "docker compose" || echo docker-compose)
DB_URL ?= postgres://postgres:postgres@localhost:5432/user_auth?sslmode=disable

setup: ## install frontend deps and start Postgres
	cd web && npm ci
	$(MAKE) db-up

db-up:
	$(COMPOSE) up -d --wait db

db-down:
	$(COMPOSE) down

run: web ## build the frontend, then run the single-binary server against local Postgres
	AUTH_DATABASE_URL=$(DB_URL) AUTH_COOKIE_SECURE=false go run ./cmd/server

run-api: ## Go API only, on :8080 (pair with `cd web && npm run dev`)
	AUTH_DATABASE_URL=$(DB_URL) AUTH_COOKIE_SECURE=false go run ./cmd/server

test: ## Go tests (integration tests use TEST_DATABASE_URL, default is the local compose DB)
	TEST_DATABASE_URL=$(DB_URL) go test -race -cover ./...

test-web:
	cd web && npm test

test-e2e: web db-up ## Playwright browser tests against the real server (builds the UI, starts it on :8081)
	cd web && npx playwright install chromium && DB_URL="$(DB_URL)" npm run e2e

lint:
	go vet ./...
	golangci-lint run
	cd web && npm run lint && npm run typecheck

web:
	cd web && npm run build

build: web
	CGO_ENABLED=0 go build -trimpath -o bin/server ./cmd/server

docker-up:
	$(COMPOSE) --profile app up --build -d --wait

docker-down:
	$(COMPOSE) --profile app down -v
screenshots: ## regenerate docs/screenshots (needs a server on :8080, e.g. make run-api after make web)
	cd web && node scripts/screenshots.mjs
