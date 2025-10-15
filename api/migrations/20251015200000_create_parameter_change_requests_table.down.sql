-- Drop trigger and function
DROP TRIGGER IF EXISTS trigger_update_parameter_change_requests_updated_at ON parameter_change_requests;
DROP FUNCTION IF EXISTS update_parameter_change_requests_updated_at();

-- Drop table
DROP TABLE IF EXISTS parameter_change_requests;

-- Drop ENUM
DROP TYPE IF EXISTS parameter_change_request_status;

