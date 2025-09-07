-- Drop indexes
DROP INDEX IF EXISTS idx_experiment_variant_parameters_experiment_id;
DROP INDEX IF EXISTS idx_experiment_variant_parameters_parameter_id;
DROP INDEX IF EXISTS idx_experiment_variant_parameters_experiment_variant_id;
DROP INDEX IF EXISTS idx_experiment_variants_experiment_id;
DROP INDEX IF EXISTS idx_experiments_hash_attribute_id;
DROP INDEX IF EXISTS idx_experiments_status;
DROP INDEX IF EXISTS idx_experiments_uuid;

-- Drop tables in reverse order
DROP TABLE IF EXISTS experiment_variant_parameters;
DROP TABLE IF EXISTS experiment_variants;
DROP TABLE IF EXISTS experiments;
