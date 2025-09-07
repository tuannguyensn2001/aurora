-- Create enum for condition operators
CREATE TYPE condition_operator AS ENUM (
    'equals',
    'not_equals',
    'contains',
    'not_contains',
    'greater_than',
    'less_than',
    'greater_than_or_equal',
    'less_than_or_equal',
    'in',
    'not_in'
);

-- Create segments table
CREATE TABLE segments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create segment_rules table
CREATE TABLE segment_rules (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    segment_id INTEGER NOT NULL
);

-- Create segment_rule_conditions table
CREATE TABLE segment_rule_conditions (
    id SERIAL PRIMARY KEY,
    attribute_id INTEGER NOT NULL,
    operator condition_operator NOT NULL,
    value TEXT NOT NULL,
    rule_id INTEGER NOT NULL
);

-- Create indexes for better performance
CREATE INDEX idx_segments_name ON segments(name);
CREATE INDEX idx_segment_rules_segment_id ON segment_rules(segment_id);
CREATE INDEX idx_segment_rule_conditions_rule_id ON segment_rule_conditions(rule_id);
CREATE INDEX idx_segment_rule_conditions_attribute_id ON segment_rule_conditions(attribute_id);

-- Create trigger to automatically update updated_at for segments
CREATE TRIGGER update_segments_updated_at 
    BEFORE UPDATE ON segments 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column(); 