-- =====================================================
-- ClickHouse System Log Optimization
-- Run this on ClickHouse to prevent log accumulation
-- =====================================================

-- 1. Check current system log sizes
SELECT
    database,
    table,
    formatReadableSize(sum(bytes_on_disk)) as size,
    sum(rows) as total_rows
FROM system.parts
WHERE database = 'system' AND table LIKE '%log%'
GROUP BY database, table
ORDER BY sum(bytes_on_disk) DESC;

-- 2. Set TTL on query_log (keep only 7 days)
ALTER TABLE system.query_log MODIFY TTL event_date + INTERVAL 7 DAY;

-- 3. Set TTL on query_thread_log (keep only 3 days)
ALTER TABLE system.query_thread_log MODIFY TTL event_date + INTERVAL 3 DAY;

-- 4. Set TTL on part_log (keep only 7 days)
ALTER TABLE system.part_log MODIFY TTL event_date + INTERVAL 7 DAY;

-- 5. Set TTL on trace_log (keep only 1 day - very verbose)
ALTER TABLE system.trace_log MODIFY TTL event_date + INTERVAL 1 DAY;

-- 6. Set TTL on metric_log (keep only 3 days)
ALTER TABLE system.metric_log MODIFY TTL event_date + INTERVAL 3 DAY;

-- 7. Set TTL on asynchronous_metric_log (keep only 1 day)
ALTER TABLE system.asynchronous_metric_log MODIFY TTL event_date + INTERVAL 1 DAY;

-- 8. Manually truncate old logs (immediate cleanup)
-- WARNING: This will delete data immediately!
TRUNCATE TABLE system.query_log;
TRUNCATE TABLE system.query_thread_log;
TRUNCATE TABLE system.trace_log;

-- 9. Force merge to reclaim disk space
OPTIMIZE TABLE system.query_log FINAL;
OPTIMIZE TABLE system.query_thread_log FINAL;
OPTIMIZE TABLE system.part_log FINAL;

-- 10. Check sizes after cleanup
SELECT
    database,
    table,
    formatReadableSize(sum(bytes_on_disk)) as size,
    sum(rows) as total_rows
FROM system.parts
WHERE database = 'system' AND table LIKE '%log%'
GROUP BY database, table
ORDER BY sum(bytes_on_disk) DESC;

-- =====================================================
-- Alternative: Disable logging entirely (config.xml)
-- Add to /etc/clickhouse-server/config.d/disable_logs.xml:
-- =====================================================
-- <clickhouse>
--     <query_log remove="1"/>
--     <query_thread_log remove="1"/>
--     <part_log remove="1"/>
--     <trace_log remove="1"/>
--     <metric_log remove="1"/>
--     <asynchronous_metric_log remove="1"/>
-- </clickhouse>
