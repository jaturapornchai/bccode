-- Migration: Add stockcalculationstate table for incremental calculation
-- Date: 2024
-- Purpose: Track checksums for item codes to enable incremental stock calculation

-- Create table for storing calculation state
CREATE TABLE IF NOT EXISTS stockcalculationstate (
    holdingcode VARCHAR(100) NOT NULL,
    itemcode VARCHAR(100) NOT NULL,
    lastchecksum CHAR(32),
    lastcalctime TIMESTAMPTZ DEFAULT NOW(),
    version INTEGER DEFAULT 0,
    createdat TIMESTAMPTZ DEFAULT NOW(),
    updatedat TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (holdingcode, itemcode)
);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idxstockcalcstateshopitem
ON stockcalculationstate(holdingcode, itemcode);

CREATE INDEX IF NOT EXISTS idxstockcalcstatelastcalctime
ON stockcalculationstate(lastcalctime);

-- Add input_checksum and version columns to processstockcost (optional, for future use)
-- ALTER TABLE processstockcost ADD COLUMN IF NOT EXISTS input_checksum CHAR(32);
-- ALTER TABLE processstockcost ADD COLUMN IF NOT EXISTS version INT DEFAULT 0;

-- Comment
COMMENT ON TABLE stockcalculationstate IS 'Tracks item checksums for incremental stock cost calculation';
COMMENT ON COLUMN stockcalculationstate.holdingcode IS 'Holding codeentifier';
COMMENT ON COLUMN stockcalculationstate.itemcode IS 'Item code';
COMMENT ON COLUMN stockcalculationstate.lastchecksum IS 'MD5 checksum of item data at last calculation';
COMMENT ON COLUMN stockcalculationstate.lastcalctime IS 'Timestamp of last calculation';
COMMENT ON COLUMN stockcalculationstate.version IS 'Version counter for optimistic locking';
