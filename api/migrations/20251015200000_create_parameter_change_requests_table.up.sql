-- Create ENUM for change request status
CREATE TYPE parameter_change_request_status AS ENUM ('pending', 'approved', 'rejected', 'cancelled');

-- Create parameter_change_requests table
CREATE TABLE parameter_change_requests (
    id SERIAL PRIMARY KEY,
    parameter_id INTEGER NOT NULL REFERENCES parameters(id) ON DELETE CASCADE,
    requested_by_user_id INTEGER NOT NULL REFERENCES users(id),
    status parameter_change_request_status NOT NULL DEFAULT 'pending',
    description TEXT,
    
    -- Store the proposed changes as JSONB
    -- This will contain the full UpdateParameterWithRulesRequest data
    change_data JSONB NOT NULL,
    
    -- Approval/rejection metadata
    reviewed_by_user_id INTEGER REFERENCES users(id),
    reviewed_at TIMESTAMP,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_parameter_change_requests_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_parameter_change_requests_updated_at
    BEFORE UPDATE ON parameter_change_requests
    FOR EACH ROW
    EXECUTE FUNCTION update_parameter_change_requests_updated_at();

