# PostgreSQL Queue & Lock System

แทนที่ระบบ Redis ด้วย PostgreSQL สำหรับ Queue Management และ Distributed Locking

## คุณสมบัติ

### 1. Queue Management
- **FIFO Queue**: จัดการ queue งาน document ตาม shop
- **Dead Letter Queue**: เก็บงานที่ล้มเหลวเกิน 3 ครั้ง
- **Retry Mechanism**: retry อัตโนม้าติ สูงสุด 3 ครั้ง
- **Status Tracking**: tracking สถานะงาน (pending, processing, completed, failed)

### 2. Distributed Locking
- **PostgreSQL Advisory Locks**: ป้องกัน race condition
- **Auto-expiry**: ล็อคหมดอายุอัตโนมัติ
- **Retry with backoff**: พยายามขอ lock อีกครั้งถ้าไม่ได้

## โครงสร้างตาราง

### Table: queues
```sql
CREATE TABLE queues (
    id BIGSERIAL PRIMARY KEY,
    holding_code VARCHAR(100) NOT NULL,
    doc_no VARCHAR(100) NOT NULL,
    trans_flag VARCHAR(10) NOT NULL,
    retry_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    status VARCHAR(20) DEFAULT 'pending',
    processed_at TIMESTAMP NULL,
    error_message TEXT NULL,

    INDEX idx_queues_shop_status (holding_code, status),
    INDEX idx_queues_pop (holding_code, created_at) WHERE status = 'pending'
);
```

### Table: dead_letter_queue
```sql
CREATE TABLE dead_letter_queue (
    id BIGSERIAL PRIMARY KEY,
    holding_code VARCHAR(100) NOT NULL,
    doc_no VARCHAR(100) NOT NULL,
    trans_flag VARCHAR(10) NOT NULL,
    retry_count INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    failed_at TIMESTAMP DEFAULT NOW(),
    error_message TEXT NOT NULL,

    INDEX idx_dlq_holding_code (holding_code),
    INDEX idx_dlq_failed_at (failed_at)
);
```

### Table: distributed_locks
```sql
CREATE TABLE distributed_locks (
    lock_key VARCHAR(500) PRIMARY KEY,
    owner VARCHAR(100) NOT NULL,
    acquired_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,

    INDEX idx_locks_expires_at (expires_at)
);
```

## การติดตั้ง

### 1. สร้างตารางในฐานข้อมูล

```bash
psql -h localhost -U postgres -d your_database -f mypostgres/schema.sql
```

### 2. ตั้งค่า environment variables

ไม่ต้องมี Redis config แล้ว ใช้ PostgreSQL ที่มีอยู่:

```bash
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your_password
POSTGRES_DB=your_database
```

### 3. อัปเดต main.go

```go
import (
    "goapi/myglobal"
    "goapi/mydb"
)

// สร้าง database manager
dbManager, err := mydb.NewDatabaseManager(dbConfig)
if err != nil {
    log.Fatal(err)
}

// ตั้งค่า global database connection
db, _ := dbManager.GetPostgreSQLConnection()
myglobal.SetGlobalDatabaseConnection(db)
```

## การใช้งาน

### Queue Management

```go
import "goapi/mypostgres"

// เพิ่มงานเข้า queue
qm := mypostgres.NewQueueManager(db)
item := mypostgres.QueueItem{
    HoldingCode:    "SHOP001",
    DocNo:     "PO-2024-001",
    TransFlag: "6",
    CreatedAt: time.Now(),
}
err := qm.AddToQueue(ctx, item)

// ดึงงานจาก queue
item, err := qm.PopFromQueue(ctx, "SHOP001")

// ดึงรายชื่อ shop ที่มีงาน
shops, err := qm.GetActiveShops(ctx)

// ดึงสถิติ queue
stats, err := qm.GetQueueStats(ctx, "SHOP001")
summary, err := qm.GetQueueSummary(ctx)
```

### Distributed Locking

```go
import "goapi/mypostgres"

// สร้าง lock
lock := mypostgres.NewDistributedLock(db, "stock:calc:SHOP001:ITEM001", 5*time.Minute)

// ขอ lock พร้อม retry
err := lock.AcquireWithRetry(ctx, 10, 500*time.Millisecond)
if err != nil {
    return err
}
defer lock.Release(ctx)

// ทำงานที่ต้อง critical section
// ...
```

## ข้อดีของ PostgreSQL เทียบกับ Redis

### 1. Durability
- ✅ **PostgreSQL**: ข้อมูล persist บน disk, ไม่สูญหายเมื่อ restart
- ❌ **Redis**: ข้อมูลอยู่ใน memory, อาจสูญหายถ้าไม่ได้ persist

### 2. ACID Compliance
- ✅ **PostgreSQL**: Transaction support เต็มรูปแบบ
- ⚠️ **Redis**: Transaction จำกัด

### 3. Complex Queries
- ✅ **PostgreSQL**: รองรับ SQL ซับซ้อน, JOIN, aggregation
- ❌ **Redis**: Query จำกัด, ไม่มี JOIN

### 4. Cost
- ✅ **PostgreSQL**: ใช้ infrastructure ที่มีอยู่
- ❌ **Redis**: ต้องมี Redis server แยก

### 5. Monitoring
- ✅ **PostgreSQL**: เครื่องมือ monitoring มากมาย
- ⚠️ **Redis**: เครื่องมือน้อยกว่า

## Performance Considerations

### Indexing
- สร้าง index บน `(holding_code, status)` สำหรับ PopFromQueue
- สร้าง partial index `WHERE status = 'pending'` สำหรับเพิ่มประสิทธิภาพ

### Connection Pooling
- ใช้ connection pooling (max 100 connections)
- ตั้งค่า `idle_in_transaction_session_timeout` เพื่อป้องกัน idle transactions

### Cleanup
- ตั้งค่า cron job ทำความสะอาด completed items เก่า:
```sql
DELETE FROM queues
WHERE status = 'completed'
  AND updated_at < NOW() - INTERVAL '7 days';
```

- ทำความสะอาด expired locks:
```sql
DELETE FROM distributed_locks
WHERE expires_at < NOW();
```

## Migration จาก Redis

### ข้อมูลใน Redis (ไม่ต้อง migrate)
- Queue items ใน Redis เป็น temporary data
- ไม่จำเป็นต้อง migrate เพราะจะ process ใหม่อัตโนมัติ

### ขั้นตอน Migration

1. **Deploy โค้ดใหม่**:
   - ปิด workers ชั่วคราว
   - Deploy โค้ดที่ใช้ PostgreSQL
   - สร้างตารางด้วย schema.sql

2. **Verify**:
   - ตรวจสอบว่าตารางถูกสร้างแล้ว
   - Test เพิ่ม/ดึงงานจาก queue

3. **Enable Workers**:
   - เปิด workers ใหม่
   - Monitor logs

4. **ลบ Redis** (หลังจากมั่นใจว่าทำงานดี):
   - หยุด Redis service
   - ลบ Redis config จาก docker-compose.yml
   - ลบโฟลเดอร์ myredis/

## Monitoring & Debugging

### Query สถิติ Queue
```sql
-- จำนวนงานตามสถานะ
SELECT status, COUNT(*)
FROM queues
GROUP BY status;

-- Shop ที่มีงานรอมากที่สุด
SELECT holding_code, COUNT(*) as pending_count
FROM queues
WHERE status = 'pending'
GROUP BY holding_code
ORDER BY pending_count DESC
LIMIT 10;

-- งานที่ล้มเหลว
SELECT holding_code, doc_no, error_message, failed_at
FROM dead_letter_queue
ORDER BY failed_at DESC
LIMIT 20;
```

### Query Active Locks
```sql
-- Locks ที่กำลัง active
SELECT lock_key, owner, acquired_at, expires_at,
       EXTRACT(EPOCH FROM (expires_at - NOW())) as ttl_seconds
FROM distributed_locks
WHERE expires_at > NOW()
ORDER BY acquired_at DESC;
```

## Troubleshooting

### ปัญหา: งาน stuck ใน processing
```sql
-- หางานที่ processing นานเกินไป (เกิน 1 ชม.)
SELECT id, holding_code, doc_no, processed_at
FROM queues
WHERE status = 'processing'
  AND processed_at < NOW() - INTERVAL '1 hour';

-- Reset กลับเป็น pending
UPDATE queues
SET status = 'pending',
    processed_at = NULL
WHERE status = 'processing'
  AND processed_at < NOW() - INTERVAL '1 hour';
```

### ปัญหา: Lock ไม่ถูกปล่อย
```sql
-- ลบ locks ที่หมดอายุ
DELETE FROM distributed_locks
WHERE expires_at < NOW();

-- Force release lock (ใช้เฉพาะกรณีฉุกเฉิน)
DELETE FROM distributed_locks
WHERE lock_key = 'stock:calc:SHOP001:ITEM001';
```

## API Reference

### QueueManager Methods

- `AddToQueue(ctx, item)` - เพิ่มงานเข้า queue
- `PopFromQueue(ctx, holdingCode)` - ดึงงานจาก queue (FIFO)
- `RequeueItem(ctx, item)` - ใส่งานกลับเข้า queue
- `AddToDeadLetterQueue(ctx, item, error)` - ส่งงานไป DLQ
- `GetActiveShops(ctx)` - ดึงรายชื่อ shop ที่มีงาน
- `GetQueueLength(ctx, holdingCode)` - ดึงความยาว queue
- `GetQueueStats(ctx, holdingCode)` - ดึงสถิติ queue ของ shop
- `GetQueueSummary(ctx)` - ดึงสรุปข้อมูล queue ทั้งหมด
- `MarkAsCompleted(ctx, itemId)` - ทำเครื่องหมายงานเสร็จ
- `CleanupOldCompletedItems(ctx, duration)` - ทำความสะอาดงานเก่า

### DistributedLock Methods

- `Acquire(ctx)` - ขอ lock (non-blocking)
- `AcquireWithRetry(ctx, maxRetries, delay)` - ขอ lock พร้อม retry
- `Release(ctx)` - ปล่อย lock
- `Extend(ctx, duration)` - ขยายระยะเวลา lock
- `IsAcquired()` - ตรวจสอบว่า lock ถูกขอแล้วหรือยัง

### LockManager Methods

- `CleanupExpiredLocks(ctx)` - ทำความสะอาด expired locks
- `GetActiveLocks(ctx)` - ดึงรายการ lock ที่ active
- `GetLockCount(ctx)` - ดึงจำนวน lock
- `ForceReleaseLock(ctx, lockKey)` - บังคับปล่อย lock
- `GetLockInfo(ctx, lockKey)` - ดึงข้อมูล lock
- `StartCleanupRoutine(ctx, interval)` - เริ่ม cleanup routine
