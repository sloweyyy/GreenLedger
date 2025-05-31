-- Drop triggers
DROP TRIGGER IF EXISTS update_calculations_updated_at ON calculations;
DROP TRIGGER IF EXISTS update_emission_factors_updated_at ON emission_factors;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables (in reverse order due to foreign key constraints)
DROP TABLE IF EXISTS activities;
DROP TABLE IF EXISTS emission_factors;
DROP TABLE IF EXISTS calculations;

-- Drop extension (optional, might be used by other services)
-- DROP EXTENSION IF EXISTS "uuid-ossp";
