ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS retry_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS last_retried_at TIMESTAMPTZ;

CREATE TABLE IF NOT EXISTS transaction_classification_attempt (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES transactions (id) ON DELETE CASCADE,
    upload_id UUID NOT NULL REFERENCES transaction_upload (id) ON DELETE CASCADE,
    task_name TEXT NOT NULL,
    parent_job_id UUID REFERENCES transaction_classification_attempt (id) ON DELETE SET NULL,
    retry_count INTEGER NOT NULL,
    last_retried_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT transaction_classification_attempt_retry_count_check CHECK (retry_count >= 1),
    CONSTRAINT transaction_classification_attempt_task_name_check CHECK (task_name IN ('transaction:classify', 'transaction:classify-react')),
    CONSTRAINT transaction_classification_attempt_unique_retry UNIQUE (transaction_id, retry_count)
);

CREATE INDEX IF NOT EXISTS idx_transaction_classification_attempt_upload_id_created_at
    ON transaction_classification_attempt (upload_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_transaction_classification_attempt_transaction_id_created_at
    ON transaction_classification_attempt (transaction_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_transaction_classification_attempt_parent_job_id
    ON transaction_classification_attempt (parent_job_id);
