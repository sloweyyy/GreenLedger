#!/bin/bash
# Run user-auth service

cd services/user-auth
export DB_TYPE=sqlite
export DB_PATH="../../data/user-auth.db"
export LOG_LEVEL=debug
export ENVIRONMENT=development

echo "🚀 Starting user-auth service..."
go run ./cmd/main.go
