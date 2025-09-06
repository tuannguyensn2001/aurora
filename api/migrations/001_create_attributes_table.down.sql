-- Drop trigger and function
DROP TRIGGER IF EXISTS update_attributes_updated_at ON attributes;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_attributes_data_type;
DROP INDEX IF EXISTS idx_attributes_name;

-- Drop table
DROP TABLE IF EXISTS attributes;

-- Drop enum type
DROP TYPE IF EXISTS data_type; 