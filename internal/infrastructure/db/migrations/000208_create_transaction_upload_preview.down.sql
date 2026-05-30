DROP TABLE IF EXISTS transaction_upload_preview;

DROP INDEX IF EXISTS idx_transaction_upload_group_id;

ALTER TABLE transaction_upload DROP COLUMN IF EXISTS group_id;
