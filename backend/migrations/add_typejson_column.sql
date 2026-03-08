-- Add typejson column to result table
-- typejson: 0 = normal data row, 1 = summary/subtotal row
ALTER TABLE result ADD COLUMN IF NOT EXISTS typejson INTEGER DEFAULT 0;

-- Create index on typejson for faster filtering
CREATE INDEX IF NOT EXISTS idx_result_typejson ON result(typejson);

-- Create composite index for common query patterns
CREATE INDEX IF NOT EXISTS idx_result_guid_querynumber_typejson ON result(guid, querynumber, typejson);
