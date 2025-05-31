# 🌱 GreenLedger Local Development Guide

This guide will help you set up and run GreenLedger services locally for development.

## 🚀 Quick Start

### Prerequisites

- **Go 1.19+** - [Download here](https://golang.org/dl/)
- **Git** - For version control
- **At least 1GB free disk space**

### One-Command Setup

```bash
./scripts/dev-setup.sh
```

This script will:
- ✅ Check your Go installation
- ✅ Create necessary directories
- ✅ Set up environment configuration
- ✅ Install development tools
- ✅ Fix module dependencies
- ✅ Create run scripts
- ✅ Run tests to verify setup

## 📁 Project Structure

```
GreenLedger/
├── .env                    # Your local environment config
├── .env.example           # Template for environment variables
├── .gitignore            # Git ignore rules
├── .golangci.yml         # Linting configuration
├── .air.toml             # Live reload configuration
├── data/                 # SQLite databases (local dev)
├── scripts/
│   ├── dev-setup.sh      # Setup script
│   └── run/              # Service run scripts
├── services/
│   ├── calculator/       # Carbon calculation service
│   ├── tracker/          # Activity tracking service
│   ├── wallet/           # Credit wallet service
│   └── user-auth/        # Authentication service
└── shared/               # Shared libraries
```

## 🏃‍♂️ Running Services

### Option 1: Run All Services

```bash
./scripts/run/all.sh
```

### Option 2: Run Individual Services

```bash
# Calculator Service (Port 8081)
./scripts/run/calculator.sh

# Tracker Service (Port 8082)
./scripts/run/tracker.sh

# Wallet Service (Port 8083)
./scripts/run/wallet.sh

# User Auth Service (Port 8084)
./scripts/run/user-auth.sh
```

### Option 3: Live Reload Development

For active development with automatic reloading:

```bash
cd services/calculator
air
```

## 🧪 Testing Services

### Health Checks

```bash
curl http://localhost:8081/health  # Calculator
curl http://localhost:8082/health  # Tracker
curl http://localhost:8083/health  # Wallet
curl http://localhost:8084/health  # User Auth
```

### API Endpoints

```bash
# Calculator Service
curl http://localhost:8081/api/v1/calculations

# Tracker Service
curl http://localhost:8082/api/v1/activities

# Wallet Service
curl http://localhost:8083/api/v1/wallets

# User Auth Service
curl -X POST http://localhost:8084/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```

## 🔧 Development Tools

### Code Linting

```bash
make lint
# or
golangci-lint run ./...
```

### Security Scanning

```bash
make security
# or
gosec ./...
```

### Running Tests

```bash
make test
# or
go test ./...
```

### Code Formatting

```bash
make format
# or
go fmt ./...
```

## 📊 Monitoring & Debugging

### Logs

Service logs are printed to stdout. For persistent logging:

```bash
./scripts/run/calculator.sh > logs/calculator.log 2>&1 &
```

### Database

Local development uses SQLite databases stored in `data/`:
- `data/calculator.db`
- `data/tracker.db`
- `data/wallet.db`
- `data/userauth.db`

To reset a database, simply delete the file:

```bash
rm data/calculator.db
```

## 🌐 API Documentation

Each service provides Swagger documentation:

- Calculator: http://localhost:8081/swagger/index.html
- Tracker: http://localhost:8082/swagger/index.html
- Wallet: http://localhost:8083/swagger/index.html
- User Auth: http://localhost:8084/swagger/index.html

## 🔧 Configuration

### Environment Variables

Edit `.env` file to customize:

```bash
# Service ports
CALCULATOR_PORT=8081
TRACKER_PORT=8082
WALLET_PORT=8083
USER_AUTH_PORT=8084

# Database
DB_TYPE=sqlite
DB_PATH=./data/

# Logging
LOG_LEVEL=debug
ENVIRONMENT=development

# Security
JWT_SECRET=your-secret-key
```

### Adding New Environment Variables

1. Add to `.env.example`
2. Update your local `.env`
3. Update service configuration loading

## 🐛 Troubleshooting

### Port Already in Use

```bash
# Find process using port
lsof -i :8081

# Kill process
kill -9 <PID>
```

### Module Issues

```bash
# Clean and rebuild modules
./scripts/fix-dependencies.sh

# Or manually
go clean -modcache
go mod download
```

### Database Issues

```bash
# Reset all databases
rm data/*.db

# Reset specific database
rm data/calculator.db
```

### Disk Space Issues

```bash
# Clean Go cache
go clean -cache

# Clean Docker (if installed)
docker system prune -a

# Clean temporary files
rm -rf tmp/ logs/
```

## 🚀 Next Steps

1. **Add Features**: Implement new endpoints in services
2. **Write Tests**: Add unit and integration tests
3. **Frontend**: Create a web interface
4. **Production**: Set up Docker and Kubernetes deployment

## 📚 Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [Gin Framework](https://gin-gonic.com/)
- [GORM Documentation](https://gorm.io/)
- [Project Architecture](docs/ARCHITECTURE.md)

## 🤝 Contributing

1. Create a feature branch
2. Make your changes
3. Run tests: `make test`
4. Run linting: `make lint`
5. Submit a pull request

---

Happy coding! 🌱 If you encounter any issues, please check the troubleshooting section or create an issue.
