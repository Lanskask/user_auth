.PHONY: help setup run run-api test test-web test-e2e screenshots lint build web db-up db-down docker-up docker-down \
	tf-init tf-plan tf-apply tf-destroy

.DEFAULT_GOAL := help

help: ## list the commands
	@grep -E '^[a-zA-Z0-9_-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-12s %s\n", $$1, $$2}'

COMPOSE := $(shell docker compose version >/dev/null 2>&1 && echo "docker compose" || echo docker-compose)
DB_URL ?= postgres://postgres:postgres@localhost:5432/user_auth?sslmode=disable

setup: ## install frontend deps and start Postgres
	cd web && npm ci
	$(MAKE) db-up

db-up: ## start local Postgres
	$(COMPOSE) up -d --wait db

db-down: ## stop local Postgres
	$(COMPOSE) down

run: web ## build the frontend, then run the single-binary server against local Postgres
	AUTH_DATABASE_URL=$(DB_URL) AUTH_COOKIE_SECURE=false go run ./cmd/server

run-api: ## Go API only, on :8080 (pair with `cd web && npm run dev`)
	AUTH_DATABASE_URL=$(DB_URL) AUTH_COOKIE_SECURE=false go run ./cmd/server

test: ## Go tests (integration tests use TEST_DATABASE_URL, default is the local compose DB)
	TEST_DATABASE_URL=$(DB_URL) go test -race -cover ./...

test-web: ## Vitest component and unit tests
	cd web && npm test

test-e2e: web db-up ## Playwright browser tests against the real server (builds the UI, starts it on :8081)
	cd web && npx playwright install chromium && DB_URL="$(DB_URL)" npm run e2e

lint: ## go vet, golangci-lint, ESLint, TypeScript
	go vet ./...
	golangci-lint run
	cd web && npm run lint && npm run typecheck

web: ## build the frontend into web/static
	cd web && npm run build

build: web ## build bin/server with the frontend embedded
	CGO_ENABLED=0 go build -trimpath -o bin/server ./cmd/server

docker-up: ## app and Postgres in containers on :8080
	$(COMPOSE) --profile app up --build -d --wait

docker-down: ## stop the containers and drop their volumes
	$(COMPOSE) --profile app down -v
screenshots: ## regenerate docs/screenshots (needs a server on :8080, e.g. make run-api after make web)
	cd web && node scripts/screenshots.mjs

# Terraform. Needs NEON_API_KEY and `gcloud auth application-default login`. Neon first: Cloud Run reads its output.
TF_NEON := infra/neon
TF_RUN := infra/cloudrun
# tf-apply answers yes by default; run `make tf-apply APPROVE=` to be asked. tf-destroy always asks.
APPROVE ?= -auto-approve
TF_DB_URL = TF_VAR_database_url="$$(terraform -chdir=$(TF_NEON) output -raw database_url)"

tf-init: ## terraform init for Neon and Cloud Run
	terraform -chdir=$(TF_NEON) init
	terraform -chdir=$(TF_RUN) init

tf-plan: ## show what tf-apply would change (apply Neon once before planning Cloud Run)
	terraform -chdir=$(TF_NEON) plan
	$(TF_DB_URL) terraform -chdir=$(TF_RUN) plan

tf-apply: ## create or update Neon Postgres, then the Cloud Run service (creates cloud resources)
	terraform -chdir=$(TF_NEON) apply $(APPROVE)
	$(TF_DB_URL) terraform -chdir=$(TF_RUN) apply $(APPROVE)

tf-destroy: ## tear down Cloud Run first, then Neon
	$(TF_DB_URL) terraform -chdir=$(TF_RUN) destroy
	terraform -chdir=$(TF_NEON) destroy
