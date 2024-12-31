CREATE TABLE translated_resources (
    id BIGINT PRIMARY KEY,
    original_resource_id BIGINT NOT NULL REFERENCES resources(id),
    filename VARCHAR(255) NOT NULL,
    minio_url VARCHAR(255) NOT NULL,
    source_language VARCHAR(50) NOT NULL,
    target_language VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_translated_resources_original_resource_id ON translated_resources(original_resource_id);
