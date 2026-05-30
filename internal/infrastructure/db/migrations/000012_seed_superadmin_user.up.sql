INSERT INTO user_group (id, name)
VALUES ('01962b8f-aeb2-7e03-a8ff-1edce1300001', 'superadmin')
ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name;

INSERT INTO "user" (id, username, password_hash, created_at, updated_at)
VALUES (
    '01962b8f-aeb2-7e03-a8ff-1edce1300002',
    'superadmih',
    '$2a$10$0mV4oN0A/1hB4WKyYVJZZu1icJv0Yx6S5nYEYTbSsfXCMVvjJAnC2',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO UPDATE
SET username = EXCLUDED.username,
    password_hash = EXCLUDED.password_hash,
    updated_at = NOW();

INSERT INTO user_profile (user_id, first_name, middle_name, last_name, group_id)
SELECT
    '01962b8f-aeb2-7e03-a8ff-1edce1300002',
    'Super',
    NULL,
    'Admin',
    id
FROM user_group
WHERE name = 'superadmin'
ON CONFLICT (user_id) DO UPDATE
SET first_name = EXCLUDED.first_name,
    middle_name = EXCLUDED.middle_name,
    last_name = EXCLUDED.last_name,
    group_id = EXCLUDED.group_id;
