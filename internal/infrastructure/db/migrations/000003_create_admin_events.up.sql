CREATE TABLE IF NOT EXISTS admin_event (
    id UUID PRIMARY KEY,
    timestamp TIMESTAMPTZ NOT NULL,
    user_id TEXT NOT NULL,
    group_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    event_data JSONB NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_admin_event_user_id ON admin_event (user_id);
CREATE INDEX IF NOT EXISTS idx_admin_event_event_type ON admin_event (event_type);
CREATE INDEX IF NOT EXISTS idx_admin_event_timestamp ON admin_event (timestamp);

CREATE TABLE IF NOT EXISTS admin_event_outbox (
    event_id UUID PRIMARY KEY,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);
