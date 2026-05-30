DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM transaction_upload
        GROUP BY content_md5
        HAVING COUNT(*) > 1
    ) THEN
        RAISE EXCEPTION 'cannot rollback 000209: duplicate transaction_upload.content_md5 values exist';
    END IF;
END $$;

ALTER TABLE transaction_upload DROP CONSTRAINT IF EXISTS uq_transaction_upload_group_content_md5;

ALTER TABLE transaction_upload
    ADD CONSTRAINT transaction_upload_content_md5_key UNIQUE (content_md5);
