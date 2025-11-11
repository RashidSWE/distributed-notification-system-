ifneq (,$(wildcard .env))
  include .env
  export
endif

CMD_DIR := cmd/app


clean: ## Remove build artifacts and cache
	@echo "🧹 Cleaning up..."
	@rm -rf bin/ *.out *.exe *.test
	go clean


# Run the app
run: ## Run the app
	@echo "🚀 Running app:"
	go run $(CMD_DIR)/main.go


# --- Database Migrations ---
migrate-up: ## Run all migrations
	@echo "Running migrations..."
	go run cmd/migrate/main.go -command=up

migrate-down: ## Rollback one migration
	@echo "Rolling back migration..."
	go run cmd/migrate/main.go -command=down

migrate-version: ## Show current migration version
	go run cmd/migrate/main.go -command=version

migrate-force: ## Force migration version (Usage: make migrate-force VERSION=1)
	go run cmd/migrate/main.go -command=force -version=$(VERSION)

migrate-steps: ## Run n migration steps (Usage: make migrate-steps STEPS=1)
	go run cmd/migrate/main.go -command=steps -steps=$(STEPS)


# --- Tidy go.mod ---
tidy: ## Tidy go.mod and go.sum
	@echo "🧹 Tidying go.mod and go.sum..."
	go mod tidy



test: ## Run all tests
	go test ./... 


test-force: ## Run tests without caching
	go test -count=1 ./... 


test-race: ## Run tests with race condition detection
	go test -race ./...


test-ci: ## Run tests with both race detection and coverage (used in CI)
	go test -race -coverprofile=coverage.out ./... 
	go tool cover -func=coverage.out


test-function: ## Usage: make test-function TEST=TestGetAllJournals
	go test -v -run ^$(TEST)$$


# Run all tests in the project,
test-log: ## Run all tests in the project, including showing logs
	go test -v ./... 


# Build the application
build: ## Build the application binary
	@echo "🔨 Building application..."
	CGO_ENABLED=0 go build -o bin/push-service ./cmd/app/main.go


# Docker commands (run from root directory)
DOCKER_COMPOSE := docker compose -f ../docker-compose.yml

docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker build -t push-service:latest .

docker-run: ## Run Docker container (standalone)
	@echo "🐳 Running Docker container..."
	docker run --env-file .env -p 8081:8081 push-service:latest

docker-up: ## Start all services with docker-compose (PostgreSQL, Redis, RabbitMQ, Push Service)
	@echo "🐳 Starting all services with docker-compose..."
	cd .. && docker compose up -d

docker-down: ## Stop all services
	@echo "🛑 Stopping all services..."
	cd .. && docker compose down

docker-logs: ## Show logs from all services
	cd .. && docker compose logs -f

docker-logs-push: ## Show logs from push-service only
	cd .. && docker compose logs -f push-service

docker-restart: ## Restart push-service
	@echo "🔄 Restarting push-service..."
	cd .. && docker compose restart push-service

docker-rebuild: ## Rebuild and restart push-service
	@echo "🔨 Rebuilding push-service..."
	cd .. && docker compose up -d --build push-service

docker-clean: ## Remove all containers, volumes, and images
	@echo "🧹 Cleaning up Docker resources..."
	cd .. && docker compose down -v
	docker rmi push-service:latest 2>/dev/null || true

docker-ps: ## Show running containers
	cd .. && docker compose ps

# Full stack testing
docker-test: docker-rebuild ## Rebuild service and show logs for testing
	@echo "✅ Service rebuilt. Waiting for startup..."
	@sleep 3
	@echo "📋 Service is ready at http://localhost:8081"
	@echo "📋 RabbitMQ Management: http://localhost:15672 (guest/guest)"
	@echo "📋 Health check: curl http://localhost:8081/api/health"
	@docker-compose logs -f push-service


# --- Documentation ---
help: ## Show this help message
	@awk 'BEGIN {FS = ":.*?## "}; /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort


.PHONY: test test-force test-function run tidy help clean test-log migrate-up migrate-down migrate-version migrate-force migrate-steps build docker-build docker-run docker-up docker-down docker-logs docker-logs-push docker-restart docker-rebuild docker-clean docker-ps docker-test