DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_available_extensions WHERE name = 'vector') THEN
        EXECUTE 'CREATE EXTENSION IF NOT EXISTS vector';

        EXECUTE '
CREATE TABLE IF NOT EXISTS list_entry_embeddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_type TEXT NOT NULL,
    entry_text TEXT NOT NULL,
    embedding vector(1536),
    embedding_json JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_list_entry_embeddings_list_type_entry_text UNIQUE (list_type, entry_text)
)';

        EXECUTE '
CREATE INDEX IF NOT EXISTS idx_list_entry_embeddings_embedding_hnsw
    ON list_entry_embeddings USING hnsw (embedding vector_cosine_ops)';
    ELSE
        EXECUTE '
CREATE TABLE IF NOT EXISTS list_entry_embeddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_type TEXT NOT NULL,
    entry_text TEXT NOT NULL,
    embedding_json JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_list_entry_embeddings_list_type_entry_text UNIQUE (list_type, entry_text)
)';
    END IF;
END $$;
