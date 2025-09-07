-- Create enum for parameter data types
CREATE TYPE parameter_data_type AS ENUM ('boolean', 'string', 'number');

-- Create enum for condition match types
CREATE TYPE condition_match_type AS ENUM ('match', 'not_match');

-- Create enum for rule types
CREATE TYPE rule_type AS ENUM ('segment', 'attribute');

-- Create parameters table
CREATE TABLE parameters (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT NOT NULL,
    data_type parameter_data_type NOT NULL DEFAULT 'string',
    default_rollout_value JSONB NOT NULL,
    usage_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create parameter_conditions table (legacy support)
CREATE TABLE parameter_conditions (
    id SERIAL PRIMARY KEY,
    parameter_id INTEGER NOT NULL,
    segment_id INTEGER NOT NULL,
    match_type condition_match_type NOT NULL,
    rollout_value JSONB NOT NULL
);

-- Create parameter_rules table
CREATE TABLE parameter_rules (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type rule_type NOT NULL,
    rollout_value JSONB NOT NULL,
    parameter_id INTEGER NOT NULL,
    segment_id INTEGER,
    match_type condition_match_type
);

-- Create parameter_rule_conditions table
CREATE TABLE parameter_rule_conditions (
    id SERIAL PRIMARY KEY,
    attribute_id INTEGER NOT NULL,
    operator condition_operator NOT NULL,
    value TEXT NOT NULL,
    rule_id INTEGER NOT NULL
);

-- Create indexes for better performance
CREATE INDEX idx_parameters_name ON parameters(name);
CREATE INDEX idx_parameters_data_type ON parameters(data_type);
CREATE INDEX idx_parameter_conditions_parameter_id ON parameter_conditions(parameter_id);
CREATE INDEX idx_parameter_conditions_segment_id ON parameter_conditions(segment_id);
CREATE INDEX idx_parameter_rules_parameter_id ON parameter_rules(parameter_id);
CREATE INDEX idx_parameter_rules_segment_id ON parameter_rules(segment_id);
CREATE INDEX idx_parameter_rules_type ON parameter_rules(type);
CREATE INDEX idx_parameter_rule_conditions_rule_id ON parameter_rule_conditions(rule_id);
CREATE INDEX idx_parameter_rule_conditions_attribute_id ON parameter_rule_conditions(attribute_id);

-- Create trigger to automatically update updated_at for parameters
CREATE TRIGGER update_parameters_updated_at 
    BEFORE UPDATE ON parameters 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column(); 