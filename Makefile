# GreenLedger Makefile

.PHONY: help build test clean docker-build docker-push run-all migrate-up migrate-down

# Default target
help:
	@echo "Available targets:"
	@echo "  build           - Build all services"
	@echo "  test            - Run all tests"
	@echo "  test-coverage   - Run tests with coverage"
	@echo "  clean           - Clean build artifacts"
	@echo "  docker-build    - Build all Docker images"
	@echo "  run-all         - Run all services locally"
	@echo "  migrate-up      - Run database migrations"
	@echo "  migrate-down    - Rollback database migrations"
	@echo "  proto-gen       - Generate protobuf files"

# Build targets
build:
	@echo "Building all services..."
	@cd services/calculator && go build -o ../../bin/calculator ./cmd/main.go
	@cd services/tracker && go build -o ../../bin/tracker ./cmd/main.go
	@cd services/wallet && go build -o ../../bin/wallet ./cmd/main.go
	@cd services/user-auth && go build -o ../../bin/user-auth ./cmd/main.go
	@cd services/reporting && go build -o ../../bin/reporting ./cmd/main.go
	@cd services/certifier && go build -o ../../bin/certifier ./cmd/main.go

# Test targets
test:
	@echo "Running tests..."
	@go test ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

test-integration:
	@echo "Running integration tests..."
	@docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit
	@docker-compose -f docker-compose.test.yml down

# Clean targets
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

# Docker targets
docker-build:
	@echo "Building Docker images..."
	@docker build -t greenledger/calculator:latest -f services/calculator/Dockerfile .
	@docker build -t greenledger/tracker:latest -f services/tracker/Dockerfile .
	@docker build -t greenledger/wallet:latest -f services/wallet/Dockerfile .
	@docker build -t greenledger/user-auth:latest -f services/user-auth/Dockerfile .
	@docker build -t greenledger/reporting:latest -f services/reporting/Dockerfile .
	@docker build -t greenledger/certifier:latest -f services/certifier/Dockerfile .

docker-build-calculator:
	@docker build -t greenledger/calculator:latest -f services/calculator/Dockerfile .

docker-build-tracker:
	@docker build -t greenledger/tracker:latest -f services/tracker/Dockerfile .

docker-build-wallet:
	@docker build -t greenledger/wallet:latest -f services/wallet/Dockerfile .

# Development targets
run-all:
	@echo "Starting all services..."
	@docker-compose up -d

run-calculator:
	@cd services/calculator && go run cmd/main.go

run-tracker:
	@cd services/tracker && go run cmd/main.go

run-wallet:
	@cd services/wallet && go run cmd/main.go

# Database migration targets
migrate-up:
	@echo "Running database migrations..."
	@cd services/calculator && migrate -path migrations -database "postgres://postgres:password@localhost:5432/calculator_db?sslmode=disable" up
	@cd services/tracker && migrate -path migrations -database "postgres://postgres:password@localhost:5433/tracker_db?sslmode=disable" up
	@cd services/wallet && migrate -path migrations -database "postgres://postgres:password@localhost:5434/wallet_db?sslmode=disable" up
	@cd services/user-auth && migrate -path migrations -database "postgres://postgres:password@localhost:5435/userauth_db?sslmode=disable" up

migrate-down:
	@echo "Rolling back database migrations..."
	@cd services/calculator && migrate -path migrations -database "postgres://postgres:password@localhost:5432/calculator_db?sslmode=disable" down
	@cd services/tracker && migrate -path migrations -database "postgres://postgres:password@localhost:5433/tracker_db?sslmode=disable" down
	@cd services/wallet && migrate -path migrations -database "postgres://postgres:password@localhost:5434/wallet_db?sslmode=disable" down
	@cd services/user-auth && migrate -path migrations -database "postgres://postgres:password@localhost:5435/userauth_db?sslmode=disable" down

# Protocol Buffer generation
proto-gen:
	@echo "Generating protobuf files..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/*.proto

# Setup development environment
setup-dev:
	@echo "Setting up development environment..."
	@go mod download
	@docker-compose up -d postgres redis kafka
	@sleep 10
	@make migrate-up

# Lint code
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Generate swagger docs
swagger:
	@echo "Generating swagger documentation..."
	@cd services/calculator && swag init -g cmd/main.go
	@cd services/tracker && swag init -g cmd/main.go
	@cd services/wallet && swag init -g cmd/main.go
	@cd services/user-auth && swag init -g cmd/main.go
	@cd services/reporting && swag init -g cmd/main.go
