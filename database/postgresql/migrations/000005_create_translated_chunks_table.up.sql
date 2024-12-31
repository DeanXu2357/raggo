CREATE TABLE translated_chunks (
    id BIGINT PRIMARY KEY,
    translated_resource_id BIGINT NOT NULL REFERENCES translated_resources(id),
    original_chunk_id BIGINT NOT NULL REFERENCES chunks(id),
    chunk_id VARCHAR(255) NOT NULL,
    minio_url VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_translated_chunks_translated_resource_id ON translated_chunks(translated_resource_id);
CREATE INDEX idx_translated_chunks_original_chunk_id ON translated_chunks(original_chunk_id);
