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
    holdingcode VARCHAR(100) NOT NULL,
    docno VARCHAR(100) NOT NULL,
    transflag VARCHAR(10) NOT NULL,
    retrycount INTEGER DEFAULT 0,
    createdat TIMESTAMP DEFAULT NOW(),
    updatedat TIMESTAMP DEFAULT NOW(),
    status VARCHAR(20) DEFAULT 'pending',
    processedat TIMESTAMP NULL,
    errormessage TEXT NULL,

    INDEX idxqueuesshopstatus (holdingcode, status),
    INDEX idxqueuespop (holdingcode, createdat) WHERE status = 'pending'
);
```

### Table: deadletterqueue
```sql
CREATE TABLE deadletterqueue (
    id BIGSERIAL PRIMARY KEY,
    holdingcode VARCHAR(100) NOT NULL,
    docno VARCHAR(100) NOT NULL,
    transflag VARCHAR(10) NOT NULL,
    retrycount INTEGER DEFAULT 0,
    createdat TIMESTAMP NOT NULL,
    failedat TIMESTAMP DEFAULT NOW(),
    errormessage TEXT NOT NULL,

    INDEX idxdlqholdingcode (holdingcode),
    INDEX idxdlqfailedat (failedat)
);
```

### Table: distributedlocks
```sql
CREATE TABLE distributedlocks (
    lockkey VARCHAR(500) PRIMARY KEY,
    owner VARCHAR(100) NOT NULL,
    acquiredat TIMESTAMP DEFAULT NOW(),
    expiresat TIMESTAMP NOT NULL,

    INDEX idxlocksexpiresat (expiresat)
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
- สร้าง index บน `(holdingcode, status)` สำหรับ PopFromQueue
- สร้าง partial index `WHERE status = 'pending'` สำหรับเพิ่มประสิทธิภาพ

### Connection Pooling
- ใช้ connection pooling (max 100 connections)
- ตั้งค่า `idle_in_transaction_session_timeout` เพื่อป้องกัน idle transactions

### Cleanup
- ตั้งค่า cron job ทำความสะอาด completed items เก่า:
```sql
DELETE FROM queues
WHERE status = 'completed'
  AND updatedat < NOW() - INTERVAL '7 days';
```

- ทำความสะอาด expired locks:
```sql
DELETE FROM distributedlocks
WHERE expiresat < NOW();
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
SELECT holdingcode, COUNT(*) as pendingcount
FROM queues
WHERE status = 'pending'
GROUP BY holdingcode
ORDER BY pendingcount DESC
LIMIT 10;

-- งานที่ล้มเหลว
SELECT holdingcode, docno, errormessage, failedat
FROM deadletterqueue
ORDER BY failedat DESC
LIMIT 20;
```

### Query Active Locks
```sql
-- Locks ที่กำลัง active
SELECT lockkey, owner, acquiredat, expiresat,
       EXTRACT(EPOCH FROM (expiresat - NOW())) as ttl_seconds
FROM distributedlocks
WHERE expiresat > NOW()
ORDER BY acquiredat DESC;
```

## Troubleshooting

### ปัญหา: งาน stuck ใน processing
```sql
-- หางานที่ processing นานเกินไป (เกิน 1 ชม.)
SELECT id, holdingcode, docno, processedat
FROM queues
WHERE status = 'processing'
  AND processedat < NOW() - INTERVAL '1 hour';

-- Reset กลับเป็น pending
UPDATE queues
SET status = 'pending',
    processedat = NULL
WHERE status = 'processing'
  AND processedat < NOW() - INTERVAL '1 hour';
```

### ปัญหา: Lock ไม่ถูกปล่อย
```sql
-- ลบ locks ที่หมดอายุ
DELETE FROM distributedlocks
WHERE expiresat < NOW();

-- Force release lock (ใช้เฉพาะกรณีฉุกเฉิน)
DELETE FROM distributedlocks
WHERE lockkey = 'stock:calc:SHOP001:ITEM001';
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
