.DEFAULT_GOAL := help
PROJECT_SCRIPT := ./scripts/project.sh

.PHONY: help install run run-backend run-backend-dummy build test test-unit test-integration lint lint-fix format clean dev-up dev-down dev-recreate test-up test-down test-recreate

help: ## List available project commands
	@$(PROJECT_SCRIPT) $@

install: ## Install frontend dependencies and download Go modules
	@$(PROJECT_SCRIPT) $@

run: ## Start the frontend dev script, if defined
	@$(PROJECT_SCRIPT) $@

run-backend: ## Start the Go backend
	@$(PROJECT_SCRIPT) $@

run-backend-dummy: ## Start the Go backend with dummy AWS credentials for local use
	AWS_ACCESS_KEY_ID=dummy AWS_SECRET_ACCESS_KEY=dummy AWS_REGION=us-east-1 $(MAKE) run-backend

build: ## Build frontend and Go packages
	@$(PROJECT_SCRIPT) $@

test: test-unit test-integration ## Run unit and Docker-backed integration tests

test-unit: ## Run frontend and Go unit tests
	@$(PROJECT_SCRIPT) $@

test-integration: ## Run Go integration tests against DynamoDB Local
	@$(PROJECT_SCRIPT) $@

lint: ## Run frontend ESLint/TypeScript checks and Go vet/format checks
	@$(PROJECT_SCRIPT) $@

lint-fix: ## Apply frontend ESLint autofixes
	@$(PROJECT_SCRIPT) $@

format: ## Run the frontend formatter and gofmt
	@$(PROJECT_SCRIPT) $@

clean: ## Run app clean scripts and go clean
	@$(PROJECT_SCRIPT) $@

dev-up: ## Start dev DynamoDB Local and seed demo data
	@$(PROJECT_SCRIPT) $@

dev-down: ## Stop dev containers and preserve local data
	@$(PROJECT_SCRIPT) $@

dev-recreate: ## Recreate dev containers and seed missing tables
	@$(PROJECT_SCRIPT) $@

test-up: ## Start test DynamoDB Local
	@$(PROJECT_SCRIPT) $@

test-down: ## Stop test DynamoDB Local
	@$(PROJECT_SCRIPT) $@

test-recreate: ## Recreate test DynamoDB Local
	@$(PROJECT_SCRIPT) $@
