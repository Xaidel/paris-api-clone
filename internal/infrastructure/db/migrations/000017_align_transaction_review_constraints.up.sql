ALTER TABLE transactions
    DROP CONSTRAINT IF EXISTS transactions_classification_check;

ALTER TABLE transactions
    DROP CONSTRAINT IF EXISTS transactions_status_check;

UPDATE transactions
SET classification = CASE classification
    WHEN 'aligned' THEN 'u1'
    WHEN 'requireds-review' THEN 'unclassifiable'
    WHEN 'unaligned' THEN 'u2'
    ELSE classification
END
WHERE classification IN ('aligned', 'requireds-review', 'unaligned');

UPDATE transactions
SET status = CASE status
    WHEN 'uploaded' THEN 'pending'
    WHEN 'ai-reviewed' THEN 'classified'
    WHEN 'professionally-reviewed' THEN 'classified'
    ELSE status
END
WHERE status IN ('uploaded', 'ai-reviewed', 'professionally-reviewed');

ALTER TABLE transactions
    ADD CONSTRAINT transactions_classification_check
        CHECK (classification IN ('unclassified', 'u1', 'u2', 'short_tenor', 'unclassifiable'));

ALTER TABLE transactions
    ADD CONSTRAINT transactions_status_check
        CHECK (status IN ('pending', 'processing', 'classified', 'failed'));
