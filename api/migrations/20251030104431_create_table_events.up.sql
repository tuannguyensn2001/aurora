-- Create evaluation_events table
CREATE TABLE evaluation_events (
    id SERIAL PRIMARY KEY,
    event_id VARCHAR(255) UNIQUE NOT NULL,
    service_name VARCHAR(255) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    parameter_name VARCHAR(255) NOT NULL,
    source VARCHAR(50) NOT NULL,
    user_attributes JSONB,
    rollout_value TEXT,
    error TEXT,
    timestamp TIMESTAMP NOT NULL,
    experiment_id INTEGER,
    experiment_uuid VARCHAR(255),
    variant_id INTEGER,
    variant_name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX idx_evaluation_events_service_name ON evaluation_events(service_name);
CREATE INDEX idx_evaluation_events_parameter_name ON evaluation_events(parameter_name);
CREATE INDEX idx_evaluation_events_event_type ON evaluation_events(event_type);
CREATE INDEX idx_evaluation_events_timestamp ON evaluation_events(timestamp);
CREATE INDEX idx_evaluation_events_experiment_id ON evaluation_events(experiment_id);
CREATE INDEX idx_evaluation_events_created_at ON evaluation_events(created_at);
