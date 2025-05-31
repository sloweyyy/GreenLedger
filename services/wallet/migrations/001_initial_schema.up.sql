-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create wallets table
CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) UNIQUE NOT NULL,
    available_credits DECIMAL(15,3) NOT NULL DEFAULT 0,
    pending_credits DECIMAL(15,3) NOT NULL DEFAULT 0,
    total_earned DECIMAL(15,3) NOT NULL DEFAULT 0,
    total_spent DECIMAL(15,3) NOT NULL DEFAULT 0,
    last_updated TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create transactions table
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    amount DECIMAL(15,3) NOT NULL,
    balance_after DECIMAL(15,3) NOT NULL,
    source VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    reference_id VARCHAR(255),
    from_user_id VARCHAR(255),
    to_user_id VARCHAR(255),
    metadata JSONB,
    processed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create transaction_batches table
CREATE TABLE transaction_batches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_id VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    total_amount DECIMAL(15,3) NOT NULL,
    description TEXT,
    processed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create credit_reservations table
CREATE TABLE credit_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    amount DECIMAL(15,3) NOT NULL,
    purpose VARCHAR(255) NOT NULL,
    reference_id VARCHAR(255),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_released BOOLEAN DEFAULT false,
    released_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create wallet_snapshots table
CREATE TABLE wallet_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    available_credits DECIMAL(15,3) NOT NULL,
    pending_credits DECIMAL(15,3) NOT NULL,
    total_earned DECIMAL(15,3) NOT NULL,
    total_spent DECIMAL(15,3) NOT NULL,
    snapshot_date TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX idx_wallets_user_id ON wallets(user_id);
CREATE INDEX idx_wallets_available_credits ON wallets(available_credits);
CREATE INDEX idx_wallets_last_updated ON wallets(last_updated);

CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);
CREATE INDEX idx_transactions_reference_id ON transactions(reference_id);
CREATE INDEX idx_transactions_from_user_id ON transactions(from_user_id);
CREATE INDEX idx_transactions_to_user_id ON transactions(to_user_id);
CREATE INDEX idx_transactions_processed_at ON transactions(processed_at);

CREATE INDEX idx_transaction_batches_batch_id ON transaction_batches(batch_id);
CREATE INDEX idx_transaction_batches_status ON transaction_batches(status);
CREATE INDEX idx_transaction_batches_created_at ON transaction_batches(created_at);

CREATE INDEX idx_credit_reservations_user_id ON credit_reservations(user_id);
CREATE INDEX idx_credit_reservations_reference_id ON credit_reservations(reference_id);
CREATE INDEX idx_credit_reservations_expires_at ON credit_reservations(expires_at);
CREATE INDEX idx_credit_reservations_is_released ON credit_reservations(is_released);

CREATE INDEX idx_wallet_snapshots_user_id ON wallet_snapshots(user_id);
CREATE INDEX idx_wallet_snapshots_snapshot_date ON wallet_snapshots(snapshot_date);

-- Add constraints
ALTER TABLE wallets ADD CONSTRAINT chk_wallets_available_credits_non_negative 
    CHECK (available_credits >= 0);
ALTER TABLE wallets ADD CONSTRAINT chk_wallets_pending_credits_non_negative 
    CHECK (pending_credits >= 0);
ALTER TABLE wallets ADD CONSTRAINT chk_wallets_total_earned_non_negative 
    CHECK (total_earned >= 0);
ALTER TABLE wallets ADD CONSTRAINT chk_wallets_total_spent_non_negative 
    CHECK (total_spent >= 0);

ALTER TABLE transactions ADD CONSTRAINT chk_transactions_amount_positive 
    CHECK (amount > 0);

ALTER TABLE credit_reservations ADD CONSTRAINT chk_credit_reservations_amount_positive 
    CHECK (amount > 0);

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_wallets_updated_at BEFORE UPDATE ON wallets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_transactions_updated_at BEFORE UPDATE ON transactions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_transaction_batches_updated_at BEFORE UPDATE ON transaction_batches
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_credit_reservations_updated_at BEFORE UPDATE ON credit_reservations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
