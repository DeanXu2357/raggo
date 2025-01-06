-- Remove test data first
DELETE FROM knowledge_bases WHERE id = 1;

-- Then remove the column
ALTER TABLE knowledge_bases DROP COLUMN model_provider;
