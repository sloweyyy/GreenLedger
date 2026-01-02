#!/bin/bash
# Run all services concurrently

echo "🌱 Starting all GreenLedger services..."

# Function to run service in background
run_service() {
    local service=$1
    local port=$2
    echo "Starting $service on port $port..."
    ./scripts/run/$service.sh &
}

# Start all services
run_service calculator 8081
run_service tracker 8082
run_service wallet 8083
run_service user-auth 8084

echo "All services started!"
echo "Press Ctrl+C to stop all services"

# Wait for all background processes
wait
