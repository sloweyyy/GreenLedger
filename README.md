# 🌱 GreenLedger

A comprehensive carbon footprint tracking and carbon credit management platform built with Go microservices architecture.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker)](https://www.docker.com/)
[![Microservices](https://img.shields.io/badge/Architecture-Microservices-FF6B6B?style=for-the-badge)](https://microservices.io/)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)

## 🚀 Features

### 🧮 **Carbon Footprint Calculator**
- Calculate CO2 emissions for vehicles, electricity, flights, and more
- Support for multiple emission factors and methodologies
- Real-time calculations with caching for performance

### 📊 **Activity Tracker**
- Track eco-friendly activities (biking, walking, recycling, etc.)
- Earn carbon credits for verified activities
- IoT device integration for automatic tracking
- Challenge and leaderboard systems

### 💰 **Carbon Credit Wallet**
- Manage carbon credit balances
- Transfer credits between users
- Transaction history and analytics
- Atomic transaction processing

### 🔐 **User Authentication & Authorization**
- JWT-based authentication
- Role-based access control (RBAC)
- Session management
- Password reset and email verification

### 📈 **Advanced Reporting**
- Generate PDF, CSV, and JSON reports
- Carbon footprint analysis
- Credit earning and spending reports
- Scheduled report generation

### 🏆 **Certificate Management**
- Issue blockchain-verified carbon offset certificates
- Certificate verification and validation
- NFT-based certificates
- Certificate marketplace (planned)

## 🏗️ Architecture

GreenLedger follows a **microservices architecture** with event-driven communication:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Calculator    │    │    Tracker      │    │     Wallet      │
│   Service       │    │    Service      │    │    Service      │
│   :8081         │    │    :8082        │    │    :8083        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
         ┌─────────────────┐    ┌─┴───────────────┐    ┌─────────────────┐
         │   User Auth     │    │     Kafka       │    │   Reporting     │
         │   Service       │    │   Event Bus     │    │   Service       │
         │   :8084         │    │   :9092         │    │   :8085         │
         └─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 🛠️ **Services Overview**

| Service | Port | Purpose | Database |
|---------|------|---------|----------|
| **Calculator** | 8081 | Carbon footprint calculations | calculator_db |
| **Tracker** | 8082 | Eco-activity tracking | tracker_db |
| **Wallet** | 8083 | Carbon credit management | wallet_db |
| **User Auth** | 8084 | Authentication & authorization | userauth_db |
| **Reporting** | 8085 | Report generation | reporting_db |
| **Certificate** | 8086 | Certificate management | certifier_db |

## 🔧 Technology Stack

### **Backend**
- **Language**: Go 1.21+
- **Framework**: Gin (HTTP), gRPC (inter-service)
- **Authentication**: JWT with RBAC
- **Validation**: Go Playground Validator

### **Data Layer**
- **Database**: PostgreSQL 15 (one per service)
- **ORM**: GORM with migrations
- **Cache**: Redis 7
- **Message Queue**: Apache Kafka

### **Infrastructure**
- **Containerization**: Docker & Docker Compose
- **API Gateway**: Traefik v3
- **Monitoring**: Prometheus + Grafana
- **Load Balancing**: Built-in with Traefik

### **Development**
- **Testing**: Testify, Load testing framework
- **Documentation**: Swagger/OpenAPI
- **Logging**: Structured logging with slog
- **Metrics**: Prometheus metrics

## 🚀 Quick Start

### Prerequisites

- **Docker** 20.10+ and **Docker Compose** v2
- **Go** 1.21+ (for development)
- **Make** (optional, for convenience commands)

### 🐳 Running with Docker Compose

1. **Clone the repository**:
```bash
git clone https://github.com/sloweyyy/GreenLedger.git
cd GreenLedger
```

2. **Start all services**:
```bash
docker-compose up -d
```

3. **Check service health**:
```bash
docker-compose ps
```

4. **View logs**:
```bash
docker-compose logs -f calculator-service
```

### 🌐 Access Points

| Service | URL | Description |
|---------|-----|-------------|
| **API Gateway** | http://localhost:8080 | Main API entry point |
| **Traefik Dashboard** | http://localhost:8090 | Load balancer dashboard |
| **Prometheus** | http://localhost:9090 | Metrics and monitoring |
| **Grafana** | http://localhost:3000 | Dashboards (admin/admin) |

### 📚 API Documentation

Each service provides Swagger documentation:

- **Calculator**: http://localhost:8081/swagger/index.html
- **Tracker**: http://localhost:8082/swagger/index.html
- **Wallet**: http://localhost:8083/swagger/index.html
- **User Auth**: http://localhost:8084/swagger/index.html
- **Reporting**: http://localhost:8085/swagger/index.html

## 💻 Development

### Local Development Setup

1. **Install dependencies**:
```bash
go mod download
cd services/calculator && go mod download
cd ../tracker && go mod download
cd ../wallet && go mod download
cd ../user-auth && go mod download
cd ../reporting && go mod download
```

2. **Set up environment variables**:
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. **Start infrastructure services**:
```bash
docker-compose up -d postgres-calculator postgres-tracker postgres-wallet postgres-userauth redis kafka
```

4. **Run database migrations**:
```bash
make migrate-up
```

5. **Start individual services**:
```bash
# Terminal 1 - Calculator Service
cd services/calculator && go run cmd/main.go

# Terminal 2 - Tracker Service
cd services/tracker && go run cmd/main.go

# Terminal 3 - Wallet Service
cd services/wallet && go run cmd/main.go

# Terminal 4 - User Auth Service
cd services/user-auth && go run cmd/main.go
```

### 🧪 Testing

**Run all tests**:
```bash
make test
```

**Run tests with coverage**:
```bash
make test-coverage
```

**Run load tests**:
```bash
make load-test
```

**Run specific service tests**:
```bash
cd services/calculator && go test ./...
```

### 📊 Monitoring & Metrics

The system includes comprehensive monitoring:

- **Prometheus Metrics**: HTTP requests, database queries, business metrics
- **Grafana Dashboards**: Service health, performance, business KPIs
- **Health Checks**: All services expose `/health` endpoints
- **Structured Logging**: JSON logs with correlation IDs

### 🔄 Event-Driven Architecture

Services communicate via Kafka events:

```
Activity Logged → Credits Earned → Wallet Updated → Report Generated
```

**Key Events**:
- `credit_earned`: When user completes eco-activities
- `balance_updated`: When wallet balance changes
- `transfer_completed`: When credits are transferred
- `certificate_issued`: When certificates are generated

## 🛡️ Security Features

- **JWT Authentication** with configurable expiration
- **Role-Based Access Control** (Admin, User, Moderator)
- **Input Validation** on all endpoints
- **SQL Injection Protection** via GORM
- **CORS Configuration** for web clients
- **Rate Limiting** (configurable)
- **Secure Headers** middleware

## 📈 Performance Features

- **Database Connection Pooling**
- **Redis Caching** for frequently accessed data
- **Horizontal Scaling** ready
- **Load Balancing** with Traefik
- **Async Processing** with Kafka
- **Database Indexing** for optimal queries

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Workflow

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Make** your changes
4. **Add** tests for new functionality
5. **Ensure** all tests pass (`make test`)
6. **Commit** your changes (`git commit -m 'Add amazing feature'`)
7. **Push** to the branch (`git push origin feature/amazing-feature`)
8. **Open** a Pull Request

### Code Standards

- Follow **Go best practices**
- Write **comprehensive tests** (>80% coverage)
- Add **Swagger documentation** for new endpoints
- Use **structured logging**
- Follow **conventional commits**

## 📄 License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **Go Community** for excellent tooling and libraries
- **GORM** for the fantastic ORM
- **Gin** for the lightweight web framework
- **Prometheus** and **Grafana** for monitoring solutions
- **Docker** for containerization platform

---

**Built with ❤️ for a sustainable future** 🌍
