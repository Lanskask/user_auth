.PHONY: setup run dev test test-web lint build web db-up db-down docker-up docker-down

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

dev: ## Go API on :8080 plus Vite dev server on :5173 (proxying /api)
	@echo "Run in two terminals: 'make run-api' and 'cd web && npm run dev'"

run-api:
	AUTH_DATABASE_URL=$(DB_URL) AUTH_COOKIE_SECURE=false go run ./cmd/server

test: ## Go tests (integration tests use TEST_DATABASE_URL, default is the local compose DB)
	TEST_DATABASE_URL=$(DB_URL) go test -race -cover ./...

test-web:
	cd web && npm test

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
