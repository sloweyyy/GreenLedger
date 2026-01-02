#!/bin/bash
# Run reporting service

cd services/reporting
export DB_TYPE=sqlite
export DB_PATH="../../data/reporting.db"
export LOG_LEVEL=debug
export ENVIRONMENT=development

echo "🚀 Starting reporting service..."
go run ./cmd/main.go
