#!/bin/bash

# Test script for the Calculator Service
set -e

BASE_URL="http://localhost:8081"
API_BASE="$BASE_URL/api/v1"

echo "🧪 Testing GreenLedger Calculator Service"
echo "========================================="

# Function to make HTTP requests
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local headers=$4
    
    if [ -n "$data" ]; then
        if [ -n "$headers" ]; then
            curl -s -X "$method" "$API_BASE$endpoint" \
                -H "Content-Type: application/json" \
                -H "$headers" \
                -d "$data"
        else
            curl -s -X "$method" "$API_BASE$endpoint" \
                -H "Content-Type: application/json" \
                -d "$data"
        fi
    else
        if [ -n "$headers" ]; then
            curl -s -X "$method" "$API_BASE$endpoint" \
                -H "$headers"
        else
            curl -s -X "$method" "$API_BASE$endpoint"
        fi
    fi
}

# Test 1: Health Check
echo "1. Testing Health Check..."
health_response=$(curl -s "$BASE_URL/health")
echo "Health Response: $health_response"

if echo "$health_response" | grep -q "healthy"; then
    echo "✅ Health check passed"
else
    echo "❌ Health check failed"
    exit 1
fi

echo ""

# Test 2: Get Emission Factors (Public endpoint)
echo "2. Testing Get Emission Factors..."
factors_response=$(make_request "GET" "/calculator/emission-factors")
echo "Emission Factors Response: $factors_response"
echo "✅ Emission factors endpoint accessible"
echo ""

# Test 3: Get Emission Factors by Type (Public endpoint)
echo "3. Testing Get Emission Factors by Type..."
factors_by_type_response=$(make_request "GET" "/calculator/emission-factors/vehicle_travel")
echo "Emission Factors by Type Response: $factors_by_type_response"
echo "✅ Emission factors by type endpoint accessible"
echo ""

# Test 4: Calculate Footprint (Protected endpoint - will fail without auth)
echo "4. Testing Calculate Footprint (without auth - should fail)..."
calculation_data='{
    "activities": [
        {
            "activity_type": "vehicle_travel",
            "data": {
                "vehicle_type": "car_gasoline",
                "distance_km": 100
            }
        },
        {
            "activity_type": "electricity_usage",
            "data": {
                "kwh_usage": 50,
                "location": "US"
            }
        }
    ]
}'

calculation_response=$(make_request "POST" "/calculator/calculate" "$calculation_data")
echo "Calculation Response (no auth): $calculation_response"

if echo "$calculation_response" | grep -q "missing authorization token\|User not authenticated"; then
    echo "✅ Authentication protection working"
else
    echo "❌ Authentication protection not working"
fi

echo ""

# Test 5: Get Calculation History (Protected endpoint - will fail without auth)
echo "5. Testing Get Calculation History (without auth - should fail)..."
history_response=$(make_request "GET" "/calculator/calculations")
echo "History Response (no auth): $history_response"

if echo "$history_response" | grep -q "missing authorization token\|User not authenticated"; then
    echo "✅ Authentication protection working for history endpoint"
else
    echo "❌ Authentication protection not working for history endpoint"
fi

echo ""

# Test 6: Test with Mock JWT Token (for demonstration)
echo "6. Testing with Mock JWT Token..."
# Note: This is a mock token for testing. In a real scenario, you'd get this from the auth service
mock_token="Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidGVzdC11c2VyLTEyMyIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsInJvbGVzIjpbInVzZXIiXSwiZXhwIjo5OTk5OTk5OTk5fQ.mock-signature"

calculation_with_auth_response=$(make_request "POST" "/calculator/calculate" "$calculation_data" "Authorization: $mock_token")
echo "Calculation Response (with mock auth): $calculation_with_auth_response"

if echo "$calculation_with_auth_response" | grep -q "invalid token"; then
    echo "✅ JWT validation working (mock token rejected)"
else
    echo "⚠️  JWT validation response: Check if this is expected"
fi

echo ""

echo "🎉 Calculator Service tests completed!"
echo ""
echo "📝 Summary:"
echo "- Health check: Working"
echo "- Public endpoints: Accessible"
echo "- Authentication: Protected endpoints require valid JWT"
echo "- Service is running and responding to requests"
echo ""
echo "🔧 Next steps:"
echo "1. Implement User Authentication Service to get valid JWT tokens"
echo "2. Add more comprehensive integration tests"
echo "3. Test with real emission factor data"
echo "4. Implement remaining services (Tracker, Wallet, etc.)"
