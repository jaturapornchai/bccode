-- Accounting stock remains product/warehouse/location based.
-- Marketplace availability and dimension-level selling prices are separate projections.

ALTER TABLE inventory_stock_balances
    ADD COLUMN IF NOT EXISTS reservedqty NUMERIC(18,4) NOT NULL DEFAULT 0;

ALTER TABLE product_costing_config
    ADD COLUMN IF NOT EXISTS cost_by_warehouse BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE inventory_stock_balances
    DROP CONSTRAINT IF EXISTS inventory_stock_balances_scope_dimension_unique,
    DROP CONSTRAINT IF EXISTS inventory_stock_balances_holding_code_itemcode_whcode_locationcode_key;

ALTER TABLE inventory_stock_balances
    ADD CONSTRAINT inventory_stock_balances_holding_code_itemcode_whcode_locationcode_key
    UNIQUE (holding_code, itemcode, whcode, locationcode);

CREATE TABLE IF NOT EXISTS marketplace_stock_balances (
    id BIGSERIAL PRIMARY KEY,
    holding_code TEXT NOT NULL,
    item_code TEXT NOT NULL,
    dimension_key TEXT NOT NULL DEFAULT '',
    dimension_values JSONB NOT NULL DEFAULT '{}'::jsonb,
    current_qty NUMERIC(18,4) NOT NULL DEFAULT 0,
    reserved_qty NUMERIC(18,4) NOT NULL DEFAULT 0,
    available_qty NUMERIC(18,4) NOT NULL DEFAULT 0,
    source TEXT NOT NULL DEFAULT 'accounting',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (holding_code, item_code, dimension_key)
);

CREATE TABLE IF NOT EXISTS marketplace_dimension_prices (
    id BIGSERIAL PRIMARY KEY,
    holding_code TEXT NOT NULL,
    item_code TEXT NOT NULL,
    barcode TEXT NOT NULL DEFAULT '',
    dimension_key TEXT NOT NULL DEFAULT '',
    dimension_values JSONB NOT NULL DEFAULT '{}'::jsonb,
    price_level TEXT NOT NULL DEFAULT '',
    marketplace TEXT NOT NULL DEFAULT '',
    currency VARCHAR(3) NOT NULL DEFAULT 'THB',
    price NUMERIC(18,4) NOT NULL DEFAULT 0,
    sale_price NUMERIC(18,4) NOT NULL DEFAULT 0,
    compare_at_price NUMERIC(18,4) NOT NULL DEFAULT 0,
    effective_from TIMESTAMPTZ,
    effective_to TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT true,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (holding_code, item_code, barcode, dimension_key, price_level, marketplace, currency)
);

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_stock_balances' AND column_name = 'itemcode')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_stock_balances' AND column_name = 'item_code') THEN
        ALTER TABLE marketplace_stock_balances RENAME COLUMN itemcode TO item_code;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_stock_balances' AND column_name = 'dimensionkey')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_stock_balances' AND column_name = 'dimension_key') THEN
        ALTER TABLE marketplace_stock_balances RENAME COLUMN dimensionkey TO dimension_key;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_stock_balances' AND column_name = 'dimensionvalues')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_stock_balances' AND column_name = 'dimension_values') THEN
        ALTER TABLE marketplace_stock_balances RENAME COLUMN dimensionvalues TO dimension_values;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_stock_balances' AND column_name = 'currentqty')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_stock_balances' AND column_name = 'current_qty') THEN
        ALTER TABLE marketplace_stock_balances RENAME COLUMN currentqty TO current_qty;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_stock_balances' AND column_name = 'reservedqty')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_stock_balances' AND column_name = 'reserved_qty') THEN
        ALTER TABLE marketplace_stock_balances RENAME COLUMN reservedqty TO reserved_qty;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_stock_balances' AND column_name = 'availableqty')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_stock_balances' AND column_name = 'available_qty') THEN
        ALTER TABLE marketplace_stock_balances RENAME COLUMN availableqty TO available_qty;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_stock_balances' AND column_name = 'updatedat')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_stock_balances' AND column_name = 'updated_at') THEN
        ALTER TABLE marketplace_stock_balances RENAME COLUMN updatedat TO updated_at;
    END IF;
END $$;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_dimension_prices' AND column_name = 'itemcode')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_dimension_prices' AND column_name = 'item_code') THEN
        ALTER TABLE marketplace_dimension_prices RENAME COLUMN itemcode TO item_code;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_dimension_prices' AND column_name = 'dimensionkey')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_dimension_prices' AND column_name = 'dimension_key') THEN
        ALTER TABLE marketplace_dimension_prices RENAME COLUMN dimensionkey TO dimension_key;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_dimension_prices' AND column_name = 'dimensionvalues')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_dimension_prices' AND column_name = 'dimension_values') THEN
        ALTER TABLE marketplace_dimension_prices RENAME COLUMN dimensionvalues TO dimension_values;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_dimension_prices' AND column_name = 'pricelevel')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_dimension_prices' AND column_name = 'price_level') THEN
        ALTER TABLE marketplace_dimension_prices RENAME COLUMN pricelevel TO price_level;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_dimension_prices' AND column_name = 'saleprice')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_dimension_prices' AND column_name = 'sale_price') THEN
        ALTER TABLE marketplace_dimension_prices RENAME COLUMN saleprice TO sale_price;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_dimension_prices' AND column_name = 'compareatprice')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_dimension_prices' AND column_name = 'compare_at_price') THEN
        ALTER TABLE marketplace_dimension_prices RENAME COLUMN compareatprice TO compare_at_price;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_dimension_prices' AND column_name = 'effectivefrom')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_dimension_prices' AND column_name = 'effective_from') THEN
        ALTER TABLE marketplace_dimension_prices RENAME COLUMN effectivefrom TO effective_from;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_dimension_prices' AND column_name = 'effectiveto')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_dimension_prices' AND column_name = 'effective_to') THEN
        ALTER TABLE marketplace_dimension_prices RENAME COLUMN effectiveto TO effective_to;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_dimension_prices' AND column_name = 'isactive')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_dimension_prices' AND column_name = 'is_active') THEN
        ALTER TABLE marketplace_dimension_prices RENAME COLUMN isactive TO is_active;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_dimension_prices' AND column_name = 'updatedat')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'marketplace_dimension_prices' AND column_name = 'updated_at') THEN
        ALTER TABLE marketplace_dimension_prices RENAME COLUMN updatedat TO updated_at;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_marketplace_stock_dimension
    ON marketplace_stock_balances (holding_code, item_code, dimension_key);

CREATE INDEX IF NOT EXISTS idx_marketplace_price_dimension
    ON marketplace_dimension_prices (holding_code, item_code, dimension_key, marketplace, price_level);
