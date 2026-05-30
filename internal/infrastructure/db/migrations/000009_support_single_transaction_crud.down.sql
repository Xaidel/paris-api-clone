DROP INDEX IF EXISTS idx_transactions_upload_id_created_at;
DROP INDEX IF EXISTS idx_transactions_upload_row_unique;

DELETE FROM transactions
WHERE upload_id IS NULL OR row_number IS NULL;

ALTER TABLE transactions
    DROP CONSTRAINT IF EXISTS transactions_pkey;

ALTER TABLE transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (upload_id, row_number);

ALTER TABLE transactions
    ALTER COLUMN upload_id SET NOT NULL,
    ALTER COLUMN row_number SET NOT NULL;

ALTER TABLE transactions
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS id;
