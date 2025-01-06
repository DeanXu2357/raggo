CREATE TABLE IF NOT EXISTS knowledge_base_resources (
    id BIGINT PRIMARY KEY,
    knowledge_base_id BIGINT NOT NULL,
    chunk_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    context_description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (knowledge_base_id) REFERENCES knowledge_bases(id),
    FOREIGN KEY (chunk_id) REFERENCES chunks(id)
);

CREATE INDEX idx_knowledge_base_resources_knowledge_base_id ON knowledge_base_resources(knowledge_base_id);
CREATE INDEX idx_knowledge_base_resources_chunk_id ON knowledge_base_resources(chunk_id);
CREATE INDEX idx_knowledge_base_resources_created_at ON knowledge_base_resources(created_at);
