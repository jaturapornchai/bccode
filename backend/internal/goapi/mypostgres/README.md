# PostgreSQL Queue System

Queue ของ goapi เก็บใน PostgreSQL (ฐานข้อมูลเดียวของระบบ)

## คุณสมบัติ

### 1. Queue Management
- **FIFO Queue**: จัดการ queue งาน document ตาม shop
- **Dead Letter Queue**: เก็บงานที่ล้มเหลวเกิน 3 ครั้ง
- **Retry Mechanism**: retry อัตโนม้าติ สูงสุด 3 ครั้ง
- **Status Tracking**: tracking สถานะงาน (pending, processing, completed, failed)

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

## การติดตั้ง

### 1. สร้างตารางในฐานข้อมูล

```bash
psql -h localhost -U postgres -d your_database -f mypostgres/schema.sql
```

### 2. ตั้งค่า environment variables

ใช้ค่าเชื่อมต่อ PostgreSQL ที่มีอยู่:

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
