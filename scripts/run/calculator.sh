#!/bin/bash
# Run calculator service

cd services/calculator
export DB_TYPE=sqlite
export DB_PATH="../../data/calculator.db"
export LOG_LEVEL=debug
export ENVIRONMENT=development

echo "🚀 Starting calculator service..."
go run ./cmd/main.go
