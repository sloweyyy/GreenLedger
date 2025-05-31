# Changelog

All notable changes to GreenLedger will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Certificate marketplace functionality
- Advanced analytics dashboard
- Mobile app API endpoints
- Blockchain integration for certificate verification

### Changed
- Improved performance for large-scale calculations
- Enhanced security for API endpoints

### Deprecated
- Legacy API v1 endpoints (will be removed in v2.0.0)

### Removed
- None

### Fixed
- None

### Security
- Updated dependencies to latest versions

## [1.0.0] - 2025-05-31

### Added
- **Calculator Service**: Complete carbon footprint calculation engine
  - Vehicle emissions calculation (cars, motorcycles, trucks)
  - Electricity consumption calculations
  - Flight emissions with distance calculations
  - Heating/cooling energy calculations
  - Configurable emission factors database
  - REST API with Swagger documentation
  - gRPC service for inter-service communication

- **Activity Tracker Service**: Eco-activity tracking and credit system
  - Multiple eco-activity types (biking, walking, recycling, etc.)
  - Credit calculation with configurable rules
  - Activity verification system
  - IoT device integration endpoints
  - Webhook support for external data sources
  - Challenge and leaderboard systems
  - Event publishing to Kafka

- **Wallet Service**: Carbon credit management system
  - Credit balance management with atomic transactions
  - Credit transfers between users
  - Transaction history and analytics
  - Credit reservation system for pending operations
  - Wallet snapshots for reporting
  - Admin operations for credit management
  - Event-driven credit earning from activities

- **User Authentication Service**: Complete auth system
  - JWT-based authentication with RBAC
  - User registration, login, and password management
  - Role-based access control (Admin, User, Moderator)
  - Email verification and password reset
  - Session management
  - Security middleware and rate limiting

- **Reporting Service**: Comprehensive reporting system
  - PDF, JSON, CSV report generation
  - Carbon footprint analysis reports
  - Credit balance and transaction reports
  - Summary reports with statistics and charts
  - Scheduled report generation
  - Report templates and customization
  - Cross-service data aggregation

- **Certificate Service**: Digital certificate management
  - Carbon offset certificate models
  - Certificate verification system
  - Blockchain integration structure
  - Certificate transfer and ownership system
  - Project management for certificates
  - NFT-based certificate support

- **Infrastructure & DevOps**:
  - Docker containerization for all services
  - Docker Compose orchestration
  - Traefik API gateway with load balancing
  - PostgreSQL databases (one per service)
  - Redis caching layer
  - Kafka event streaming
  - Prometheus monitoring and metrics
  - Grafana dashboards
  - Comprehensive Makefile for development

- **Testing & Quality**:
  - Unit tests for all services (>80% coverage)
  - Integration test framework
  - Load testing capabilities
  - Automated linting and code quality checks
  - Security scanning in CI/CD

- **Documentation**:
  - Comprehensive README with quick start guide
  - API documentation with Swagger/OpenAPI
  - Architecture documentation
  - Contributing guidelines
  - Security policy
  - Code of conduct

### Security
- JWT authentication with configurable expiration
- Input validation and sanitization on all endpoints
- SQL injection prevention through GORM ORM
- CORS configuration for web clients
- Secure headers middleware
- Rate limiting structure
- Non-root container users
- Security scanning in CI/CD pipeline

## [0.9.0] - 2025-05-30

### Added
- Initial project structure
- Basic microservices architecture
- Database schema design
- Core API endpoints for calculator service

### Changed
- Migrated from monolithic to microservices architecture

## [0.8.0] - 2025-05-15

### Added
- Basic carbon footprint calculation
- User management system
- Simple reporting functionality

### Fixed
- Database connection issues
- Authentication token expiration handling

## [0.7.0] - 2025-05-01

### Added
- Initial MVP release
- Basic web interface
- User registration and login
- Simple carbon tracking

---

## Release Notes Format

Each release includes:

### Added
- New features and functionality

### Changed
- Changes in existing functionality

### Deprecated
- Soon-to-be removed features

### Removed
- Features removed in this release

### Fixed
- Bug fixes

### Security
- Security improvements and vulnerability fixes

## Version Numbering

We use [Semantic Versioning](https://semver.org/):

- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality additions
- **PATCH** version for backwards-compatible bug fixes

## Release Schedule

- **Major releases**: Every 6-12 months
- **Minor releases**: Every 1-2 months
- **Patch releases**: As needed for critical fixes

## Migration Guides

For breaking changes, detailed migration guides are provided:

### Upgrading from v0.x to v1.0

1. **Database Migration**: Run the provided migration scripts
2. **API Changes**: Update client code to use new API endpoints
3. **Configuration**: Update environment variables and config files
4. **Dependencies**: Update Docker images and dependencies

### API Versioning

- **v1**: Current stable API
- **v2**: Next major version (in development)
- **Legacy**: v1 will be supported for 12 months after v2 release

## Support

- **Current version**: Full support with new features and bug fixes
- **Previous major version**: Security fixes and critical bug fixes only
- **Older versions**: No longer supported

For support questions, please:
- Check the [FAQ](docs/FAQ.md)
- Search [existing issues](https://github.com/sloweyyy/GreenLedger/issues)
- Create a [new issue](https://github.com/sloweyyy/GreenLedger/issues/new)
- Contact [truonglevinhphuc2006@gmail.com](mailto:truonglevinhphuc2006@gmail.com)

---

**Note**: This changelog is automatically updated as part of our release process. For the most up-to-date information, please check our [GitHub releases](https://github.com/sloweyyy/GreenLedger/releases).
