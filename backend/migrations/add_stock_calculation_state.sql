-- Migration: Add stock_calculation_state table for incremental calculation
-- Date: 2024
-- Purpose: Track checksums for item codes to enable incremental stock calculation

-- Create table for storing calculation state
CREATE TABLE IF NOT EXISTS stock_calculation_state (
    shop_id VARCHAR(100) NOT NULL,
    item_code VARCHAR(100) NOT NULL,
    last_checksum CHAR(32),
    last_calc_time TIMESTAMPTZ DEFAULT NOW(),
    version INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (shop_id, item_code)
);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_stock_calc_state_shop_item
ON stock_calculation_state(shop_id, item_code);

CREATE INDEX IF NOT EXISTS idx_stock_calc_state_last_calc_time
ON stock_calculation_state(last_calc_time);

-- Add input_checksum and version columns to processstockcost (optional, for future use)
-- ALTER TABLE processstockcost ADD COLUMN IF NOT EXISTS input_checksum CHAR(32);
-- ALTER TABLE processstockcost ADD COLUMN IF NOT EXISTS version INT DEFAULT 0;

-- Comment
COMMENT ON TABLE stock_calculation_state IS 'Tracks item checksums for incremental stock cost calculation';
COMMENT ON COLUMN stock_calculation_state.shop_id IS 'Shop identifier';
COMMENT ON COLUMN stock_calculation_state.item_code IS 'Item code';
COMMENT ON COLUMN stock_calculation_state.last_checksum IS 'MD5 checksum of item data at last calculation';
COMMENT ON COLUMN stock_calculation_state.last_calc_time IS 'Timestamp of last calculation';
COMMENT ON COLUMN stock_calculation_state.version IS 'Version counter for optimistic locking';
