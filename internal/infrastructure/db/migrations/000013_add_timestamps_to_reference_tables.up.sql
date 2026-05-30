ALTER TABLE u1_list
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS created_by TEXT REFERENCES "user"(id);

UPDATE u1_list
SET created_by = COALESCE(created_by, '01962b8f-aeb2-7e03-a8ff-1edce1300002');

UPDATE u1_list
SET created_at = COALESCE(created_at, NOW()),
    updated_at = COALESCE(updated_at, COALESCE(created_at, NOW()));

ALTER TABLE u1_list
    ALTER COLUMN created_by SET NOT NULL,
    ALTER COLUMN created_at SET NOT NULL,
    ALTER COLUMN updated_at SET NOT NULL;

ALTER TABLE sector
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS created_by TEXT REFERENCES "user"(id);

UPDATE sector
SET created_by = COALESCE(created_by, '01962b8f-aeb2-7e03-a8ff-1edce1300002');

UPDATE sector
SET created_at = COALESCE(created_at, NOW()),
    updated_at = COALESCE(updated_at, COALESCE(created_at, NOW()));

ALTER TABLE sector
    ALTER COLUMN created_by SET NOT NULL,
    ALTER COLUMN created_at SET NOT NULL,
    ALTER COLUMN updated_at SET NOT NULL;

ALTER TABLE exclusion_list
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS created_by TEXT REFERENCES "user"(id);

UPDATE exclusion_list
SET created_by = COALESCE(created_by, '01962b8f-aeb2-7e03-a8ff-1edce1300002');

UPDATE exclusion_list
SET created_at = COALESCE(created_at, NOW()),
    updated_at = COALESCE(updated_at, COALESCE(created_at, NOW()));

ALTER TABLE exclusion_list
    ALTER COLUMN created_by SET NOT NULL,
    ALTER COLUMN created_at SET NOT NULL,
    ALTER COLUMN updated_at SET NOT NULL;

ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS created_by TEXT REFERENCES "user"(id),
    ADD COLUMN IF NOT EXISTS pipeline_result JSONB,
    ADD COLUMN IF NOT EXISTS failure_reason TEXT;

UPDATE transactions
SET created_by = COALESCE(created_by, '01962b8f-aeb2-7e03-a8ff-1edce1300002');

ALTER TABLE transactions
    ALTER COLUMN created_by SET NOT NULL;
