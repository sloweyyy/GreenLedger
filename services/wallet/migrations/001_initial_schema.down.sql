-- Drop triggers
DROP TRIGGER IF EXISTS update_wallets_updated_at ON wallets;
DROP TRIGGER IF EXISTS update_transactions_updated_at ON transactions;
DROP TRIGGER IF EXISTS update_transaction_batches_updated_at ON transaction_batches;
DROP TRIGGER IF EXISTS update_credit_reservations_updated_at ON credit_reservations;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables (in reverse order due to foreign key constraints)
DROP TABLE IF EXISTS wallet_snapshots;
DROP TABLE IF EXISTS credit_reservations;
DROP TABLE IF EXISTS transaction_batches;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS wallets;

-- Drop extension (optional, might be used by other services)
-- DROP EXTENSION IF EXISTS "uuid-ossp";
