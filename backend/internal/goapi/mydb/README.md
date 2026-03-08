# Unified Database Connection System - ระบบเชื่อมต่อฐานข้อมูลแบบรวมศูนย์

## ภาพรวม

ระบบนี้ได้รับการออกแบบมาเพื่อสร้างมาตรฐานเดียวกันสำหรับการเชื่อมต่อและจัดการฐานข้อมูล PostgreSQL และ ClickHouse ในโปรเจกต์

## ✨ อัปเดทล่าสุด

### เวอร์ชัน 2.0 - Enhanced Edition

**เพิ่มเติม:**
- 🛡️ **Circuit Breaker Pattern** - ป้องกัน cascade failures
- 🔄 **Retry Logic** พร้อม exponential backoff
- ⚡ **Connection Pool Tuning** สำหรับ high traffic (100 max open conns)
- 📊 **Performance Monitoring** ที่ดีขึ้น
- 🎯 **Fine-tuned Timeouts** และ connection limits

## โครงสร้างไฟล์

### 1. `mydb/circuit_breaker.go` 🆕
- **Circuit Breaker Pattern** implementation
- ป้องกัน cascade failures เมื่อ database มีปัญหา
- **States:**
  - `CLOSED` - ปกติ ทำงานได้
  - `OPEN` - เปิดวงจร หยุดส่ง request (หลัง fail 5 ครั้ง)
  - `HALF_OPEN` - ลองส่ง request บางส่วนเพื่อทดสอบ
- **Auto-recovery** - รอ 10 วินาทีแล้วลองใหม่
- **Retry Logic** พร้อม exponential backoff (100ms → 5s max)

### 2. `mydb/performance_logger.go`
- ระบบ logging performance ที่ครบถ้วน
- วัดเวลาตอบสนองของ query แต่ละตัว
- จัดระดับการแสดงผลตาม performance
- รวมระบบ stats collector สำหรับการวิเคราะห์

**Features:**
- `QueryTimer` และ `BatchQueryTimer` สำหรับจับเวลา
- `DatabasePerformanceLogger` สำหรับบันทึก metrics
- `QueryStatsCollector` สำหรับรวบรวมสถิติ
- การจำแนก slow queries (เกิน 5 วินาที)

### 3. `mydb/unified_database_manager.go` ⚡ Enhanced
- Database manager แบบ unified ที่ใช้มาตรฐานเดียวกัน
- **Circuit Breaker** ในทุก operation
- **Retry Logic** อัตโนมัติเมื่อเชื่อมต่อ
- Connection pooling สำหรับทั้ง PostgreSQL และ ClickHouse
- Performance logging อัตโนมัติในทุก operation

**Features:**
- **PostgreSQL Connection Pool** (ลดลงครึ่งหนึ่งเพื่อไม่กระทบระบบอื่น):
  - MaxOpenConns: **50** ⬇️ (ลดลงครึ่งหนึ่ง)
  - MaxIdleConns: **15** ⬇️ (ลดลงครึ่งหนึ่ง)
  - ConnMaxLifetime: **1 ชั่วโมง** ⬇️ (ลดจาก 2h)
  - ConnMaxIdleTime: **15 นาที** ⬇️ (ลดจาก 30m)

- **ClickHouse Connection:**
  - ใช้ existing connection pool จาก myclickhouse
  - เชื่อมต่อผ่าน `ClickHouseFastConnect()`
  - มี health check และ auto-reconnection

- **Unified Operations:**
  - `QueryPostgreSQL()` - Query PostgreSQL พร้อม performance logging
  - `QueryClickHouse()` - Query ClickHouse พร้อม performance logging  
  - `ExecPostgreSQL()` - Execute บน PostgreSQL พร้อม performance logging
  - `ExecClickHouse()` - Execute บน ClickHouse พร้อม performance logging
  - `BatchInsertPostgreSQL()` - Batch insert บน PostgreSQL
  - `PingAll()` - ทดสอบการเชื่อมต่อทั้งหมด

### 3. `examples/database/unified_database_example.go`
- ตัวอย่างการใช้งานระบบ unified database manager
- ทดสอบการเชื่อมต่อ, query, และ performance logging

## วิธีการใช้งาน

### 1. สร้าง Database Manager

```go
import "goapi/mydb"

config := mydb.DatabaseConfig{
    PostgreSQLHost:     "localhost",
    PostgreSQLPort:     "5432", 
    PostgreSQLUser:     "postgres",
    PostgreSQLPassword: "password",
    PostgreSQLDatabase: "test_db",
    PostgreSQLSSLMode:  "disable",
    
    ClickHouseHost:     "localhost",
    ClickHousePort:     "9000",
    ClickHouseUser:     "default", 
    ClickHousePassword: "",
    ClickHouseDatabase: "default",
}

dbManager, err := mydb.NewDatabaseManager(config)
if err != nil {
    log.Fatal(err)
}
defer dbManager.Close()
```

### 2. ใช้งาน PostgreSQL

```go
ctx := context.Background()

// Query
results, err := dbManager.QueryPostgreSQL(ctx, 
    "SELECT * FROM users WHERE age > $1", 18)

// Exec  
_, err = dbManager.ExecPostgreSQL(ctx, 
    "INSERT INTO users (name, age) VALUES ($1, $2)", "สมชาย", 25)

// Batch Insert
data := [][]any{
    {"สมชาย", 25},
    {"สมหญิง", 30}, 
}
err = dbManager.BatchInsertPostgreSQL(ctx, "users", 
    []string{"name", "age"}, data)
```

### 3. ใช้งาน ClickHouse

```go
// Query
chResults, err := dbManager.QueryClickHouse(ctx, 
    "SELECT * FROM events WHERE date = today()")

// Exec
err = dbManager.ExecClickHouse(ctx, 
    "INSERT INTO events (id, name, date) VALUES (1, 'event1', today())")
```

### 4. ตรวจสอบ Performance

```go
// สถิติ PostgreSQL
pgStats := dbManager.GetStats(mydb.PostgreSQL)

// สถิติ ClickHouse  
chStats := dbManager.GetStats(mydb.ClickHouse)

// ทดสอบการเชื่อมต่อ
pingResults := dbManager.PingAll()
```

## Performance Logging

ระบบจะแสดงผล performance ตามระดับ:

- **⚡ < 100ms:** Debug level
- **🔄 < 1s:** Info level  
- **🐌 < 5s:** Warning level
- **⏰ > 5s:** Error level (Slow Query Detection)

### Example Log Output:
```
[PERF] PostgreSQL.Query: SELECT * FROM users (duration: 45ms, rows: 100, status: success)
[PERF] ClickHouse.Query: SELECT * FROM events (duration: 234ms, rows: 50, status: success)  
🔍 SLOW QUERY DETECTED - Database: PostgreSQL, Operation: Query, Duration: 12.5s
Query: SELECT * FROM large_table WHERE condition...
```

## ข้อดีของระบบใหม่

### 1. **มาตรฐานเดียวกัน**
- ใช้รูปแบบ function เดียวกันทั้ง PostgreSQL และ ClickHouse
- Interface เดียวกันสำหรับ database operations
- Error handling แบบเดียวกัน

### 2. **Connection Pool ที่เหมาะสม**  
- PostgreSQL: 50 connections, 20 idle, 2h lifetime
- ClickHouse: ใช้ existing pool ที่ optimized แล้ว
- Auto-reconnection และ health monitoring

### 3. **Performance Logging**
- บันทึกเวลาตอบสนองของทุก query
- ตรวจจับ slow queries อัตโนมัติ
- สถิติ performance ที่ครบถ้วน

### 4. **การจัดการแบบ Unified**
- เปิด-ปิด connection ทั้งหมดด้วย function เดียว
- ตรวจสอบสถานะ connection ทั้งหมด
- Error handling แบบ centralized

## การย้ายจากระบบเดิม

### PostgreSQL (จาก `mypg`)
```go
// เดิม
db, err := mypg.Connect("database")
results, err := mypg.QuerySelectAll(db, "SELECT * FROM table")

// ใหม่  
results, err := dbManager.QueryPostgreSQL(ctx, "SELECT * FROM table")
```

### ClickHouse (จาก `myclickhouse`)  
```go
// เดิม
conn, err := myclickhouse.ClickHouseFastConnect()
results, err := myclickhouse.QuerySelectAll(conn, "SELECT * FROM table")

// ใหม่
results, err := dbManager.QueryClickHouse(ctx, "SELECT * FROM table")
```

## ข้อแนะนำ

1. **ใช้ context** สำหรับทุก database operation เพื่อ timeout control
2. **Monitor performance logs** เพื่อหา slow queries  
3. **ใช้ Batch operations** สำหรับการ insert ข้อมูลจำนวนมาก
4. **ปิด Database Manager** เมื่อไม่ใช้งานแล้ว

## TODO

- [ ] ทดสอบการทำงานจริงกับ databases
- [ ] ปรับแต่ง performance logging format
- [ ] เพิ่ม metrics export สำหรับ monitoring tools
- [ ] เพิ่ม connection retry logic ที่ซับซ้อนขึ้น