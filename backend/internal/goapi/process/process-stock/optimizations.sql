-- สร้าง Index เพื่อเพิ่มความเร็วในการ query
CREATE INDEX CONCURRENTLY IF NOT EXISTS idxdocdetailitemcodetransflagdocdatetime
ON docdetail(itemcode, transflag, docdatetime, linenumber);

-- เพิ่ม partial index สำหรับ transflag ที่ใช้บ่อย
CREATE INDEX CONCURRENTLY IF NOT EXISTS idxdocdetailtransflagactive
ON docdetail(transflag, itemcode) 
WHERE transflag IN (44, 56, 72, 20, 21, 62, 12, 310, 60, 61, 54, 66, 48, 16, 866, 868);

-- ปรับ PostgreSQL configuration
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';  
ALTER SYSTEM SET maintenance_work_mem = '64MB';
SELECT pg_reload_conf();
