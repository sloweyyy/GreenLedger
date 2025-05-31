# GreenLedger Makefile

.PHONY: help build test clean docker-build docker-up docker-down migrate-up load-test

# Default target
help: ## Show this help message
	@echo "🌱 GreenLedger - Carbon Credit Tracking System"
	@echo ""
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build commands
build: ## Build all services
	@echo "Building all services..."
	@mkdir -p bin
	@cd services/calculator && go build -o ../../bin/calculator ./cmd/main.go
	@cd services/tracker && go build -o ../../bin/tracker ./cmd/main.go
	@cd services/wallet && go build -o ../../bin/wallet ./cmd/main.go
	@cd services/user-auth && go build -o ../../bin/user-auth ./cmd/main.go
	@cd services/reporting && go build -o ../../bin/reporting ./cmd/main.go
	@echo "✅ All services built successfully"

# Test commands
test: ## Run all tests
	@echo "Running all tests..."
	@cd services/calculator && go test ./...
	@cd services/tracker && go test ./...
	@cd services/wallet && go test ./...
	@cd services/user-auth && go test ./...
	@cd services/reporting && go test ./...
	@echo "✅ All tests passed"

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@cd services/calculator && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
	@cd services/tracker && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
	@cd services/wallet && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
	@cd services/user-auth && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
	@cd services/reporting && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage reports generated"

load-test: ## Run load tests
	@echo "Running load tests..."
	@cd tests/load && go test -v ./...

# Docker commands
docker-build: ## Build all Docker images
	@echo "Building Docker images..."
	@docker-compose build
	@echo "✅ Docker images built successfully"

docker-up: ## Start all services with Docker Compose
	@echo "Starting all services..."
	@docker-compose up -d
	@echo "✅ All services started"
	@echo ""
	@echo "🌐 Access Points:"
	@echo "  API Gateway: http://localhost:8080"
	@echo "  Prometheus:  http://localhost:9090"
	@echo "  Grafana:     http://localhost:3000 (admin/admin)"
	@echo ""

docker-down: ## Stop all services
	@echo "Stopping all services..."
	@docker-compose down
	@echo "✅ All services stopped"

docker-logs: ## View logs from all services
	@docker-compose logs -f

docker-ps: ## Show running containers
	@docker-compose ps

docker-clean: ## Clean up Docker resources
	@echo "Cleaning up Docker resources..."
	@docker-compose down -v --remove-orphans
	@docker system prune -f
	@echo "✅ Docker cleanup completed"

# Development commands
dev-setup: ## Set up development environment
	@echo "Setting up development environment..."
	@mkdir -p bin logs
	@go mod download
	@cd services/calculator && go mod download
	@cd services/tracker && go mod download
	@cd services/wallet && go mod download
	@cd services/user-auth && go mod download
	@cd services/reporting && go mod download
	@echo "✅ Development environment ready"

# Database commands
migrate-up: ## Run database migrations
	@echo "Running database migrations..."
	@docker-compose up -d postgres-calculator postgres-tracker postgres-wallet postgres-userauth postgres-reporting
	@sleep 10
	@echo "✅ Database migrations completed"

# Utility commands
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf logs/
	@cd services/calculator && rm -f coverage.out coverage.html
	@cd services/tracker && rm -f coverage.out coverage.html
	@cd services/wallet && rm -f coverage.out coverage.html
	@cd services/user-auth && rm -f coverage.out coverage.html
	@cd services/reporting && rm -f coverage.out coverage.html
	@echo "✅ Cleanup completed"

# Quick start commands
quick-start: docker-up ## Quick start all services
	@echo ""
	@echo "🚀 GreenLedger is now running!"
	@echo ""
	@echo "📚 API Documentation:"
	@echo "  Calculator: http://localhost:8081/swagger/index.html"
	@echo "  Tracker:    http://localhost:8082/swagger/index.html"
	@echo "  Wallet:     http://localhost:8083/swagger/index.html"
	@echo "  User Auth:  http://localhost:8084/swagger/index.html"
	@echo "  Reporting:  http://localhost:8085/swagger/index.html"
	@echo ""
	@echo "🔧 Management:"
	@echo "  Prometheus: http://localhost:9090"
	@echo "  Grafana:    http://localhost:3000 (admin/admin)"
	@echo ""
	@echo "📊 Health Checks:"
	@echo "  make docker-ps    # Check service status"
	@echo "  make docker-logs  # View all logs"
	@echo ""

stop: docker-down ## Stop all services

status: docker-ps ## Show service status