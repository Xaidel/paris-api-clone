DROP INDEX IF EXISTS idx_list_entry_embeddings_embedding_hnsw;

DROP TABLE IF EXISTS list_entry_embeddings;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector') THEN
        EXECUTE 'DROP EXTENSION IF EXISTS vector';
    END IF;
END $$;
