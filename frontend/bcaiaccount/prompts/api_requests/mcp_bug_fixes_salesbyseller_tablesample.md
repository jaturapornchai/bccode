# Bug Fix Request: MCP Tools — get_sales_by_seller & get_table_sample

## สถานะ
- **วันที่ทดสอบ:** 26 ก.พ. 2026
- **Test Shop ID:** `3A3rVmgRs67lD68jMCIn9ibzoM2`
- **ผลทดสอบรวม:** 32/34 tools ผ่าน ✅ — เหลือ 2 bugs ด้านล่าง

---

## Bug #1: `get_sales_by_seller` — ClickHouse column error

### Error Message
```
code: 47, message: Missing columns: 'd.salescode' while processing query:
'SELECT d.salescode AS seller_code, ...
```

### สาเหตุ
Query ใช้ `d.salescode` แต่ ClickHouse table `doc` ไม่มี column `salescode`

### วิธีทดสอบ
```json
{
  "tool": "get_sales_by_seller",
  "params": {
    "from_date": "2025-01-01",
    "to_date": "2025-12-31",
    "shop_id": "3A3rVmgRs67lD68jMCIn9ibzoM2"
  }
}
```

### แนวทางแก้ไข
1. ตรวจสอบ ClickHouse table `doc` ว่า column ชื่อจริงคืออะไร (อาจเป็น `salecode`, `sale_code`, `seller_code`)
2. รัน `DESCRIBE doc` เพื่อดู columns ทั้งหมด
3. แก้ query ใน handler ให้ใช้ชื่อ column ที่ถูกต้อง
4. ถ้า column ไม่มีจริง → อาจต้องเพิ่ม column หรือ join กับ table อื่น

---

## Bug #2: `get_table_sample` — PostgreSQL table not found

### Error Message
```
pq: relation "products" does not exist
```

### วิธีทดสอบ
```json
{
  "tool": "get_table_sample",
  "params": {
    "table_name": "products",
    "shop_id": "3A3rVmgRs67lD68jMCIn9ibzoM2"
  }
}
```

### สาเหตุ
- `get_table_sample` อาจ query PostgreSQL schema ที่ไม่ถูกต้อง
- Table name อาจแตกต่างกันในแต่ละ shop/database (เช่น `product` ไม่ใช่ `products`)
- หรือ PostgreSQL connection ชี้ไป database ที่ไม่มี table นี้

### แนวทางแก้ไข
1. ตรวจสอบว่า `get_table_sample` ใช้ database/schema ไหน
2. ตรวจสอบ table name ที่ถูกต้อง (ใช้ `get_database_schema` ดูก่อน)
3. ถ้า tool `get_database_schema` ทำงานปกติ → ควร validate table name ก่อน query
4. แนะนำ: เพิ่ม validation ว่า table_name ต้อง match กับ `information_schema.tables` ก่อน query จริง

---

## Bug เสริม (optional): `query_clickhouse` type conversion

### Error Message (เฉพาะ `SELECT 1`)
```
converting UInt8 to *uint64 is unsupported
```

### หมายเหตุ
- `SHOW TABLES` ทำงานปกติ ✅
- เฉพาะ `SELECT 1` หรือ query ที่ return literal UInt8 จะเจอ bug นี้
- เป็น Go driver issue (`clickhouse-go`) เรื่อง type mapping UInt8 → uint64
- **Priority ต่ำ** — ไม่กระทบ query ปกติ แต่ควรแก้เพื่อความสมบูรณ์

### แนวทางแก้ไข
1. ใช้ `SELECT toUInt64(1)` แทน `SELECT 1` ใน internal queries
2. หรือ update `clickhouse-go` driver ให้ handle UInt8 → uint64 conversion ได้
3. หรือปรับ result scanner ให้ detect type แล้ว cast ให้ถูก

---

## สรุป

| Bug | Tool | Priority | ผลกระทบ |
|-----|------|----------|---------|
| #1 | `get_sales_by_seller` | **สูง** | ใช้งานไม่ได้เลย — ดูยอดขายตาม seller ไม่ได้ |
| #2 | `get_table_sample` | **กลาง** | ดู sample data จาก PostgreSQL ไม่ได้ (dev tool) |
| #3 | `query_clickhouse` | **ต่ำ** | เฉพาะ SELECT literal — query ปกติทำงานได้ |
