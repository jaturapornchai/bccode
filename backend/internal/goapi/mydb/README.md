# mydb - ตัวจัดการการเชื่อมต่อ PostgreSQL ของ goapi

ระบบใช้ PostgreSQL เป็นฐานข้อมูลเดียว แพ็กเกจนี้ดูแล connection pool ต่อ holding พร้อม circuit breaker และ performance log

## ไฟล์

- `manager_pool.go` - `InitManagerPool(DatabaseManagerConfig)` / `GetGlobalManagerPool().GetManager(holdingCode)` / `CloseAllManagers()`
- `unified_database_manager.go` - `DatabaseManager`: `GetPostgreSQLConnection`, `QueryPostgreSQL`, `ExecPostgreSQL`, `BatchInsertPostgreSQL`, `PingAll`, `GetStats`
- `circuit_breaker.go` - circuit breaker + `ExecuteWithRetry`
- `performance_logger.go` - log เวลาที่ใช้ของแต่ละ query
- `connection_pool_manager.go` - pool ต่อชื่อฐานข้อมูล + maintenance
- `brand_db.go` - connection เดี่ยวของ brand mappings

## ใช้งาน

```go
mydb.InitManagerPool(mydb.DatabaseManagerConfig{
    PostgreSQLHost:     os.Getenv("POSTGRES_HOST"),
    PostgreSQLPort:     os.Getenv("POSTGRES_PORT"),
    PostgreSQLUser:     os.Getenv("POSTGRES_USER"),
    PostgreSQLPassword: os.Getenv("POSTGRES_PASSWORD"),
    PostgreSQLSSLMode:  os.Getenv("POSTGRES_SSL_MODE"),
})

manager, err := mydb.GetGlobalManagerPool().GetManager(holdingCode)
rows, err := manager.QueryPostgreSQL(ctx, "SELECT code, name_1 FROM ic_inventory WHERE code = $1", itemCode)
```

ใช้ parameterized query (`$1`, `$2`) เสมอ ห้ามต่อ string เป็น SQL
