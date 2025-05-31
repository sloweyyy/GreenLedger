-- Drop triggers
DROP TRIGGER IF EXISTS update_activity_types_updated_at ON activity_types;
DROP TRIGGER IF EXISTS update_eco_activities_updated_at ON eco_activities;
DROP TRIGGER IF EXISTS update_credit_rules_updated_at ON credit_rules;
DROP TRIGGER IF EXISTS update_activity_challenges_updated_at ON activity_challenges;
DROP TRIGGER IF EXISTS update_iot_devices_updated_at ON iot_devices;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables (in reverse order due to foreign key constraints)
DROP TABLE IF EXISTS iot_devices;
DROP TABLE IF EXISTS challenge_participants;
DROP TABLE IF EXISTS activity_challenges;
DROP TABLE IF EXISTS credit_rules;
DROP TABLE IF EXISTS eco_activities;
DROP TABLE IF EXISTS activity_types;

-- Drop extension (optional, might be used by other services)
-- DROP EXTENSION IF EXISTS "uuid-ossp";
