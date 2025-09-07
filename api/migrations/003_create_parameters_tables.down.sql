-- Drop trigger
DROP TRIGGER IF EXISTS update_parameters_updated_at ON parameters;

-- Drop indexes
DROP INDEX IF EXISTS idx_parameter_rule_conditions_attribute_id;
DROP INDEX IF EXISTS idx_parameter_rule_conditions_rule_id;
DROP INDEX IF EXISTS idx_parameter_rules_type;
DROP INDEX IF EXISTS idx_parameter_rules_segment_id;
DROP INDEX IF EXISTS idx_parameter_rules_parameter_id;
DROP INDEX IF EXISTS idx_parameter_conditions_segment_id;
DROP INDEX IF EXISTS idx_parameter_conditions_parameter_id;
DROP INDEX IF EXISTS idx_parameters_data_type;
DROP INDEX IF EXISTS idx_parameters_name;

-- Drop tables in reverse order
DROP TABLE IF EXISTS parameter_rule_conditions;
DROP TABLE IF EXISTS parameter_rules;
DROP TABLE IF EXISTS parameter_conditions;
DROP TABLE IF EXISTS parameters;

-- Drop enum types
DROP TYPE IF EXISTS rule_type;
DROP TYPE IF EXISTS condition_match_type;
DROP TYPE IF EXISTS parameter_data_type; 