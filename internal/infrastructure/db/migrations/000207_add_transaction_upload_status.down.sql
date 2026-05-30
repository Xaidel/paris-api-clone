ALTER TABLE transaction_upload DROP CONSTRAINT IF EXISTS chk_transaction_upload_status;

ALTER TABLE transaction_upload DROP COLUMN IF EXISTS status;
