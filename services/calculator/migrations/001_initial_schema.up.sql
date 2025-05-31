-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create calculations table
CREATE TABLE calculations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    total_co2_kg DECIMAL(10,3) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index on user_id for faster queries
CREATE INDEX idx_calculations_user_id ON calculations(user_id);
CREATE INDEX idx_calculations_created_at ON calculations(created_at);

-- Create activities table
CREATE TABLE activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    calculation_id UUID NOT NULL REFERENCES calculations(id) ON DELETE CASCADE,
    activity_type VARCHAR(100) NOT NULL,
    co2_kg DECIMAL(10,3) NOT NULL,
    emission_factor DECIMAL(10,6) NOT NULL,
    factor_source VARCHAR(255) NOT NULL,
    activity_data JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index on calculation_id for faster queries
CREATE INDEX idx_activities_calculation_id ON activities(calculation_id);
CREATE INDEX idx_activities_activity_type ON activities(activity_type);

-- Create emission_factors table
CREATE TABLE emission_factors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_type VARCHAR(100) NOT NULL,
    sub_type VARCHAR(100) NOT NULL,
    factor_co2_per_unit DECIMAL(10,6) NOT NULL,
    unit VARCHAR(50) NOT NULL,
    source VARCHAR(255) NOT NULL,
    location VARCHAR(100),
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for emission factors
CREATE INDEX idx_emission_factors_activity_type ON emission_factors(activity_type);
CREATE INDEX idx_emission_factors_sub_type ON emission_factors(sub_type);
CREATE INDEX idx_emission_factors_location ON emission_factors(location);
CREATE UNIQUE INDEX idx_emission_factors_unique ON emission_factors(activity_type, sub_type, location) WHERE location IS NOT NULL;
CREATE UNIQUE INDEX idx_emission_factors_unique_null_location ON emission_factors(activity_type, sub_type) WHERE location IS NULL;

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_calculations_updated_at BEFORE UPDATE ON calculations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_emission_factors_updated_at BEFORE UPDATE ON emission_factors
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
