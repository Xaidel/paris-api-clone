ALTER TABLE transactions
    DROP CONSTRAINT IF EXISTS transactions_classification_check;

ALTER TABLE transactions
    DROP CONSTRAINT IF EXISTS transactions_status_check;

UPDATE transactions
SET classification = CASE classification
    WHEN 'u1' THEN 'aligned'
    WHEN 'short_tenor' THEN 'aligned'
    WHEN 'u2' THEN 'not-aligned'
    WHEN 'unclassifiable' THEN 'not-aligned'
    WHEN 'requireds-review' THEN 'not-aligned'
    WHEN 'unaligned' THEN 'not-aligned'
    ELSE classification
END
WHERE classification IN ('u1', 'u2', 'short_tenor', 'unclassifiable', 'requireds-review', 'unaligned');

UPDATE transactions
SET status = CASE status
    WHEN 'classified' THEN 'ai-reviewed'
    ELSE status
END
WHERE status IN ('classified');

ALTER TABLE transactions
    ADD CONSTRAINT transactions_classification_check
        CHECK (classification IN ('unclassified', 'aligned', 'not-aligned'));

ALTER TABLE transactions
    ADD CONSTRAINT transactions_status_check
        CHECK (status IN ('pending', 'processing', 'ai-reviewed', 'professionally-reviewed', 'failed'));
