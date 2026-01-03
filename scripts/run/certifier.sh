#!/bin/bash
# Run certifier service

cd services/certifier
export DB_TYPE=sqlite
export DB_PATH="../../data/certifier.db"
export LOG_LEVEL=debug
export ENVIRONMENT=development

echo "🚀 Starting certifier service..."
go run ./cmd/main.go
