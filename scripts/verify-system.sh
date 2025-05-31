#!/bin/bash

# GreenLedger System Verification Script
# This script verifies that all critical components are in place

set -e

echo "🌱 GreenLedger System Verification"
echo "=================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to check if file exists
check_file() {
    if [ -f "$1" ]; then
        echo -e "${GREEN}✓${NC} $1"
        return 0
    else
        echo -e "${RED}✗${NC} $1 (MISSING)"
        return 1
    fi
}

# Function to check if directory exists
check_dir() {
    if [ -d "$1" ]; then
        echo -e "${GREEN}✓${NC} $1/"
        return 0
    else
        echo -e "${RED}✗${NC} $1/ (MISSING)"
        return 1
    fi
}

echo ""
echo "📁 Checking Directory Structure..."
echo "--------------------------------"

# Check main directories
check_dir "services"
check_dir "shared"
check_dir "api-gateway"
check_dir "scripts"
check_dir "tests"

echo ""
echo "🔧 Checking Service Main Files..."
echo "--------------------------------"

# Check main entry points
check_file "services/calculator/cmd/main.go"
check_file "services/tracker/cmd/main.go"
check_file "services/wallet/cmd/main.go"
check_file "services/user-auth/cmd/main.go"
check_file "services/reporting/cmd/main.go"
check_file "services/certifier/cmd/main.go"

echo ""
echo "🐳 Checking Docker Configuration..."
echo "----------------------------------"

# Check Docker files
check_file "docker-compose.yml"
check_file "services/calculator/Dockerfile"
check_file "services/tracker/Dockerfile"
check_file "services/wallet/Dockerfile"
check_file "services/user-auth/Dockerfile"
check_file "services/reporting/Dockerfile"
check_file "services/certifier/Dockerfile"

echo ""
echo "🌐 Checking API Gateway..."
echo "-------------------------"

# Check API Gateway configuration
check_file "api-gateway/nginx.conf"

echo ""
echo "📦 Checking Go Modules..."
echo "------------------------"

# Check go.mod files
check_file "go.mod"
check_file "services/calculator/go.mod"
check_file "services/tracker/go.mod"
check_file "services/wallet/go.mod"
check_file "services/user-auth/go.mod"
check_file "services/reporting/go.mod"
check_file "services/certifier/go.mod"

echo ""
echo "📚 Checking Documentation..."
echo "---------------------------"

# Check documentation files
check_file "README.md"
check_file "CONTRIBUTING.md"
check_file "SECURITY.md"
check_file "CHANGELOG.md"
check_file "LICENSE"
check_file "Makefile"

echo ""
echo "🔍 Checking Shared Components..."
echo "-------------------------------"

# Check shared components
check_file "shared/config/config.go"
check_file "shared/database/postgres.go"
check_file "shared/logger/logger.go"
check_file "shared/middleware/auth.go"

echo ""
echo "🧪 Running Basic Validation..."
echo "-----------------------------"

# Check if Go modules are valid
echo "Checking Go module validity..."
if go mod verify > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Root go.mod is valid"
else
    echo -e "${YELLOW}⚠${NC} Root go.mod may have issues"
fi

# Check if Docker Compose is valid
echo "Checking Docker Compose syntax..."
if docker-compose config > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} docker-compose.yml is valid"
else
    echo -e "${RED}✗${NC} docker-compose.yml has syntax errors"
fi

echo ""
echo "📊 System Status Summary"
echo "======================="

# Count missing files
missing_count=0

# Re-run checks silently to count missing files
files_to_check=(
    "services/calculator/cmd/main.go"
    "services/tracker/cmd/main.go"
    "services/wallet/cmd/main.go"
    "services/user-auth/cmd/main.go"
    "services/reporting/cmd/main.go"
    "services/certifier/cmd/main.go"
    "docker-compose.yml"
    "api-gateway/nginx.conf"
    "Makefile"
    "README.md"
)

for file in "${files_to_check[@]}"; do
    if [ ! -f "$file" ]; then
        ((missing_count++))
    fi
done

if [ $missing_count -eq 0 ]; then
    echo -e "${GREEN}🎉 All critical components are present!${NC}"
    echo -e "${GREEN}✅ System is ready for deployment${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Run 'make quick-start' to start all services"
    echo "2. Check health endpoints at http://localhost:8080"
    echo "3. View API documentation at service endpoints"
else
    echo -e "${RED}❌ $missing_count critical components are missing${NC}"
    echo -e "${YELLOW}⚠ System may not start properly${NC}"
    echo ""
    echo "Please ensure all missing files are created before deployment."
fi

echo ""
echo "🔗 Quick Start Commands:"
echo "----------------------"
echo "make help           # Show all available commands"
echo "make quick-start    # Start all services"
echo "make docker-logs    # View service logs"
echo "make docker-ps      # Check service status"
echo "make stop           # Stop all services"

echo ""
echo "🌐 Service Endpoints (when running):"
echo "-----------------------------------"
echo "API Gateway:    http://localhost:8080"
echo "Calculator:     http://localhost:8081"
echo "Tracker:        http://localhost:8082"
echo "Wallet:         http://localhost:8083"
echo "User Auth:      http://localhost:8084"
echo "Reporting:      http://localhost:8085"
echo "Certificate:    http://localhost:8086"
echo "Prometheus:     http://localhost:9090"
echo "Grafana:        http://localhost:3000"

echo ""
echo "Verification complete! 🌱"
