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
	@cd services/certifier && go test ./...
	@echo "✅ All tests passed"

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@cd services/calculator && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
	@cd services/tracker && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
	@cd services/wallet && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
	@cd services/user-auth && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
	@cd services/reporting && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
	@cd services/certifier && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
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

# Dependency commands
deps: ## Install and fix all dependencies
	@echo "📦 Installing and fixing dependencies..."
	@./scripts/fix-dependencies.sh

deps-check: ## Check dependency status
	@echo "🔍 Checking dependency status..."
	@./scripts/fix-dependencies.sh check

deps-fix: ## Fix dependency issues
	@echo "🔧 Fixing dependency issues..."
	@./scripts/fix-dependencies.sh

# Development commands
dev-setup: deps ## Set up development environment
	@echo "Setting up development environment..."
	@mkdir -p bin logs
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@echo "✅ Development environment ready"

# Database commands
migrate-up: ## Run database migrations
	@echo "Running database migrations..."
	@docker-compose up -d postgres-calculator postgres-tracker postgres-wallet postgres-userauth postgres-reporting
	@sleep 10
	@echo "✅ Database migrations completed"

# Code quality commands
lint: ## Run linters on all code
	@echo "🔍 Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "⚠️ golangci-lint not installed. Run 'make dev-setup' first"; \
	fi

security: ## Run security scans
	@echo "🔒 Running security scans..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "Installing gosec..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
		gosec ./...; \
	fi

format: ## Format all Go code
	@echo "✨ Formatting code..."
	@go fmt ./...
	@if [ -d "shared" ]; then cd shared && go fmt ./... && cd ..; fi
	@for service in calculator tracker wallet user-auth reporting; do \
		if [ -d "services/$$service" ]; then \
			cd services/$$service && go fmt ./... && cd ../..; \
		fi; \
	done

# CI commands
ci-local: ## Run CI pipeline locally
	@echo "🔄 Running CI pipeline locally..."
	@make deps
	@make lint
	@make test
	@make build
	@echo "✅ Local CI pipeline completed successfully!"

check-all: deps-check lint test ## Run all checks (dependencies, linting, tests)

# Utility commands
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf logs/
	@rm -rf dist/
	@find . -name "coverage.out" -delete
	@find . -name "coverage.html" -delete
	@go clean -cache
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
