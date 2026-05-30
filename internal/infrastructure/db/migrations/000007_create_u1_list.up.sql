CREATE TABLE IF NOT EXISTS u1_list (
    id UUID PRIMARY KEY,
    sector TEXT NOT NULL,
    eligible_operation_type TEXT NOT NULL,
    condition_guidance TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_u1_list_sector ON u1_list (sector);
CREATE INDEX IF NOT EXISTS idx_u1_list_eligible_operation_type ON u1_list (eligible_operation_type);
