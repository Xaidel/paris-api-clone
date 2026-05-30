ALTER TABLE transactions
    DROP CONSTRAINT IF EXISTS transactions_classification_check;

ALTER TABLE transactions
    DROP CONSTRAINT IF EXISTS transactions_status_check;

UPDATE transactions
SET classification = CASE classification
    WHEN 'aligned' THEN 'u1'
    WHEN 'not-aligned' THEN 'u2'
    ELSE classification
END
WHERE classification IN ('aligned', 'not-aligned');

UPDATE transactions
SET status = CASE status
    WHEN 'ai-reviewed' THEN 'classified'
    ELSE status
END
WHERE status IN ('ai-reviewed');

ALTER TABLE transactions
    ADD CONSTRAINT transactions_classification_check
        CHECK (classification IN ('unclassified', 'u1', 'u2', 'short_tenor', 'unclassifiable'));

ALTER TABLE transactions
    ADD CONSTRAINT transactions_status_check
        CHECK (status IN ('pending', 'processing', 'classified', 'failed'));
