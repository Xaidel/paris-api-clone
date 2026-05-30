CREATE TABLE bug_reports (
    id             CHAR(32)     PRIMARY KEY,
    user_id        TEXT         NOT NULL REFERENCES "user" (id),
    transaction_id UUID         NOT NULL REFERENCES transactions (id),
    title          TEXT         NOT NULL,
    description    TEXT         NOT NULL,
    status         VARCHAR(16)  NOT NULL DEFAULT 'Open' CHECK (status IN ('Open', 'Closed')),
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bug_reports_user_id ON bug_reports (user_id);
CREATE INDEX idx_bug_reports_transaction_id ON bug_reports (transaction_id);
CREATE INDEX idx_bug_reports_status ON bug_reports (status);
CREATE INDEX idx_bug_reports_created_at ON bug_reports (created_at DESC);
