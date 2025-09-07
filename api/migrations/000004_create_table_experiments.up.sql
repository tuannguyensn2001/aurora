-- Create experiments table
CREATE TABLE experiments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    uuid VARCHAR(255) NOT NULL UNIQUE,
    hypothesis TEXT NOT NULL,
    description TEXT NOT NULL,
    start_date BIGINT NOT NULL,
    end_date BIGINT,
    hash_attribute_id INTEGER NOT NULL,
    population_size INTEGER NOT NULL DEFAULT 100,
    strategy VARCHAR(255) NOT NULL DEFAULT 'random',
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    status VARCHAR(255) NOT NULL DEFAULT 'draft'
);

-- Create experiment_variants table
CREATE TABLE experiment_variants (
    id SERIAL PRIMARY KEY,
    experiment_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    traffic_allocation INTEGER NOT NULL DEFAULT 0
);

-- Create experiment_variant_parameters table
CREATE TABLE experiment_variant_parameters (
    id SERIAL PRIMARY KEY,
    experiment_variant_id INTEGER NOT NULL,
    parameter_data_type VARCHAR(255) NOT NULL,
    parameter_id INTEGER NOT NULL,
    parameter_name VARCHAR(255) NOT NULL,
    rollout_value TEXT NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    experiment_id INTEGER NOT NULL
);

-- Create indexes for better performance
CREATE INDEX idx_experiments_uuid ON experiments(uuid);
CREATE INDEX idx_experiments_status ON experiments(status);
CREATE INDEX idx_experiments_hash_attribute_id ON experiments(hash_attribute_id);
CREATE INDEX idx_experiment_variants_experiment_id ON experiment_variants(experiment_id);
CREATE INDEX idx_experiment_variant_parameters_experiment_variant_id ON experiment_variant_parameters(experiment_variant_id);
CREATE INDEX idx_experiment_variant_parameters_parameter_id ON experiment_variant_parameters(parameter_id);
CREATE INDEX idx_experiment_variant_parameters_experiment_id ON experiment_variant_parameters(experiment_id);
