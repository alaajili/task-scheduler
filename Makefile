.PHONY: help setup dev test clean docker-up docker-down migrate-up migrate-down

# Variables
DOCKER_COMPOSE = docker-compose
MIGRATE = migrate
DB_URL = postgres://postgres:postgres@localhost:5432/taskscheduler?sslmode=disable
DB_TEST_URL = postgres://postgres:postgres@localhost:5432/taskscheduler_test?sslmode=disable

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

setup: ## Set up the development environment
	@echo "Installing dependencies..."
	@cd shared && go mod download
#	@cd api-server && go mod download
#	@cd scheduler && go mod download
#	@cd worker && go mod download
	@echo "Setup complete!"

docker-up: ## Start all Docker containers
	@$(DOCKER_COMPOSE) up -d
	@echo "Waiting for services to start..."
	@sleep 5
	@$(docker-compose) ps

docker-down: ## Stop all Docker containers
	@$(DOCKER_COMPOSE) down
	@echo "All services stopped."

docker-clean: ## Remove all Docker containers and volumes
	@$(DOCKER_COMPOSE) down -v
	@echo "All services and volumes removed."

migrate-up: ## Apply database migrations
	@echo "Applying database migrations..."
	@$(MIGRATE) -path ./migrations -database "$(DB_URL)" up
	@echo "Migrations applied."

migrate-test-up: ## Apply database migrations to the test database
	@echo "Applying database migrations to test database..."
	@$(MIGRATE) -path ./migrations -database "$(DB_TEST_URL)" up
	@echo "Migrations applied to test database."

migrate-down: ## Rollback the last database migration
	@echo "Rolling back the last database migration..."
	@$(MIGRATE) -path ./migrations -database "$(DB_URL)" down
	@echo "Last migration rolled back."

migrate-test-down: ## Rollback the last database migration in the test database
	@echo "Rolling back the last database migration in test database..."
	@$(MIGRATE) -path ./migrations -database "$(DB_TEST_URL)" down
	@echo "Last migration rolled back in test database."

migrate-create: ## Create a new database migration. Usage: make migrate-create NAME=<migration_name>
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME variable is required. Usage: make migrate-create NAME=<migration_name>"; \
		exit 1; \
	fi
	@echo "Creating new migration: $(NAME)"
	@$(MIGRATE) create -ext sql -dir ./migrations -seq $(NAME)

test: ## Run all tests
	@echo "Running tests..."
	@cd shared && go test -v -race -coverprofile=coverage.out ./...
	@cd api-server && go test -v -race -coverprofile=coverage.out ./...
	@cd scheduler && go test -v -race -coverprofile=coverage.out ./...
	@cd worker && go test -v -race -coverprofile=coverage.out ./...
	@echo "Tests completed."

test-coverage: ## Generate test coverage report
	@echo "Generating test coverage report..."
	@cd shared && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

lint: ## Run linters on the codebase
	@echo "Running linters..."
	@cd shared && golangci-lint run
	@cd api-server && golangci-lint run
	@cd scheduler && golangci-lint run
	@cd worker && golangci-lint run
	@echo "Linting completed."

fmt: ## Format the codebase
	@echo "Formatting code..."
	@find . -name '*.go' -not -path "./vendor/*" -exec gofmt -w {} \;
	@echo "Code formatted."

build-api: ## Build the API server
	@echo "Building API server..."
	@cd api-server && go build -o ../bin/api-server ./cmd/main.go
	@echo "API server built at ./bin/api-server"

build-scheduler: ## Build the Scheduler service
	@echo "Building Scheduler service..."
	@cd scheduler && go build -o ../bin/scheduler ./cmd/main.go
	@echo "Scheduler service built at ./bin/scheduler"

build-worker: ## Build the Worker service
	@echo "Building Worker service..."
	@cd worker && go build -o ../bin/worker ./cmd/main.go
	@echo "Worker service built at ./bin/worker"

build-all: build-api build-scheduler build-worker ## Build all services

run-api: ## Run the API server
	@echo "Running API server..."
	@cd api-server && go run ./cmd/main.go

run-scheduler: ## Run the Scheduler service
	@echo "Running Scheduler service..."
	@cd scheduler && go run ./cmd/main.go

run-worker: ## Run the Worker service
	@echo "Running Worker service..."
	@cd worker && go run ./cmd/main.go

dev: docker-up migrate-up ## Start development environment with Docker and apply migrations

clean: ## Clean build artifacts and temporary files
	@echo "Cleaning..."
	@rm -rf ./bin/*
	@rm -rf */coverage.out
	@rm -rf */coverage.html
	@echo "Clean complete."

.DEFAULT_GOAL := help
