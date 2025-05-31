# GreenLedger - Carbon Credit Tracking System

A comprehensive microservices-based carbon credit tracking system built with Golang that helps users calculate their carbon footprint, earn credits for eco-friendly activities, and manage their carbon credit portfolio through a scalable, event-driven architecture.

## Architecture Overview

GreenLedger consists of the following microservices:

- **Carbon Footprint Calculator Service**: Calculates CO₂ emissions from user activities
- **Activity Tracker Service**: Tracks eco-friendly activities and publishes events
- **Carbon Credit Wallet Service**: Manages user carbon credit balances and transactions
- **User Management & Authentication Service**: Handles user registration, authentication, and RBAC
- **Reporting Service**: Generates carbon footprint reports and analytics
- **Certificate & Verification Service**: Issues digital certificates and blockchain verification

## Project Structure

```
greenledger/
├── services/
│   ├── calculator/          # Carbon footprint calculation service
│   ├── tracker/            # Activity tracking service
│   ├── wallet/             # Carbon credit wallet service
│   ├── user-auth/          # User management and authentication
│   ├── reporting/          # Reporting and analytics service
│   └── certifier/          # Certificate and verification service
├── proto/                  # Protocol Buffer definitions
├── shared/                 # Shared Go packages and utilities
├── api-gateway/           # API Gateway configuration
├── k8s/                   # Kubernetes manifests
├── docker-compose.yml     # Local development setup
├── docs/                  # Documentation
└── scripts/               # Build and deployment scripts
```

## Technology Stack

- **Backend**: Go (Gin/Fiber for HTTP, gRPC for inter-service communication)
- **Databases**: PostgreSQL (per service), Redis (caching)
- **Message Queue**: Kafka or NATS
- **API Gateway**: Traefik or Kong
- **Monitoring**: Prometheus, Grafana
- **Containerization**: Docker, Kubernetes
- **Blockchain**: Celo testnet (for certificate verification)

## Quick Start

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- PostgreSQL
- Redis
- Kafka/NATS (optional for local development)

### Local Development Setup

1. Clone the repository:
```bash
git clone <repository-url>
cd greenledger
```

2. Start infrastructure services:
```bash
docker-compose up -d postgres redis kafka
```

3. Run database migrations:
```bash
make migrate-up
```

4. Start all services:
```bash
make run-all
```

5. Access the API Gateway:
```
http://localhost:8080
```

## API Documentation

- **Calculator Service**: http://localhost:8081/swagger
- **Tracker Service**: http://localhost:8082/swagger
- **Wallet Service**: http://localhost:8083/swagger
- **User Auth Service**: http://localhost:8084/swagger
- **Reporting Service**: http://localhost:8085/swagger

## Development

### Running Individual Services

```bash
# Calculator Service
cd services/calculator && go run main.go

# Activity Tracker Service
cd services/tracker && go run main.go

# Wallet Service
cd services/wallet && go run main.go
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run integration tests
make test-integration
```

### Building Docker Images

```bash
# Build all services
make docker-build

# Build specific service
make docker-build-calculator
```

## Deployment

### Kubernetes

```bash
# Deploy to Kubernetes
kubectl apply -f k8s/

# Check deployment status
kubectl get pods -n greenledger
```

### Docker Compose (Production)

```bash
docker-compose -f docker-compose.prod.yml up -d
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details
