-- Migration: organization branch code scoped by company
-- Purpose: allow each company in the same shop to have its own branch code series.
-- UP

BEGIN;

DROP INDEX IF EXISTS idx_branchcode;

DO $$
BEGIN
    IF to_regclass('public.organization_branches') IS NOT NULL THEN
        CREATE UNIQUE INDEX IF NOT EXISTS idx_branch_shop_company_code
            ON organization_branches(holdingcode, companyguid, code)
            WHERE deletedat IS NULL;

        CREATE INDEX IF NOT EXISTS idx_branch_shop_company
            ON organization_branches(holdingcode, companyguid)
            WHERE deletedat IS NULL;
    END IF;
END $$;

COMMIT;

-- DOWN (manual rollback)
-- BEGIN;
-- DROP INDEX IF EXISTS idx_branch_shop_company;
-- DROP INDEX IF EXISTS idx_branch_shop_company_code;
-- CREATE UNIQUE INDEX IF NOT EXISTS idx_branchcode
--     ON organization_branches(code);
-- COMMIT;
