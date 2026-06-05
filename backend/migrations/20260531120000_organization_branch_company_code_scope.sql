-- Migration: organization branch code scoped by company
-- Purpose: allow each company in the same shop to have its own branch code series.
-- UP

BEGIN;

DROP INDEX IF EXISTS idxbranchcode;

DO $$
BEGIN
    IF to_regclass('public.organization_branches') IS NOT NULL THEN
        CREATE UNIQUE INDEX IF NOT EXISTS idxbranchshopcompanycode
            ON organization_branches(holdingcode, companyguid, code)
            WHERE deletedat IS NULL;

        CREATE INDEX IF NOT EXISTS idxbranchshopcompany
            ON organization_branches(holdingcode, companyguid)
            WHERE deletedat IS NULL;
    END IF;
END $$;

COMMIT;

-- DOWN (manual rollback)
-- BEGIN;
-- DROP INDEX IF EXISTS idxbranchshopcompany;
-- DROP INDEX IF EXISTS idxbranchshopcompanycode;
-- CREATE UNIQUE INDEX IF NOT EXISTS idxbranchcode
--     ON organization_branches(code);
-- COMMIT;
