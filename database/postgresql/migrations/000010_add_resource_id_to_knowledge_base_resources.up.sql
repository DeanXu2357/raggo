ALTER TABLE knowledge_base_resources
ADD COLUMN resource_id BIGINT NOT NULL,
ADD FOREIGN KEY (resource_id) REFERENCES resources(id);

CREATE INDEX idx_knowledge_base_resources_resource_id ON knowledge_base_resources(resource_id);
