CREATE TABLE transaction_feedback (
    id             UUID         PRIMARY KEY,
    user_id        TEXT         NOT NULL REFERENCES "user" (id),
    transaction_id UUID         NOT NULL REFERENCES transactions (id) ON DELETE CASCADE,
    kind           VARCHAR(16)  NOT NULL CHECK (kind IN ('thumbs_up', 'thumbs_down')),
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, transaction_id)
);

CREATE INDEX idx_transaction_feedback_transaction_id ON transaction_feedback (transaction_id);
CREATE INDEX idx_transaction_feedback_user_id ON transaction_feedback (user_id);
