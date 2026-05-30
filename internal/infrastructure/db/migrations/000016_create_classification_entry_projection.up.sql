DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_available_extensions WHERE name = 'vector') THEN
        EXECUTE 'CREATE EXTENSION IF NOT EXISTS vector';

        EXECUTE '
CREATE TABLE IF NOT EXISTS classification_entry (
    entry_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    list_type TEXT NOT NULL,
    source_row_id TEXT NOT NULL,
    canonical_text TEXT NOT NULL,
    content_hash TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_classification_entry_list_type_source_row_id UNIQUE (list_type, source_row_id)
)';

        EXECUTE '
CREATE TABLE IF NOT EXISTS classification_entry_embedding (
    embedding_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_id UUID NOT NULL REFERENCES classification_entry(entry_id) ON DELETE CASCADE,
    list_type TEXT NOT NULL,
    embedding_model TEXT NOT NULL,
    embedding_dim INTEGER NOT NULL,
    embedding_version TEXT NOT NULL,
    embedding vector(1536) NOT NULL,
    vector_norm DOUBLE PRECISION NOT NULL,
    content_hash_at_embedding TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    embedded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_classification_entry_embedding_entry_model_version UNIQUE (entry_id, embedding_model, embedding_version)
)';

        EXECUTE 'CREATE INDEX IF NOT EXISTS idx_classification_entry_list_type_active ON classification_entry (list_type, is_active)';
        EXECUTE 'CREATE INDEX IF NOT EXISTS idx_classification_entry_embedding_model_active ON classification_entry_embedding (embedding_model, embedding_version, list_type, is_active)';

        EXECUTE '
CREATE INDEX IF NOT EXISTS idx_classification_entry_embedding_u1_hnsw
    ON classification_entry_embedding USING hnsw (embedding vector_cosine_ops)
    WHERE is_active = TRUE AND list_type = ''u1''';

        EXECUTE '
CREATE INDEX IF NOT EXISTS idx_classification_entry_embedding_u2_hnsw
    ON classification_entry_embedding USING hnsw (embedding vector_cosine_ops)
    WHERE is_active = TRUE AND list_type = ''u2''';

        EXECUTE '
CREATE INDEX IF NOT EXISTS idx_classification_entry_embedding_sector_hnsw
    ON classification_entry_embedding USING hnsw (embedding vector_cosine_ops)
    WHERE is_active = TRUE AND list_type = ''sector''';
    END IF;
END $$;
