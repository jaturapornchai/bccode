-- Migration: organization branch code scoped by company
-- Purpose: allow each company in the same shop to have its own branch code series.
-- UP

BEGIN;

DROP INDEX IF EXISTS idx_branch_code;

DO $$
BEGIN
    IF to_regclass('public.organization_branches') IS NOT NULL THEN
        CREATE UNIQUE INDEX IF NOT EXISTS idx_branch_shop_company_code
            ON organization_branches(holding_code, company_guid, code)
            WHERE deleted_at IS NULL;

        CREATE INDEX IF NOT EXISTS idx_branch_shop_company
            ON organization_branches(holding_code, company_guid)
            WHERE deleted_at IS NULL;
    END IF;
END $$;

COMMIT;

-- DOWN (manual rollback)
-- BEGIN;
-- DROP INDEX IF EXISTS idx_branch_shop_company;
-- DROP INDEX IF EXISTS idx_branch_shop_company_code;
-- CREATE UNIQUE INDEX IF NOT EXISTS idx_branch_code
--     ON organization_branches(code);
-- COMMIT;
