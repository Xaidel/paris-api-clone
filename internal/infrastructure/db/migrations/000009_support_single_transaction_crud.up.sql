ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS id UUID;

UPDATE transactions
SET id = (
    substr(md5(coalesce(upload_id::text, '') || ':' || coalesce(row_number::text, '')), 1, 8) || '-' ||
    substr(md5(coalesce(upload_id::text, '') || ':' || coalesce(row_number::text, '')), 9, 4) || '-' ||
    substr(md5(coalesce(upload_id::text, '') || ':' || coalesce(row_number::text, '')), 13, 4) || '-' ||
    substr(md5(coalesce(upload_id::text, '') || ':' || coalesce(row_number::text, '')), 17, 4) || '-' ||
    substr(md5(coalesce(upload_id::text, '') || ':' || coalesce(row_number::text, '')), 21, 12)
)::uuid
WHERE id IS NULL;

ALTER TABLE transactions
    ALTER COLUMN id SET NOT NULL;

ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ;

UPDATE transactions
SET updated_at = created_at
WHERE updated_at IS NULL;

ALTER TABLE transactions
    ALTER COLUMN updated_at SET NOT NULL;

ALTER TABLE transactions
    DROP CONSTRAINT IF EXISTS transactions_pkey;

ALTER TABLE transactions
    ALTER COLUMN upload_id DROP NOT NULL,
    ALTER COLUMN row_number DROP NOT NULL;

ALTER TABLE transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_upload_row_unique
    ON transactions (upload_id, row_number)
    WHERE upload_id IS NOT NULL AND row_number IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_transactions_upload_id_created_at
    ON transactions (upload_id, created_at);
