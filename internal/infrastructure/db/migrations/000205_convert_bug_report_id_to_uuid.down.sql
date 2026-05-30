ALTER TABLE bug_reports
    ALTER COLUMN id TYPE CHAR(32)
    USING REPLACE(id::text, '-', '');