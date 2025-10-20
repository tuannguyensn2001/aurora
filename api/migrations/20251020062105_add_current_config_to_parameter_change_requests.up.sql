-- Add current_config column to parameter_change_requests table
-- This column will store the current parameter configuration at the time of change request
ALTER TABLE parameter_change_requests 
ADD COLUMN current_config JSONB NOT NULL DEFAULT '{}';

-- Update existing records to have a basic current_config structure
-- For existing records, we'll set a minimal structure since we don't have historical data
UPDATE parameter_change_requests 
SET current_config = '{"name": "", "description": "", "dataType": "string", "defaultRolloutValue": null, "rules": []}'::jsonb
WHERE current_config = '{}'::jsonb;

-- Add a comment to document the purpose of this column
COMMENT ON COLUMN parameter_change_requests.current_config IS 'Stores the current parameter configuration at the time of change request creation for comparison and audit purposes';
