ALTER TABLE bug_reports
    ALTER COLUMN id TYPE UUID
    USING (
        CASE
            WHEN POSITION('-' IN id) > 0 THEN id::uuid
            ELSE (
                SUBSTRING(id FROM 1 FOR 8) || '-' ||
                SUBSTRING(id FROM 9 FOR 4) || '-' ||
                SUBSTRING(id FROM 13 FOR 4) || '-' ||
                SUBSTRING(id FROM 17 FOR 4) || '-' ||
                SUBSTRING(id FROM 21 FOR 12)
            )::uuid
        END
    );