CREATE TABLE transaction_step_4 (
    transaction_id      UUID        PRIMARY KEY REFERENCES transactions (id) ON DELETE CASCADE,
    sector_id           UUID        NOT NULL REFERENCES sector (id),
    additional_context  TEXT        NOT NULL,
    is_high_emitting    BOOLEAN     NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transaction_step_4_sector_id ON transaction_step_4 (sector_id);
