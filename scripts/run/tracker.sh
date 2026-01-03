#!/bin/bash
# Run tracker service

cd services/tracker
export DB_TYPE=sqlite
export DB_PATH="../../data/tracker.db"
export LOG_LEVEL=debug
export ENVIRONMENT=development

echo "🚀 Starting tracker service..."
go run ./cmd/main.go
