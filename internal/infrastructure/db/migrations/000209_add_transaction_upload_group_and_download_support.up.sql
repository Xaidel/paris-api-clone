ALTER TABLE transaction_upload
    DROP CONSTRAINT IF EXISTS transaction_upload_content_md5_key;

UPDATE transaction_upload tu
SET group_id = up.group_id
FROM (
    SELECT DISTINCT ON (t.upload_id)
        t.upload_id,
        up.group_id
    FROM transactions t
    JOIN user_profile up ON up.user_id = t.created_by
    WHERE up.group_id IS NOT NULL
    ORDER BY t.upload_id, t.created_at ASC, t.row_number ASC
) up
WHERE tu.id = up.upload_id
  AND tu.group_id = '01962b8f-aeb2-7e03-a8ff-1edce1300001';

-- Legacy uploads that still cannot infer ownership remain assigned
-- to the seeded superadmin group set by 000208.

ALTER TABLE transaction_upload
    ADD CONSTRAINT uq_transaction_upload_group_content_md5
    UNIQUE (group_id, content_md5);
