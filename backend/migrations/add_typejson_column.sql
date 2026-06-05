-- Add typejson column to result table
-- typejson: 0 = normal data row, 1 = summary/subtotal row
ALTER TABLE result ADD COLUMN IF NOT EXISTS typejson INTEGER DEFAULT 0;

-- Create index on typejson for faster filtering
CREATE INDEX IF NOT EXISTS idxresulttypejson ON result(typejson);

-- Create composite index for common query patterns
CREATE INDEX IF NOT EXISTS idxresultguidquerynumbertypejson ON result(guid, querynumber, typejson);
