CREATE TABLE IF NOT EXISTS exclusion_list (
    id UUID PRIMARY KEY,
    activity_type TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_exclusion_list_activity_type ON exclusion_list (activity_type);
