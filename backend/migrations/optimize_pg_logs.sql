-- =====================================================
-- PostgreSQL Log & WAL Optimization
-- Run this on each PostgreSQL database
-- =====================================================

-- 1. Check current WAL settings
SELECT name, setting, unit, context
FROM pg_settings
WHERE name IN ('max_wal_size', 'min_wal_size', 'wal_keep_size', 'archive_mode', 'log_statement');

-- 2. Reduce WAL retention (requires superuser and pg_reload_conf())
-- ALTER SYSTEM SET max_wal_size = '1GB';        -- Default 1GB, reduce if disk is limited
-- ALTER SYSTEM SET min_wal_size = '80MB';       -- Minimum WAL to keep
-- ALTER SYSTEM SET wal_keep_size = '0';         -- Don't keep extra WAL for standby
-- ALTER SYSTEM SET archive_mode = 'off';        -- Disable WAL archiving if not using replication
-- SELECT pg_reload_conf();

-- 3. Reduce logging verbosity
-- ALTER SYSTEM SET log_statement = 'none';              -- Don't log all statements
-- ALTER SYSTEM SET log_min_duration_statement = 1000;   -- Only log queries > 1 second
-- ALTER SYSTEM SET log_connections = 'off';
-- ALTER SYSTEM SET log_disconnections = 'off';
-- ALTER SYSTEM SET log_lock_waits = 'off';
-- SELECT pg_reload_conf();

-- 4. Manually force WAL checkpoint to reclaim space
CHECKPOINT;

-- 5. Check WAL file count and size
SELECT
    pg_size_pretty(sum(size)) as total_wal_size,
    count(*) as wal_file_count
FROM pg_ls_waldir();

-- 6. Vacuum to reclaim space from dead tuples
-- VACUUM (VERBOSE, ANALYZE);

-- 7. Check table bloat (optional)
SELECT
    schemaname,
    relname as table_name,
    pg_size_pretty(pg_total_relation_size(schemaname || '.' || relname)) as total_size,
    n_dead_tup as dead_tuples,
    n_live_tup as live_tuples,
    round(100.0 * n_dead_tup / nullif(n_live_tup + n_dead_tup, 0), 2) as dead_percentage
FROM pg_stat_user_tables
WHERE n_dead_tup > 1000
ORDER BY n_dead_tup DESC
LIMIT 20;
