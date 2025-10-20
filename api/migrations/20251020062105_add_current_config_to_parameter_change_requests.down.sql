-- Remove current_config column from parameter_change_requests table
ALTER TABLE parameter_change_requests 
DROP COLUMN current_config;
