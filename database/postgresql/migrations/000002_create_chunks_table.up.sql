CREATE TABLE IF NOT EXISTS chunks (
    id BIGINT PRIMARY KEY,
    resource_id BIGINT NOT NULL,
    chunk_id VARCHAR(255) NOT NULL,
    minio_url TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (resource_id) REFERENCES resources(id)
);

CREATE INDEX idx_chunks_resource_id ON chunks(resource_id);
CREATE INDEX idx_chunks_created_at ON chunks(created_at);
