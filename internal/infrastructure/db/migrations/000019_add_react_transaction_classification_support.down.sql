DROP INDEX IF EXISTS idx_transaction_processing_queue_task_name_created_at;

ALTER TABLE transaction_processing_queue
    DROP CONSTRAINT IF EXISTS transaction_processing_queue_pkey;

ALTER TABLE transaction_processing_queue
    ADD CONSTRAINT transaction_processing_queue_pkey PRIMARY KEY (transaction_id);

ALTER TABLE transaction_processing_queue
    DROP COLUMN IF EXISTS task_name;

DROP INDEX IF EXISTS idx_transaction_description_embeddings_embedding_hnsw;
DROP INDEX IF EXISTS idx_transaction_description_embeddings_exact_lookup;
DROP TABLE IF EXISTS transaction_description_embeddings;

ALTER TABLE transactions
    DROP CONSTRAINT IF EXISTS transactions_classification_check;

ALTER TABLE transactions
    DROP CONSTRAINT IF EXISTS transactions_status_check;

UPDATE transactions
SET classification = 'unclassified'
WHERE classification = 'next_step';

ALTER TABLE transactions
    ADD CONSTRAINT transactions_classification_check
        CHECK (classification IN ('unclassified', 'aligned', 'not-aligned'));

ALTER TABLE transactions
    ADD CONSTRAINT transactions_status_check
        CHECK (status IN ('pending', 'processing', 'ai-reviewed', 'professionally-reviewed', 'failed'));
