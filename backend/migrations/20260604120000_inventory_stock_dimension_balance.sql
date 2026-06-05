-- Accounting stock remains product/warehouse/location based.
-- Marketplace availability and dimension-level selling prices are separate projections.

ALTER TABLE inventorystockbalances
    ADD COLUMN IF NOT EXISTS reservedqty NUMERIC(18,4) NOT NULL DEFAULT 0;

ALTER TABLE productcostingconfig
    ADD COLUMN IF NOT EXISTS cost_by_warehouse BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE inventorystockbalances
    DROP CONSTRAINT IF EXISTS inventorystockbalances_scope_dimension_unique,
    DROP CONSTRAINT IF EXISTS inventorystockbalancesholdingcodeitemcodewhcodelocationcodekey;

ALTER TABLE inventorystockbalances
    ADD CONSTRAINT inventorystockbalancesholdingcodeitemcodewhcodelocationcodekey
    UNIQUE (holdingcode, itemcode, whcode, locationcode);

CREATE TABLE IF NOT EXISTS marketplacestockbalances (
    id BIGSERIAL PRIMARY KEY,
    holdingcode TEXT NOT NULL,
    itemcode TEXT NOT NULL,
    dimensionkey TEXT NOT NULL DEFAULT '',
    dimensionvalues JSONB NOT NULL DEFAULT '{}'::jsonb,
    currentqty NUMERIC(18,4) NOT NULL DEFAULT 0,
    reservedqty NUMERIC(18,4) NOT NULL DEFAULT 0,
    availableqty NUMERIC(18,4) NOT NULL DEFAULT 0,
    source TEXT NOT NULL DEFAULT 'accounting',
    updatedat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (holdingcode, itemcode, dimensionkey)
);

CREATE TABLE IF NOT EXISTS marketplacedimensionprices (
    id BIGSERIAL PRIMARY KEY,
    holdingcode TEXT NOT NULL,
    itemcode TEXT NOT NULL,
    barcode TEXT NOT NULL DEFAULT '',
    dimensionkey TEXT NOT NULL DEFAULT '',
    dimensionvalues JSONB NOT NULL DEFAULT '{}'::jsonb,
    pricelevel TEXT NOT NULL DEFAULT '',
    marketplace TEXT NOT NULL DEFAULT '',
    currency VARCHAR(3) NOT NULL DEFAULT 'THB',
    price NUMERIC(18,4) NOT NULL DEFAULT 0,
    saleprice NUMERIC(18,4) NOT NULL DEFAULT 0,
    compareatprice NUMERIC(18,4) NOT NULL DEFAULT 0,
    effectivefrom TIMESTAMPTZ,
    effectiveto TIMESTAMPTZ,
    isactive BOOLEAN NOT NULL DEFAULT true,
    updatedat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (holdingcode, itemcode, barcode, dimensionkey, pricelevel, marketplace, currency)
);

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacestockbalances' AND column_name = 'itemcode')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacestockbalances' AND column_name = 'itemcode') THEN
        ALTER TABLE marketplacestockbalances RENAME COLUMN itemcode TO itemcode;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacestockbalances' AND column_name = 'dimensionkey')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacestockbalances' AND column_name = 'dimensionkey') THEN
        ALTER TABLE marketplacestockbalances RENAME COLUMN dimensionkey TO dimensionkey;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacestockbalances' AND column_name = 'dimensionvalues')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacestockbalances' AND column_name = 'dimensionvalues') THEN
        ALTER TABLE marketplacestockbalances RENAME COLUMN dimensionvalues TO dimensionvalues;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacestockbalances' AND column_name = 'currentqty')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacestockbalances' AND column_name = 'currentqty') THEN
        ALTER TABLE marketplacestockbalances RENAME COLUMN currentqty TO currentqty;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacestockbalances' AND column_name = 'reservedqty')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacestockbalances' AND column_name = 'reservedqty') THEN
        ALTER TABLE marketplacestockbalances RENAME COLUMN reservedqty TO reservedqty;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacestockbalances' AND column_name = 'availableqty')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacestockbalances' AND column_name = 'availableqty') THEN
        ALTER TABLE marketplacestockbalances RENAME COLUMN availableqty TO availableqty;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacestockbalances' AND column_name = 'updatedat')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacestockbalances' AND column_name = 'updatedat') THEN
        ALTER TABLE marketplacestockbalances RENAME COLUMN updatedat TO updatedat;
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacedimensionprices' AND column_name = 'itemcode')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacedimensionprices' AND column_name = 'itemcode') THEN
        ALTER TABLE marketplacedimensionprices RENAME COLUMN itemcode TO itemcode;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacedimensionprices' AND column_name = 'dimensionkey')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacedimensionprices' AND column_name = 'dimensionkey') THEN
        ALTER TABLE marketplacedimensionprices RENAME COLUMN dimensionkey TO dimensionkey;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacedimensionprices' AND column_name = 'dimensionvalues')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacedimensionprices' AND column_name = 'dimensionvalues') THEN
        ALTER TABLE marketplacedimensionprices RENAME COLUMN dimensionvalues TO dimensionvalues;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacedimensionprices' AND column_name = 'pricelevel')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacedimensionprices' AND column_name = 'pricelevel') THEN
        ALTER TABLE marketplacedimensionprices RENAME COLUMN pricelevel TO pricelevel;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacedimensionprices' AND column_name = 'saleprice')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacedimensionprices' AND column_name = 'saleprice') THEN
        ALTER TABLE marketplacedimensionprices RENAME COLUMN saleprice TO saleprice;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacedimensionprices' AND column_name = 'compareatprice')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacedimensionprices' AND column_name = 'compareatprice') THEN
        ALTER TABLE marketplacedimensionprices RENAME COLUMN compareatprice TO compareatprice;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacedimensionprices' AND column_name = 'effectivefrom')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacedimensionprices' AND column_name = 'effectivefrom') THEN
        ALTER TABLE marketplacedimensionprices RENAME COLUMN effectivefrom TO effectivefrom;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacedimensionprices' AND column_name = 'effectiveto')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacedimensionprices' AND column_name = 'effectiveto') THEN
        ALTER TABLE marketplacedimensionprices RENAME COLUMN effectiveto TO effectiveto;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacedimensionprices' AND column_name = 'isactive')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacedimensionprices' AND column_name = 'isactive') THEN
        ALTER TABLE marketplacedimensionprices RENAME COLUMN isactive TO isactive;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacedimensionprices' AND column_name = 'updatedat')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplacedimensionprices' AND column_name = 'updatedat') THEN
        ALTER TABLE marketplacedimensionprices RENAME COLUMN updatedat TO updatedat;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idxmarketplacestockdimension
    ON marketplacestockbalances (holdingcode, itemcode, dimensionkey);

CREATE INDEX IF NOT EXISTS idxmarketplacepricedimension
    ON marketplacedimensionprices (holdingcode, itemcode, dimensionkey, marketplace, pricelevel);
