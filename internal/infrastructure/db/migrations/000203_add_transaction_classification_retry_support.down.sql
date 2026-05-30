DROP INDEX IF EXISTS idx_transaction_classification_attempt_parent_job_id;
DROP INDEX IF EXISTS idx_transaction_classification_attempt_transaction_id_created_at;
DROP INDEX IF EXISTS idx_transaction_classification_attempt_upload_id_created_at;

DROP TABLE IF EXISTS transaction_classification_attempt;

ALTER TABLE transactions
    DROP COLUMN IF EXISTS last_retried_at,
    DROP COLUMN IF EXISTS retry_count;
