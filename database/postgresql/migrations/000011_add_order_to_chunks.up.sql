ALTER TABLE chunks
ADD COLUMN chunk_order INTEGER NOT NULL DEFAULT 0;

CREATE INDEX idx_chunks_order ON chunks(chunk_order);
