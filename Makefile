

DOCKER_COMPOSE := docker compose -f docker-compose.yml



docker-up: ## Start all services with docker-compose (PostgreSQL, Redis, RabbitMQ, API Gateway, User Service, Push Service)
	@echo "🐳 Starting all services with docker-compose..."
	docker compose up -d

docker-up-deps: ## Start only dependencies (PostgreSQL, Redis, RabbitMQ)
	@echo "🐳 Starting dependencies..."
	docker compose up -d postgres redis rabbitmq


docker-down: ## Stop all services
	@echo "🛑 Stopping all services..."
	docker compose down

docker-logs: ## Show logs from all services
	docker compose logs -f

docker-logs-push: ## Show logs from push-service only
	docker compose logs -f push-service

docker-restart: ## Restart
	@echo "🔄 Restarting push-service..."
	docker compose restart

docker-rebuild: ## Rebuild and restart push service
	@echo "🔨 Rebuilding push-service..."
	docker compose up -d --build push-service

docker-logs-push: ## Show logs from push service only
	docker compose logs -f push-service

docker-logs-user: ## Show logs from user service only
	docker compose logs -f user-service

docker-logs-gateway: ## Show logs from api gateway only
	docker compose logs -f api-gateway

docker-logs-email: ## Show logs from email service only
	docker compose logs -f email-service

docker-ps: ## Show running containers
	docker compose ps



# --- Documentation ---
help: ## Show this help message
	@awk 'BEGIN {FS = ":.*?## "}; /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort


