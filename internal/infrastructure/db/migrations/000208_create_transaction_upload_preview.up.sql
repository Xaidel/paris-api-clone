ALTER TABLE transaction_upload ADD COLUMN group_id TEXT REFERENCES user_group (id);

UPDATE transaction_upload SET group_id = (SELECT id FROM user_group WHERE name = 'superadmin') WHERE group_id IS NULL;

ALTER TABLE transaction_upload ALTER COLUMN group_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_transaction_upload_group_id ON transaction_upload (group_id);

CREATE TABLE IF NOT EXISTS transaction_upload_preview (
    upload_id UUID PRIMARY KEY REFERENCES transaction_upload (id) ON DELETE CASCADE,
    columns JSONB NOT NULL,
    rows JSONB NOT NULL,
    total_rows INTEGER NOT NULL,
    validation_errors JSONB NOT NULL
);
