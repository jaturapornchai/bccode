-- Migration: productbarcode projection classification support
-- Purpose: keep PostgreSQL productbarcode projections compatible with product sets.
-- UP

CREATE TABLE IF NOT EXISTS productbarcode (
    holdingcode TEXT NOT NULL,
    barcode TEXT NOT NULL,
    parid TEXT DEFAULT '',
    guidfixed TEXT DEFAULT '',
    itemcode TEXT DEFAULT '',
    name0 TEXT DEFAULT '',
    names JSONB NOT NULL DEFAULT '[]'::jsonb,
    unitcode TEXT DEFAULT '',
    unitname TEXT DEFAULT '',
    unitnames JSONB NOT NULL DEFAULT '[]'::jsonb,
    groupcode TEXT DEFAULT '',
    groupnames TEXT DEFAULT '',
    group_code TEXT DEFAULT '',
    group_names JSONB NOT NULL DEFAULT '[]'::jsonb,
    groupsubonecode TEXT DEFAULT '',
    groupsubonenames TEXT DEFAULT '',
    groupsubtwocode TEXT DEFAULT '',
    groupsubtwonames TEXT DEFAULT '',
    categorycode TEXT DEFAULT '',
    categorynames TEXT DEFAULT '',
    category_names JSONB NOT NULL DEFAULT '[]'::jsonb,
    brandcode TEXT DEFAULT '',
    brand_code TEXT DEFAULT '',
    brandnames TEXT DEFAULT '',
    classcode TEXT DEFAULT '',
    classnames TEXT DEFAULT '',
    designcode TEXT DEFAULT '',
    designnames TEXT DEFAULT '',
    gradecode TEXT DEFAULT '',
    gradenames TEXT DEFAULT '',
    modelcode TEXT DEFAULT '',
    modelnames TEXT DEFAULT '',
    patterncode TEXT DEFAULT '',
    patternnames TEXT DEFAULT '',
    barcoderef TEXT DEFAULT '',
    mainbarcoderef TEXT DEFAULT '',
    barcoderefunitstand NUMERIC NOT NULL DEFAULT 1,
    barcoderefunitdivide NUMERIC NOT NULL DEFAULT 1,
    standvalue NUMERIC NOT NULL DEFAULT 1,
    dividevalue NUMERIC NOT NULL DEFAULT 1,
    balance_qty NUMERIC NOT NULL DEFAULT 0,
    balanceamount NUMERIC NOT NULL DEFAULT 0,
    averagecost NUMERIC NOT NULL DEFAULT 0,
    price1 NUMERIC NOT NULL DEFAULT 0,
    price_retail NUMERIC NOT NULL DEFAULT 0,
    isstock SMALLINT NOT NULL DEFAULT 0,
    isusesubbarcodes BOOLEAN NOT NULL DEFAULT FALSE,
    itemtype SMALLINT NOT NULL DEFAULT 0,
    materialtype SMALLINT NOT NULL DEFAULT 0,
    imageuri TEXT DEFAULT '',
    checksum TEXT DEFAULT '',
    bom JSONB NOT NULL DEFAULT '[]'::jsonb,
    createdat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updatedat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (holdingcode, barcode)
);

ALTER TABLE IF EXISTS productbarcode
    ADD COLUMN IF NOT EXISTS materialtype SMALLINT NOT NULL DEFAULT 0;

ALTER TABLE IF EXISTS productbarcode
    ADD COLUMN IF NOT EXISTS itemtype SMALLINT NOT NULL DEFAULT 0;

DO $$
BEGIN
    IF to_regclass('public.productbarcode') IS NOT NULL THEN
        IF NOT EXISTS (
            SELECT 1
            FROM pg_constraint
            WHERE conname = 'chk_productbarcode_itemtype'
              AND conrelid = 'public.productbarcode'::regclass
        ) THEN
            ALTER TABLE productbarcode
                ADD CONSTRAINT chk_productbarcode_itemtype
                CHECK (itemtype IN (0, 1, 2, 3));
        END IF;

        IF NOT EXISTS (
            SELECT 1
            FROM pg_constraint
            WHERE conname = 'chk_productbarcode_materialtype'
              AND conrelid = 'public.productbarcode'::regclass
        ) THEN
            ALTER TABLE productbarcode
                ADD CONSTRAINT chk_productbarcode_materialtype
                CHECK (materialtype IN (0, 1, 2, 3, 4));
        END IF;

        IF NOT EXISTS (
            SELECT 1
            FROM pg_constraint
            WHERE conname = 'chk_productbarcode_set_materialtype'
              AND conrelid = 'public.productbarcode'::regclass
        ) THEN
            ALTER TABLE productbarcode
                ADD CONSTRAINT chk_productbarcode_set_materialtype
                CHECK (itemtype <> 2 OR materialtype = 3);
        END IF;
    END IF;
END $$;

DO $$
BEGIN
    IF to_regclass('public.productbarcode') IS NOT NULL THEN
        CREATE INDEX IF NOT EXISTS idx_productbarcode_shop_barcode
            ON productbarcode(holdingcode, barcode);

        CREATE INDEX IF NOT EXISTS idx_productbarcode_shop_item_material
            ON productbarcode(holdingcode, itemtype, materialtype);
    END IF;
END $$;

-- DOWN (manual rollback)
-- DROP INDEX IF EXISTS idx_productbarcode_shop_item_material;
-- DROP INDEX IF EXISTS idx_productbarcode_shop_barcode;
-- ALTER TABLE IF EXISTS productbarcode DROP CONSTRAINT IF EXISTS chk_productbarcode_set_materialtype;
-- ALTER TABLE IF EXISTS productbarcode DROP CONSTRAINT IF EXISTS chk_productbarcode_materialtype;
-- ALTER TABLE IF EXISTS productbarcode DROP CONSTRAINT IF EXISTS chk_productbarcode_itemtype;
-- ALTER TABLE IF EXISTS productbarcode DROP COLUMN IF EXISTS materialtype;
