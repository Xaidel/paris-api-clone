CREATE TABLE IF NOT EXISTS sector (
    id UUID PRIMARY KEY,
    type TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sector_type ON sector (type);
CREATE INDEX IF NOT EXISTS idx_sector_name ON sector (name);
