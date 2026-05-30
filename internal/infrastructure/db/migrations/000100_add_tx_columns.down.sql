ALTER TABLE transactions
    DROP COLUMN IF EXISTS transaction_value,
    DROP COLUMN IF EXISTS ref_num,
    DROP COLUMN IF EXISTS product,
    DROP COLUMN IF EXISTS processed_year,
    DROP COLUMN IF EXISTS processed_month,
    DROP COLUMN IF EXISTS dmc_ib,
    DROP COLUMN IF EXISTS dmc,
    DROP COLUMN IF EXISTS partner_bank;