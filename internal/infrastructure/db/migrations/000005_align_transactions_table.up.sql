DROP TABLE IF EXISTS transaction_row;

CREATE TABLE IF NOT EXISTS transactions (
    upload_id UUID NOT NULL REFERENCES transaction_upload (id) ON DELETE CASCADE,
    row_number INTEGER NOT NULL,
    transaction_count INTEGER NOT NULL,
    goods_description TEXT NOT NULL,
    goods_classification TEXT NOT NULL,
    applicant_country TEXT NOT NULL,
    beneficiary_country TEXT NOT NULL,
    source_country TEXT NOT NULL,
    destination_country TEXT NOT NULL,
    tenor_description TEXT NOT NULL,
    es_category TEXT NOT NULL,
    pa_alignment TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (upload_id, row_number)
);

CREATE INDEX IF NOT EXISTS idx_transactions_upload_id ON transactions (upload_id);
