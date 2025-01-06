ALTER TABLE knowledge_bases
ADD COLUMN model_provider VARCHAR(255) NOT NULL DEFAULT 'openai';

-- Insert test data for ollama and nomic-embed-text
INSERT INTO knowledge_bases (id, name, description, embedding_model, model_provider)
VALUES (
    1,
    'Test Knowledge Base',
    'Test knowledge base using ollama and nomic-embed-text',
    'nomic-embed-text',
    'ollama'
);
