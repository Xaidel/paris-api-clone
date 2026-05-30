DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_available_extensions WHERE name = 'vector') THEN
        EXECUTE 'CREATE EXTENSION IF NOT EXISTS vector';

        IF EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_name = 'list_entry_embeddings'
              AND column_name = 'embedding_json'
        ) AND NOT EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_name = 'list_entry_embeddings'
              AND column_name = 'embedding'
        ) THEN
            EXECUTE 'ALTER TABLE list_entry_embeddings ADD COLUMN embedding vector(1536)';
        END IF;

        IF NOT EXISTS (
            SELECT 1
            FROM pg_indexes
            WHERE tablename = 'list_entry_embeddings'
              AND indexname = 'idx_list_entry_embeddings_embedding_hnsw'
        ) THEN
            EXECUTE '
CREATE INDEX idx_list_entry_embeddings_embedding_hnsw
    ON list_entry_embeddings USING hnsw (embedding vector_cosine_ops)';
        END IF;

        IF EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_name = 'list_entry_embeddings'
              AND column_name = 'embedding_json'
        ) AND EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_name = 'list_entry_embeddings'
              AND column_name = 'embedding'
        ) THEN
            EXECUTE '
UPDATE list_entry_embeddings
SET embedding = (
        (
            ''['' || (
                SELECT string_agg(element_value, '','' ORDER BY element_index)
                FROM jsonb_array_elements_text(embedding_json) WITH ORDINALITY AS embedding_values(element_value, element_index)
            ) || '']''
        )::vector
    ),
    updated_at = NOW()
WHERE embedding IS NULL
  AND embedding_json IS NOT NULL
  AND jsonb_typeof(embedding_json) = ''array''
  AND jsonb_array_length(embedding_json) = 1536';
        END IF;
    END IF;
END $$;
