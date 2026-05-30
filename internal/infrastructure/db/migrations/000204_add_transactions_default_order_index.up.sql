CREATE INDEX IF NOT EXISTS idx_transactions_created_at_upload_row_id_order
    ON transactions (created_at DESC, upload_id ASC NULLS LAST, row_number ASC NULLS LAST, id DESC);