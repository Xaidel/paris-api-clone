DROP TABLE IF EXISTS transaction_processing_queue;

ALTER TABLE transactions
    DROP CONSTRAINT IF EXISTS transactions_status_check;

ALTER TABLE transactions
    DROP CONSTRAINT IF EXISTS transactions_classification_check;

ALTER TABLE transactions
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS classification;
