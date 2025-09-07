-- Drop trigger
DROP TRIGGER IF EXISTS update_segments_updated_at ON segments;

-- Drop indexes
DROP INDEX IF EXISTS idx_segment_rule_conditions_attribute_id;
DROP INDEX IF EXISTS idx_segment_rule_conditions_rule_id;
DROP INDEX IF EXISTS idx_segment_rules_segment_id;
DROP INDEX IF EXISTS idx_segments_name;

-- Drop tables in reverse order (due to foreign key constraints)
DROP TABLE IF EXISTS segment_rule_conditions;
DROP TABLE IF EXISTS segment_rules;
DROP TABLE IF EXISTS segments;

-- Drop enum type
DROP TYPE IF EXISTS condition_operator; 