ALTER TABLE transactions
    DROP CONSTRAINT IF EXISTS transactions_classification_check;

ALTER TABLE transactions
    DROP CONSTRAINT IF EXISTS transactions_status_check;

ALTER TABLE transactions
    ADD CONSTRAINT transactions_classification_check
        CHECK (classification IN ('unclassified', 'aligned', 'not-aligned', 'next_step'));

ALTER TABLE transactions
    ADD CONSTRAINT transactions_status_check
        CHECK (status IN ('pending', 'processing', 'ai-reviewed', 'from-previous-transactions', 'professionally-reviewed', 'failed'));

ALTER TABLE transaction_processing_queue
    ADD COLUMN IF NOT EXISTS task_name TEXT NOT NULL DEFAULT 'transaction:classify';

ALTER TABLE transaction_processing_queue
    DROP CONSTRAINT IF EXISTS transaction_processing_queue_pkey;

ALTER TABLE transaction_processing_queue
    ADD CONSTRAINT transaction_processing_queue_pkey PRIMARY KEY (task_name, transaction_id);

CREATE INDEX IF NOT EXISTS idx_transaction_processing_queue_task_name_created_at
    ON transaction_processing_queue (task_name, created_at);

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_available_extensions WHERE name = 'vector') THEN
        EXECUTE 'CREATE EXTENSION IF NOT EXISTS vector';

        EXECUTE '
CREATE TABLE IF NOT EXISTS transaction_description_embeddings (
    transaction_id UUID NOT NULL REFERENCES transactions (id) ON DELETE CASCADE,
    goods_description TEXT NOT NULL,
    classifier_family TEXT NOT NULL,
    classifier_version TEXT NOT NULL,
    embedding_model TEXT NOT NULL,
    embedding_version TEXT NOT NULL,
    embedding vector(1536) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT transaction_description_embeddings_pkey PRIMARY KEY (transaction_id, classifier_family, classifier_version, embedding_model, embedding_version)
)';

        EXECUTE 'CREATE INDEX IF NOT EXISTS idx_transaction_description_embeddings_exact_lookup ON transaction_description_embeddings (goods_description, classifier_family, classifier_version)';
        EXECUTE 'CREATE INDEX IF NOT EXISTS idx_transaction_description_embeddings_embedding_hnsw ON transaction_description_embeddings USING hnsw (embedding vector_cosine_ops)';
    END IF;
END $$;
