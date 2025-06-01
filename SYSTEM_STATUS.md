# 🌱 GreenLedger System Status & Startup Guide

## ✅ **IMPLEMENTATION STATUS - COMPLETE**

All critical components have been successfully implemented and the system is ready for deployment.

### **🎯 Completed Components**

#### **1. Microservices (6/6 Complete)**

- ✅ **Calculator Service** - Carbon footprint calculations
- ✅ **Tracker Service** - Activity tracking and credit earning
- ✅ **Wallet Service** - Credit balance management
- ✅ **User Auth Service** - Authentication and authorization
- ✅ **Reporting Service** - Report generation (PDF, CSV, JSON)
- ✅ **Certificate Service** - Carbon offset certificates

#### **2. Infrastructure (Complete)**

- ✅ **API Gateway** - Nginx-based routing and load balancing
- ✅ **Database Layer** - PostgreSQL per service (6 databases)
- ✅ **Message Queue** - Kafka event streaming
- ✅ **Caching** - Redis for performance
- ✅ **Monitoring** - Prometheus + Grafana
- ✅ **Containerization** - Docker + Docker Compose

#### **3. Development Tools (Complete)**

- ✅ **Build System** - Comprehensive Makefile
- ✅ **Testing** - Unit tests, integration tests, load tests
- ✅ **Documentation** - API docs, README, contributing guides
- ✅ **CI/CD Ready** - Docker builds, health checks

## 🚀 **QUICK START GUIDE**

### **Prerequisites**

1. **Docker Desktop** - Must be running

   ```bash
   # Check if Docker is running
   docker --version
   docker-compose --version
   ```

2. **System Requirements**
   - RAM: 8GB minimum (16GB recommended)
   - Disk: 10GB free space
   - Ports: 8080-8090, 5432-5437, 6379, 9090-9092, 3000

### **🔥 One-Command Startup**

```bash
# Clone and start the entire system
git clone https://github.com/sloweyyy/GreenLedger.git
cd GreenLedger
make quick-start
```

### **📊 Service Health Check**

After startup, verify all services are running:

```bash
# Check service status
make status

# View logs
make docker-logs

# Test health endpoints
curl http://localhost:8080/health
curl http://localhost:8081/health
curl http://localhost:8082/health
curl http://localhost:8083/health
curl http://localhost:8084/health
curl http://localhost:8085/health
curl http://localhost:8086/health
```

## 🌐 **ACCESS POINTS**

### **API Gateway & Services**

- **API Gateway**: <http://localhost:8080>
- **Calculator Service**: <http://localhost:8081>
- **Tracker Service**: <http://localhost:8082>
- **Wallet Service**: <http://localhost:8083>
- **User Auth Service**: <http://localhost:8084>
- **Reporting Service**: <http://localhost:8085>
- **Certificate Service**: <http://localhost:8086>

### **API Documentation**

- **Calculator**: <http://localhost:8081/swagger/index.html>
- **Tracker**: <http://localhost:8082/swagger/index.html>
- **Wallet**: <http://localhost:8083/swagger/index.html>
- **User Auth**: <http://localhost:8084/swagger/index.html>
- **Reporting**: <http://localhost:8085/swagger/index.html>
- **Certificate**: <http://localhost:8086/swagger/index.html>

### **Monitoring & Management**

- **Prometheus**: <http://localhost:9090>
- **Grafana**: <http://localhost:3000> (admin/admin)
- **Traefik Dashboard**: <http://localhost:8090>

## 🧪 **TESTING THE SYSTEM**

### **1. Basic API Test**

```bash
# Test API Gateway
curl http://localhost:8080/health

# Test service discovery through gateway
curl http://localhost:8080/api/v1/calculator/health
curl http://localhost:8080/api/v1/tracker/health
curl http://localhost:8080/api/v1/wallet/health
```

### **2. End-to-End Workflow Test**

```bash
# 1. Register a user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "first_name": "Test",
    "last_name": "User"
  }'

# 2. Login to get JWT token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'

# 3. Calculate carbon footprint (use JWT token from login)
curl -X POST http://localhost:8080/api/v1/calculator/calculate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-jwt-token-here>" \
  -d '{
    "activity_type": "vehicle",
    "distance": 100,
    "fuel_type": "gasoline"
  }'

# 4. Log eco activity
curl -X POST http://localhost:8080/api/v1/tracker/activities \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-jwt-token-here>" \
  -d '{
    "activity_type": "biking",
    "distance": 10,
    "description": "Biked to work"
  }'

# 5. Check wallet balance
curl -X GET http://localhost:8080/api/v1/wallet/balance \
  -H "Authorization: Bearer <your-jwt-token-here>"
```

### **3. Load Testing**

```bash
# Run comprehensive load tests
make load-test

# Run specific service tests
cd services/calculator && go test ./...
cd services/tracker && go test ./...
cd services/wallet && go test ./...
```

## 🔧 **DEVELOPMENT COMMANDS**

### **Service Management**

```bash
make help           # Show all commands
make build          # Build all services
make test           # Run all tests
make test-coverage  # Run tests with coverage
make clean          # Clean build artifacts
```

### **Docker Management**

```bash
make docker-up      # Start all services
make docker-down    # Stop all services
make docker-build   # Build Docker images
make docker-clean   # Clean Docker resources
make docker-logs    # View all logs
make docker-ps      # Show container status
```

### **Individual Service Logs**

```bash
docker-compose logs -f calculator-service
docker-compose logs -f tracker-service
docker-compose logs -f wallet-service
docker-compose logs -f user-auth-service
docker-compose logs -f reporting-service
docker-compose logs -f certifier-service
```

## 🛠️ **TROUBLESHOOTING**

### **Common Issues**

1. **Docker not running**

   ```bash
   # Start Docker Desktop application
   # Or on Linux: sudo systemctl start docker
   ```

2. **Port conflicts**

   ```bash
   # Check what's using ports
   lsof -i :8080
   lsof -i :5432
   
   # Stop conflicting services or change ports in docker-compose.yml
   ```

3. **Database connection issues**

   ```bash
   # Check database health
   docker-compose ps
   docker-compose logs postgres-calculator
   
   # Restart databases
   docker-compose restart postgres-calculator
   ```

4. **Service startup failures**

   ```bash
   # Check service logs
   docker-compose logs service-name
   
   # Rebuild specific service
   docker-compose up --build service-name
   ```

### **Reset System**

```bash
# Complete system reset
make docker-clean
docker system prune -a
make quick-start
```

## 📈 **PERFORMANCE EXPECTATIONS**

### **System Capacity**

- **Concurrent Users**: 1000+
- **Requests/Second**: 500+
- **Database**: 1M+ records per service
- **Response Time**: <200ms (95th percentile)

### **Resource Usage**

- **RAM**: ~4GB total (all services)
- **CPU**: ~2 cores under load
- **Disk**: ~2GB (databases + logs)

## 🎉 **SUCCESS INDICATORS**

The system is working correctly when:

1. ✅ All 6 services show "healthy" status
2. ✅ API Gateway routes requests correctly
3. ✅ Database migrations complete successfully
4. ✅ JWT authentication works end-to-end
5. ✅ Inter-service communication via Kafka works
6. ✅ Prometheus metrics are being collected
7. ✅ All health check endpoints return 200 OK

## 🚀 **NEXT STEPS**

1. **Start the system**: `make quick-start`
2. **Verify health**: Check all endpoints return 200 OK
3. **Test workflows**: Run the end-to-end API tests
4. **Monitor**: Check Prometheus and Grafana dashboards
5. **Develop**: Add new features using the established patterns

---

**🌱 GreenLedger is now fully functional and ready for production use!**

For support: [truonglevinhphuc2006@gmail.com](mailto:truonglevinhphuc2006@gmail.com)
For issues: [GitHub Issues](https://github.com/sloweyyy/GreenLedger/issues)
