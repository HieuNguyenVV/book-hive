.PHONY: help db-up db-down db-logs migrate-diff migrate-apply migrate-status migrate-hash migrate-lint migrate-docker migrate-validate swagger

ATLAS ?= atlas
COMPOSE ?= docker compose
ENV ?= local
NAME ?=
SWAG ?= swag

help: ## Show available targets
	@echo "Migration targets:"
	@echo "  make db-up              Start Postgres"
	@echo "  make db-down            Stop Postgres (+ migrator)"
	@echo "  make db-logs            Tail Postgres logs"
	@echo "  make migrate-diff name=<migration_name>"
	@echo "                          Generate migration from schema/schema.sql"
	@echo "  make migrate-apply      Apply pending migrations (local Atlas)"
	@echo "  make migrate-status     Show migration status"
	@echo "  make migrate-hash       Recalculate migrations/atlas.sum"
	@echo "  make migrate-lint       Lint pending migrations"
	@echo "  make migrate-validate   Validate migration directory integrity"
	@echo "  make migrate-docker     Build & run migrator via docker compose"

db-up: ## Start Postgres
	$(COMPOSE) up -d postgres

db-down: ## Stop compose services
	$(COMPOSE) down

db-logs: ## Tail Postgres logs
	$(COMPOSE) logs -f postgres

migrate-diff: ## Generate a new migration (requires name=...)
ifndef NAME
	$(error Usage: make migrate-diff name=<migration_name>)
endif
	$(ATLAS) migrate diff $(NAME) --env $(ENV)

migrate-apply: ## Apply pending migrations locally
	$(ATLAS) migrate apply --env $(ENV)

migrate-status: ## Show applied / pending migrations
	$(ATLAS) migrate status --env $(ENV)

migrate-hash: ## Update atlas.sum checksums
	$(ATLAS) migrate hash --env $(ENV)

migrate-lint: ## Lint migrations against current DB
	$(ATLAS) migrate lint --env $(ENV) --latest 1

migrate-validate: ## Validate migration files and atlas.sum
	$(ATLAS) migrate validate --env $(ENV)

migrate-docker: ## Apply migrations with the migrator container
	$(COMPOSE) up --build --abort-on-container-exit migrator

swagger: ## Generate Swagger docs from annotations
	$(SWAG) init -g cmd/server/main.go -o docs
