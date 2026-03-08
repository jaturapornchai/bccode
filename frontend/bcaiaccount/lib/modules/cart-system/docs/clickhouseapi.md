# ClickHouse SELECT API - ตัวอย่างการใช้งาน

## 📊 Endpoints

1. **POST `/clickhouse/query`** - Custom SQL Query (ยืดหยุ่นที่สุด)
2. **POST `/clickhouse/select`** - Simplified SELECT (ง่ายต่อการใช้งาน)

---

## 1. Custom SQL Query

### Endpoint
```
POST /clickhouse/query
```

### Request Body
```json
{
  "query": "SELECT * FROM table_name WHERE condition LIMIT 10"
}
```

### ตัวอย่าง 1: Query ทั้งหมด
```powershell
$body = @{
    query = "SELECT * FROM sales LIMIT 10"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:9999/clickhouse/query" `
    -Method POST -Body $body -ContentType "application/json"
```

**Response:**
```json
{
  "status": "success",
  "code": 200,
  "count": 10,
  "data": [
    {
      "id": 1,
      "shop_id": "shop001",
      "amount": 1500.50,
      "date": "2025-01-15"
    },
    ...
  ]
}
```

### ตัวอย่าง 2: Query กับเงื่อนไข
```powershell
$body = @{
    query = "SELECT shop_id, SUM(amount) as total FROM sales WHERE date >= '2025-01-01' GROUP BY shop_id"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:9999/clickhouse/query" `
    -Method POST -Body $body -ContentType "application/json"
```

### ตัวอย่าง 3: Query ข้อมูลรายวัน
```powershell
$body = @{
    query = @"
SELECT 
    toDate(created_at) as date,
    shop_id,
    COUNT(*) as total_orders,
    SUM(total_amount) as total_sales
FROM orders
WHERE shop_id = 'shop001'
    AND created_at >= '2025-01-01'
    AND created_at < '2025-02-01'
GROUP BY date, shop_id
ORDER BY date DESC
"@
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:9999/clickhouse/query" `
    -Method POST -Body $body -ContentType "application/json"
```

### ตัวอย่าง 4: JOIN Tables
```powershell
$body = @{
    query = @"
SELECT 
    o.order_id,
    o.shop_id,
    o.total_amount,
    c.customer_name,
    c.email
FROM orders o
LEFT JOIN customers c ON o.customer_id = c.customer_id
WHERE o.shop_id = 'shop001'
LIMIT 20
"@
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:9999/clickhouse/query" `
    -Method POST -Body $body -ContentType "application/json"
```

---

## 2. Simplified SELECT

### Endpoint
```
POST /clickhouse/select
```

### Request Body
```json
{
  "table": "table_name",
  "columns": ["col1", "col2"],  // optional, empty = SELECT *
  "where": {                     // optional
    "shop_id": "shop001",
    "status": "active"
  },
  "order_by": "created_at DESC", // optional
  "limit": 10,                   // optional
  "offset": 0                    // optional
}
```

### ตัวอย่าง 1: SELECT * ทั้งหมด
```powershell
$body = @{
    table = "products"
    limit = 10
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:9999/clickhouse/select" `
    -Method POST -Body $body -ContentType "application/json"
```

**Response:**
```json
{
  "status": "success",
  "code": 200,
  "count": 10,
  "data": [...],
  "query": "SELECT * FROM products LIMIT 10"
}
```

### ตัวอย่าง 2: SELECT เฉพาะ columns
```powershell
$body = @{
    table = "orders"
    columns = @("order_id", "shop_id", "total_amount", "created_at")
    limit = 20
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:9999/clickhouse/select" `
    -Method POST -Body $body -ContentType "application/json"
```

### ตัวอย่าง 3: SELECT กับเงื่อนไข WHERE
```powershell
$body = @{
    table = "sales"
    columns = @("shop_id", "product_id", "quantity", "amount")
    where = @{
        shop_id = "shop001"
        status = "completed"
    }
    order_by = "created_at DESC"
    limit = 50
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:9999/clickhouse/select" `
    -Method POST -Body $body -ContentType "application/json"
```

**Generated Query:**
```sql
SELECT shop_id, product_id, quantity, amount 
FROM sales 
WHERE shop_id = 'shop001' AND status = 'completed' 
ORDER BY created_at DESC 
LIMIT 50
```

### ตัวอย่าง 4: Pagination
```powershell
# หน้าที่ 1
$body = @{
    table = "customers"
    columns = @("customer_id", "name", "email", "phone")
    where = @{
        shop_id = "shop001"
    }
    order_by = "created_at DESC"
    limit = 20
    offset = 0
} | ConvertTo-Json

$page1 = Invoke-RestMethod -Uri "http://localhost:9999/clickhouse/select" `
    -Method POST -Body $body -ContentType "application/json"

# หน้าที่ 2
$body = @{
    table = "customers"
    columns = @("customer_id", "name", "email", "phone")
    where = @{
        shop_id = "shop001"
    }
    order_by = "created_at DESC"
    limit = 20
    offset = 20
} | ConvertTo-Json

$page2 = Invoke-RestMethod -Uri "http://localhost:9999/clickhouse/select" `
    -Method POST -Body $body -ContentType "application/json"
```

---

## Use Cases จริง

### 1. Dashboard - ยอดขายรายวัน
```powershell
$body = @{
    query = @"
SELECT 
    toDate(order_date) as date,
    COUNT(*) as total_orders,
    SUM(total_amount) as total_sales,
    AVG(total_amount) as avg_order_value
FROM orders
WHERE shop_id = 'shop001'
    AND order_date >= today() - INTERVAL 30 DAY
GROUP BY date
ORDER BY date DESC
"@
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:9999/clickhouse/query" `
    -Method POST -Body $body -ContentType "application/json"
```

### 2. Top Products - สินค้าขายดี
```powershell
$body = @{
    query = @"
SELECT 
    product_id,
    product_name,
    COUNT(*) as order_count,
    SUM(quantity) as total_quantity,
    SUM(amount) as total_sales
FROM order_items
WHERE shop_id = 'shop001'
    AND created_at >= '2025-01-01'
GROUP BY product_id, product_name
ORDER BY total_sales DESC
LIMIT 10
"@
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:9999/clickhouse/query" `
    -Method POST -Body $body -ContentType "application/json"
```

### 3. Customer Analysis - วิเคราะห์ลูกค้า
```powershell
$body = @{
    query = @"
SELECT 
    customer_id,
    customer_name,
    COUNT(DISTINCT order_id) as total_orders,
    SUM(total_amount) as lifetime_value,
    MAX(order_date) as last_order_date
FROM orders
WHERE shop_id = 'shop001'
GROUP BY customer_id, customer_name
HAVING total_orders > 5
ORDER BY lifetime_value DESC
LIMIT 100
"@
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:9999/clickhouse/query" `
    -Method POST -Body $body -ContentType "application/json"
```

### 4. Inventory Report - รายงานสต็อก
```powershell
$body = @{
    table = "inventory"
    columns = @("product_id", "product_name", "quantity", "warehouse", "last_updated")
    where = @{
        shop_id = "shop001"
        warehouse = "WH001"
    }
    order_by = "quantity ASC"
    limit = 100
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:9999/clickhouse/select" `
    -Method POST -Body $body -ContentType "application/json"
```

### 5. Time Range Query - ช่วงเวลา
```powershell
$startDate = (Get-Date).AddDays(-7).ToString("yyyy-MM-dd")
$endDate = (Get-Date).ToString("yyyy-MM-dd")

$body = @{
    query = "SELECT * FROM sales WHERE shop_id = 'shop001' AND date >= '$startDate' AND date <= '$endDate' ORDER BY date DESC"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:9999/clickhouse/query" `
    -Method POST -Body $body -ContentType "application/json"
```

---

## Error Handling

### Error 1: Query ว่าง
```json
{
  "status": "error",
  "code": 400,
  "message": "Query is required"
}
```

### Error 2: Table name ว่าง
```json
{
  "status": "error",
  "code": 400,
  "message": "Table name is required"
}
```

### Error 3: SQL Error
```json
{
  "status": "error",
  "code": 500,
  "message": "Failed to execute query",
  "error": "DB::Exception: Table default.invalid_table doesn't exist"
}
```

### Error 4: Connection Failed
```json
{
  "status": "error",
  "code": 500,
  "message": "Failed to connect to ClickHouse",
  "error": "dial tcp: connection refused"
}
```

---

## สรุปการใช้งาน

| Endpoint | ใช้เมื่อไหร่ | ข้อดี | ข้อเสีย |
|----------|-------------|-------|---------|
| `/clickhouse/query` | ต้องการ SQL แบบซับซ้อน (JOIN, GROUP BY, Subquery) | ยืดหยุ่นมาก, ใช้ SQL เต็มรูปแบบ | ต้องเขียน SQL เอง |
| `/clickhouse/select` | Query แบบธรรมดา (SELECT + WHERE) | ง่าย, ไม่ต้องเขียน SQL | จำกัดเฉพาะ SELECT พื้นฐาน |

**คำแนะนำ:**
- ใช้ `/clickhouse/select` สำหรับ query ทั่วไป
- ใช้ `/clickhouse/query` สำหรับ analytics และ reporting ที่ซับซ้อน
- ใช้ `LIMIT` เสมอเพื่อจำกัดผลลัพธ์
- ใช้ index และ WHERE clause เพื่อ performance ที่ดี
