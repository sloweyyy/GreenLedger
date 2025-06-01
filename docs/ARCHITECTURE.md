# 🌱 GreenLedger Architecture Documentation

## Overview

GreenLedger is a comprehensive microservices-based carbon credit tracking system built with Golang. The system helps users calculate their carbon footprint, earn credits for eco-friendly activities, and manage their carbon credit portfolio through a scalable, event-driven architecture.

## System Architecture

### High-Level Architecture

```bash
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Client    │    │  Mobile Client  │    │  External APIs  │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────┴─────────────┐
                    │      API Gateway          │
                    │      (Traefik)           │
                    └─────────────┬─────────────┘
                                 │
        ┌────────────────────────┼────────────────────────┐
        │                       │                        │
┌───────▼───────┐    ┌──────────▼──────────┐    ┌────────▼────────┐
│   Calculator  │    │      Tracker        │    │     Wallet      │
│   Service     │    │      Service        │    │    Service      │
│   (Port 8081) │    │    (Port 8082)      │    │  (Port 8083)    │
└───────┬───────┘    └──────────┬──────────┘    └────────┬────────┘
        │                       │                        │
        │              ┌────────▼────────┐               │
        │              │     Message     │               │
        │              │     Queue       │               │
        │              │    (Kafka)      │               │
        │              └────────┬────────┘               │
        │                       │                        │
┌───────▼───────┐    ┌──────────▼──────────┐    ┌────────▼────────┐
│   User Auth   │    │    Reporting        │    │   Certificate   │
│   Service     │    │    Service          │    │   Service       │
│  (Port 8084)  │    │   (Port 8085)       │    │  (Port 8086)    │
└───────┬───────┘    └──────────┬──────────┘    └────────┬────────┘
        │                       │                        │
        └───────────────────────┼────────────────────────┘
                                │
                    ┌───────────▼───────────┐
                    │     Data Layer        │
                    │  PostgreSQL + Redis   │
                    └───────────────────────┘
```

### Service Breakdown

#### 1. Calculator Service (Port 8081)

**Purpose**: Calculate carbon footprint for various activities

**Responsibilities**:

- Accept activity data (vehicle travel, electricity usage, purchases, flights, heating)
- Apply emission factors to calculate CO₂ equivalents
- Store calculation history
- Provide calculation APIs (REST + gRPC)

**Database**: `calculator_db`

- Tables: `calculations`, `activities`, `emission_factors`

**Key APIs**:

- `POST /api/v1/calculator/calculate` - Calculate footprint
- `GET /api/v1/calculator/calculations` - Get calculation history
- `GET /api/v1/calculator/emission-factors` - Get emission factors

#### 2. Activity Tracker Service (Port 8082)

**Purpose**: Track eco-friendly activities and publish events

**Responsibilities**:

- Receive eco-friendly activity data (biking, public transit, renewable energy)
- Support webhook and MQTT/IoT payloads
- Validate and publish events to message queue
- Calculate carbon credits earned

**Database**: `tracker_db`

- Tables: `activities`, `activity_types`, `credit_rules`

**Key APIs**:

- `POST /api/v1/tracker/activities` - Log eco-friendly activity
- `GET /api/v1/tracker/activities` - Get activity history
- `POST /api/v1/tracker/webhook` - Webhook endpoint for IoT devices

#### 3. Carbon Credit Wallet Service (Port 8083)

**Purpose**: Manage user carbon credit balances and transactions

**Responsibilities**:

- Maintain user carbon credit balances
- Process credit earning, spending, and transfers
- Consume events from calculator and tracker services
- Provide transaction history and audit trails

**Database**: `wallet_db`

- Tables: `wallets`, `transactions`, `balance_history`

**Key APIs**:

- `GET /api/v1/wallet/balance` - Get user balance
- `POST /api/v1/wallet/transfer` - Transfer credits
- `GET /api/v1/wallet/transactions` - Get transaction history

#### 4. User Management & Authentication Service (Port 8084)

**Purpose**: Handle user registration, authentication, and authorization

**Responsibilities**:

- User registration and profile management
- JWT-based authentication
- Role-based access control (RBAC)
- OAuth2-compatible endpoints

**Database**: `userauth_db`

- Tables: `users`, `roles`, `user_roles`, `sessions`

**Key APIs**:

- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `GET /api/v1/auth/profile` - Get user profile
- `POST /api/v1/auth/refresh` - Refresh JWT token

#### 5. Reporting Service (Port 8085)

**Purpose**: Generate reports and analytics

**Responsibilities**:

- Aggregate data from all services
- Generate PDF and JSON reports
- Scheduled report generation
- Email/webhook delivery

**Database**: Reads from all service databases

- Tables: `reports`, `report_schedules`, `report_deliveries`

**Key APIs**:

- `GET /api/v1/reports/footprint` - Carbon footprint reports
- `GET /api/v1/reports/credits` - Credit earning reports
- `POST /api/v1/reports/schedule` - Schedule recurring reports

#### 6. Certificate & Verification Service (Port 8086) [Optional]

**Purpose**: Issue digital certificates and blockchain verification

**Responsibilities**:

- Issue certificates for carbon credit milestones
- Blockchain integration (Celo testnet)
- Smart contract interactions
- Certificate verification

**Database**: `certifier_db`

- Tables: `certificates`, `blockchain_transactions`, `verification_logs`

## Data Flow

### Carbon Footprint Calculation Flow

1. User submits activity data to Calculator Service
2. Calculator Service retrieves emission factors
3. Calculates CO₂ emissions and stores results
4. Publishes calculation event to message queue
5. Wallet Service consumes event (if applicable for offsets)

### Eco-Activity Tracking Flow

1. User/IoT device submits eco-activity to Tracker Service
2. Tracker Service validates and calculates credits earned
3. Publishes credit earning event to message queue
4. Wallet Service consumes event and credits user's balance
5. Certificate Service may issue milestone certificates

### Credit Transfer Flow

1. User initiates transfer via Wallet Service
2. Wallet Service validates balance and processes transfer
3. Updates both sender and receiver balances
4. Records transaction history
5. Publishes transfer event for audit/reporting

## Technology Stack

### Backend Services

- **Language**: Go 1.21+
- **HTTP Framework**: Gin
- **gRPC**: Protocol Buffers + gRPC-Go
- **Database**: PostgreSQL (per service)
- **Cache**: Redis
- **Message Queue**: Kafka
- **Authentication**: JWT

### Infrastructure

- **Containerization**: Docker
- **Orchestration**: Kubernetes
- **API Gateway**: Traefik
- **Monitoring**: Prometheus + Grafana
- **Logging**: Structured logging with slog

### External Integrations

- **Blockchain**: Celo testnet (for certificates)
- **Email**: SMTP/SendGrid (for reports)
- **IoT**: MQTT broker support

## Database Design

### Calculator Service Schema

```sql
-- Calculations table
CREATE TABLE calculations (
    id UUID PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    total_co2_kg DECIMAL(10,3) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE
);

-- Activities table
CREATE TABLE activities (
    id UUID PRIMARY KEY,
    calculation_id UUID REFERENCES calculations(id),
    activity_type VARCHAR(100) NOT NULL,
    co2_kg DECIMAL(10,3) NOT NULL,
    emission_factor DECIMAL(10,6) NOT NULL,
    factor_source VARCHAR(255) NOT NULL,
    activity_data JSONB,
    created_at TIMESTAMP WITH TIME ZONE
);

-- Emission factors table
CREATE TABLE emission_factors (
    id UUID PRIMARY KEY,
    activity_type VARCHAR(100) NOT NULL,
    sub_type VARCHAR(100) NOT NULL,
    factor_co2_per_unit DECIMAL(10,6) NOT NULL,
    unit VARCHAR(50) NOT NULL,
    source VARCHAR(255) NOT NULL,
    location VARCHAR(100),
    last_updated TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE
);
```

## Security Considerations

### Authentication & Authorization

- JWT tokens with configurable expiration
- Role-based access control (RBAC)
- API rate limiting via API Gateway
- HTTPS/TLS encryption for all communications

### Data Protection

- Database encryption at rest
- Sensitive data masking in logs
- Input validation and sanitization
- SQL injection prevention via ORM

### Network Security

- Service-to-service communication via gRPC with TLS
- Network segmentation via Docker networks
- Firewall rules for production deployment

## Monitoring & Observability

### Metrics (Prometheus)

- HTTP request metrics (latency, status codes, throughput)
- Database connection pool metrics
- Business metrics (calculations per day, credits earned)
- System metrics (CPU, memory, disk usage)

### Logging

- Structured JSON logging
- Request tracing with correlation IDs
- Error tracking and alerting
- Log aggregation and search

### Health Checks

- Service health endpoints
- Database connectivity checks
- Dependency health monitoring
- Kubernetes readiness/liveness probes

## Deployment

### Local Development

```bash
# Start infrastructure
docker-compose up -d postgres redis kafka

# Run migrations
make migrate-up

# Start services
make run-all
```

### Production (Kubernetes)

```bash
# Deploy to Kubernetes
kubectl apply -f k8s/

# Check deployment
kubectl get pods -n greenledger
```

## API Documentation

Each service exposes OpenAPI/Swagger documentation:

- Calculator: <http://localhost:8081/swagger>
- Tracker: <http://localhost:8082/swagger>
- Wallet: <http://localhost:8083/swagger>
- User Auth: <http://localhost:8084/swagger>
- Reporting: <http://localhost:8085/swagger>

## Testing Strategy

### Unit Tests

- Service layer business logic
- Repository layer data access
- Utility functions and helpers
- Target: >80% code coverage

### Integration Tests

- Service-to-service communication
- Database operations
- Message queue interactions
- API endpoint testing

### Load Testing

- Calculator service performance
- Wallet service transaction throughput
- Database query optimization
- API Gateway rate limiting

## Future Enhancements

### Phase 2 Features

- Machine learning for emission factor optimization
- Real-time IoT device integration
- Mobile app with offline capabilities
- Advanced analytics and insights

### Scalability Improvements

- Database sharding for high-volume services
- Event sourcing for audit trails
- CQRS pattern for read/write separation
- Microservice mesh (Istio) for advanced traffic management
