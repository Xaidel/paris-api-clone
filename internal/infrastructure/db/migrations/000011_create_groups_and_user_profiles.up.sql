CREATE TABLE IF NOT EXISTS user_group (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS user_profile (
    user_id TEXT PRIMARY KEY REFERENCES "user"(id) ON DELETE CASCADE,
    first_name TEXT NOT NULL,
    middle_name TEXT NULL,
    last_name TEXT NOT NULL,
    group_id TEXT NOT NULL REFERENCES user_group(id)
);

CREATE INDEX IF NOT EXISTS idx_user_profile_group_id ON user_profile(group_id);
