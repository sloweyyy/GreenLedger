#!/bin/bash

# GreenLedger System Test Script
echo "🧪 GreenLedger System Test"
echo "========================="

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
API_GATEWAY="http://localhost:8080"

echo ""
echo "🔍 Pre-flight Checks"
echo "-------------------"

# Check if Docker is running
if ! docker info >/dev/null 2>&1; then
    echo -e "${RED}✗${NC} Docker is not running. Please start Docker Desktop."
    exit 1
else
    echo -e "${GREEN}✓${NC} Docker is running"
fi

echo ""
echo "🌐 Testing API Gateway"
echo "---------------------"

if curl -s --max-time 10 "$API_GATEWAY/health" >/dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} API Gateway is accessible"
else
    echo -e "${RED}✗${NC} API Gateway is not accessible"
    echo "Run 'make quick-start' to start the system"
    exit 1
fi

echo ""
echo -e "${GREEN}🎉 Basic system test complete!${NC}"
echo ""
echo "🌐 Access Points:"
echo "  API Gateway:    $API_GATEWAY"
echo "  Prometheus:     http://localhost:9090"
echo "  Grafana:        http://localhost:3000"
echo ""
echo -e "${BLUE}System is ready! 🚀${NC}"
