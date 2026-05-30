ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS classification TEXT,
    ADD COLUMN IF NOT EXISTS status TEXT;

UPDATE transactions
SET classification = 'unclassified'
WHERE classification IS NULL;

UPDATE transactions
SET status = 'uploaded'
WHERE status IS NULL;

ALTER TABLE transactions
    ALTER COLUMN classification SET NOT NULL,
    ALTER COLUMN status SET NOT NULL;

ALTER TABLE transactions
    ADD CONSTRAINT transactions_classification_check
        CHECK (classification IN ('unclassified', 'aligned', 'requireds-review', 'unaligned'));

ALTER TABLE transactions
    ADD CONSTRAINT transactions_status_check
        CHECK (status IN ('uploaded', 'processing', 'ai-reviewed', 'professionally-reviewed'));

CREATE TABLE IF NOT EXISTS transaction_processing_queue (
    transaction_id UUID PRIMARY KEY REFERENCES transactions (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
