ifneq (,$(wildcard .env))
  include .env
  export
endif

CMD_DIR := cmd/app


clean: ## Remove build artifacts and cache
	@echo "🧹 Cleaning up..."
	@rm -rf bin/ *.out *.exe *.test
	go clean



run: ## Run the app
	@echo "🚀 Running app:"
	go run $(CMD_DIR)/main.go



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


test-log: ## Run all tests in the project, including showing logs
	go test -v ./... 



build: ## Build the application binary
	@echo "🔨 Building application..."
	CGO_ENABLED=0 go build -o bin/push-service ./cmd/app/main.go



DOCKER_COMPOSE := docker compose -f ../docker-compose.yml


docker-up: ## Start push-service with dependencies
	@echo "🐳 Starting push-service with dependencies..."
	cd .. && docker compose up -d postgres redis rabbitmq push-service

docker-down: ## Stop all services
	@echo "🛑 Stopping all services..."
	cd .. && docker compose down


docker-logs: ## Show logs from push-service only
	cd .. && docker compose logs -f push-service

docker-restart: ## Restart push-service
	@echo "🔄 Restarting push-service..."
	cd .. && docker compose restart push-service

docker-rebuild: ## Rebuild and restart push-service
	@echo "🔨 Rebuilding push-service..."
	cd .. && docker compose up -d --build push-service

docker-stop: ## Stop only push-service
	@echo "🛑 Stopping push-service..."
	cd .. && docker compose stop push-service

docker-ps: ## Show running containers
	cd .. && docker compose ps

docker-shell: ## Open shell in push-service container
	@echo "🐚 Opening shell in push-service container..."
	cd .. && docker compose exec push-service sh

# Full stack testing
docker-test: docker-rebuild ## Rebuild service and show logs for testing
	@echo "✅ Service rebuilt. Waiting for startup..."
	@sleep 5
	@echo "📋 Service is ready at http://localhost:8081"
	@echo "📋 Health check: curl http://localhost:8081/health"
	@cd .. && docker compose logs -f push-service


# --- Documentation ---
help: ## Show this help message
	@awk 'BEGIN {FS = ":.*?## "}; /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort


.PHONY: test test-force test-function run tidy help clean test-log migrate-up migrate-down migrate-version migrate-force migrate-steps build docker-build docker-run docker-up docker-down docker-logs docker-logs-push docker-restart docker-rebuild docker-clean docker-ps docker-test