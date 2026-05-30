ALTER TABLE transaction_upload ADD COLUMN status TEXT;

UPDATE transaction_upload SET status = 'uploaded' WHERE status IS NULL;

ALTER TABLE transaction_upload ALTER COLUMN status SET NOT NULL;

ALTER TABLE transaction_upload
    ADD CONSTRAINT chk_transaction_upload_status
    CHECK (status IN ('uploaded', 'failed'));
