#!/bin/bash
# Run wallet service

cd services/wallet
export DB_TYPE=sqlite
export DB_PATH="../../data/wallet.db"
export LOG_LEVEL=debug
export ENVIRONMENT=development

echo "🚀 Starting wallet service..."
go run ./cmd/main.go
