ALTER TABLE transactions
    DROP CONSTRAINT IF EXISTS transactions_classification_check;

ALTER TABLE transactions
    DROP CONSTRAINT IF EXISTS transactions_status_check;

UPDATE transactions
SET classification = CASE classification
    WHEN 'u1' THEN 'aligned'
    WHEN 'unclassifiable' THEN 'requireds-review'
    WHEN 'u2' THEN 'unaligned'
    ELSE classification
END
WHERE classification IN ('u1', 'u2', 'unclassifiable');

UPDATE transactions
SET status = CASE status
    WHEN 'pending' THEN 'uploaded'
    WHEN 'classified' THEN 'ai-reviewed'
    WHEN 'failed' THEN 'professionally-reviewed'
    ELSE status
END
WHERE status IN ('pending', 'classified', 'failed');

ALTER TABLE transactions
    ADD CONSTRAINT transactions_classification_check
        CHECK (classification IN ('unclassified', 'aligned', 'requireds-review', 'unaligned'));

ALTER TABLE transactions
    ADD CONSTRAINT transactions_status_check
        CHECK (status IN ('uploaded', 'processing', 'ai-reviewed', 'professionally-reviewed'));
