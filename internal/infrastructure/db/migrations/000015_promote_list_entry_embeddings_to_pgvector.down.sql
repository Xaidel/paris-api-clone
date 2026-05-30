DROP INDEX IF EXISTS idx_list_entry_embeddings_embedding_hnsw;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'list_entry_embeddings'
          AND column_name = 'embedding'
    ) THEN
        EXECUTE 'ALTER TABLE list_entry_embeddings DROP COLUMN embedding';
    END IF;
END $$;
