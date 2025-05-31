-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create activity_types table
CREATE TABLE activity_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    category VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    base_credits_per_unit DECIMAL(10,3) NOT NULL,
    unit VARCHAR(50) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    requires_verification BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create eco_activities table
CREATE TABLE eco_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    activity_type_id UUID NOT NULL REFERENCES activity_types(id) ON DELETE RESTRICT,
    description TEXT NOT NULL,
    duration INTEGER DEFAULT 0,
    distance DECIMAL(10,3) DEFAULT 0,
    quantity DECIMAL(10,3) DEFAULT 0,
    unit VARCHAR(50),
    location VARCHAR(255),
    credits_earned DECIMAL(10,3) NOT NULL DEFAULT 0,
    is_verified BOOLEAN DEFAULT false,
    verified_at TIMESTAMP WITH TIME ZONE,
    verified_by VARCHAR(255),
    source VARCHAR(50) NOT NULL DEFAULT 'manual',
    source_data JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create credit_rules table
CREATE TABLE credit_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_type_id UUID NOT NULL REFERENCES activity_types(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    min_value DECIMAL(10,3) DEFAULT 0,
    max_value DECIMAL(10,3) DEFAULT 0,
    credits_per_unit DECIMAL(10,6) NOT NULL,
    multiplier DECIMAL(5,2) DEFAULT 1.0,
    is_active BOOLEAN DEFAULT true,
    valid_from TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    valid_to TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create activity_challenges table
CREATE TABLE activity_challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE NOT NULL,
    target_value DECIMAL(10,3) NOT NULL,
    target_unit VARCHAR(50) NOT NULL,
    reward_credits DECIMAL(10,3) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create challenge_participants table
CREATE TABLE challenge_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    challenge_id UUID NOT NULL REFERENCES activity_challenges(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    progress DECIMAL(10,3) DEFAULT 0,
    is_completed BOOLEAN DEFAULT false,
    completed_at TIMESTAMP WITH TIME ZONE,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(challenge_id, user_id)
);

-- Create iot_devices table
CREATE TABLE iot_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    device_id VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    last_seen TIMESTAMP WITH TIME ZONE,
    api_key VARCHAR(255) UNIQUE NOT NULL,
    settings JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX idx_eco_activities_user_id ON eco_activities(user_id);
CREATE INDEX idx_eco_activities_activity_type_id ON eco_activities(activity_type_id);
CREATE INDEX idx_eco_activities_created_at ON eco_activities(created_at);
CREATE INDEX idx_eco_activities_is_verified ON eco_activities(is_verified);
CREATE INDEX idx_eco_activities_source ON eco_activities(source);

CREATE INDEX idx_activity_types_category ON activity_types(category);
CREATE INDEX idx_activity_types_is_active ON activity_types(is_active);

CREATE INDEX idx_credit_rules_activity_type_id ON credit_rules(activity_type_id);
CREATE INDEX idx_credit_rules_is_active ON credit_rules(is_active);
CREATE INDEX idx_credit_rules_valid_from ON credit_rules(valid_from);
CREATE INDEX idx_credit_rules_valid_to ON credit_rules(valid_to);

CREATE INDEX idx_activity_challenges_start_date ON activity_challenges(start_date);
CREATE INDEX idx_activity_challenges_end_date ON activity_challenges(end_date);
CREATE INDEX idx_activity_challenges_is_active ON activity_challenges(is_active);

CREATE INDEX idx_challenge_participants_challenge_id ON challenge_participants(challenge_id);
CREATE INDEX idx_challenge_participants_user_id ON challenge_participants(user_id);
CREATE INDEX idx_challenge_participants_is_completed ON challenge_participants(is_completed);

CREATE INDEX idx_iot_devices_user_id ON iot_devices(user_id);
CREATE INDEX idx_iot_devices_device_id ON iot_devices(device_id);
CREATE INDEX idx_iot_devices_api_key ON iot_devices(api_key);
CREATE INDEX idx_iot_devices_is_active ON iot_devices(is_active);

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_activity_types_updated_at BEFORE UPDATE ON activity_types
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_eco_activities_updated_at BEFORE UPDATE ON eco_activities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_credit_rules_updated_at BEFORE UPDATE ON credit_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_activity_challenges_updated_at BEFORE UPDATE ON activity_challenges
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_iot_devices_updated_at BEFORE UPDATE ON iot_devices
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
