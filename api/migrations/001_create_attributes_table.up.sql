CREATE TYPE data_type AS ENUM ('boolean', 'string', 'number', 'enum');

CREATE TABLE attributes (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT NOT NULL,
    data_type data_type NOT NULL DEFAULT 'string',
    hash_attribute BOOLEAN NOT NULL DEFAULT false,
    enum_options TEXT[] DEFAULT '{}',
    usage_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index on name for faster lookups
CREATE INDEX idx_attributes_name ON attributes(name);

-- Create index on data_type for filtering
CREATE INDEX idx_attributes_data_type ON attributes(data_type);

-- Create trigger to automatically update updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_attributes_updated_at 
    BEFORE UPDATE ON attributes 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column(); 