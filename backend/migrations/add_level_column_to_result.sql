-- Adds the level column to the result table so each record stores its hierarchy depth.
ALTER TABLE IF EXISTS result
    ADD COLUMN IF NOT EXISTS level INTEGER DEFAULT 0;

COMMENT ON COLUMN result.level IS 'ระดับของข้อมูล (0=data, 1=summary level 1, 2=summary level 2, ..., 99=grand total)';

-- Ensure existing rows have a non-null value.
UPDATE result SET level = COALESCE(level, 0);
