CREATE TABLE transaction_step_5_data (
    transaction_id                         UUID        PRIMARY KEY REFERENCES transactions (id) ON DELETE CASCADE,
    screening_question_1_answer           BOOLEAN     NOT NULL,
    screening_question_1_justification    TEXT        NOT NULL CHECK (LENGTH(BTRIM(screening_question_1_justification)) > 0),
    screening_question_2_answer           BOOLEAN     NOT NULL,
    screening_question_2_justification    TEXT        NOT NULL CHECK (LENGTH(BTRIM(screening_question_2_justification)) > 0),
    reviewer_notes                        TEXT,
    is_final                              BOOLEAN     NOT NULL,
    created_at                            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
