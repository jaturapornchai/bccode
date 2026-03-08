# MCP API Key — Permission Preset UI

## บริบท

Backend ปรับระบบ MCP API Key permission แล้ว — แยก tools เป็น 2 กลุ่ม:
- **Readonly (35 tools):** ดูข้อมูลอย่างเดียว (sales, dashboard, inventory, units ฯลฯ)
- **Write (6 tools):** สร้าง/แก้ไข/ลบข้อมูล (create_unit, create_units, update_unit, delete_unit, delete_units)

## MCP มีหลาย user ที่ใช้ — ต้องแยกให้ชัดเจน

| ใครใช้ | ใช้ทำอะไร | ระดับสิทธิ์ที่ควรได้ |
|--------|-----------|---------------------|
| **Claude Desktop** (frontend dev) | ค้นหา API, schema, enums ตอน dev | Readonly |
| **Claude Code** (frontend project) | ค้นหา API spec ตอน dev | Readonly |
| **Claude Code** (backend dev) | CRUD data ผ่าน MCP tools | Developer (read+write) |
| **ผู้ใช้ทั่วไป** (ดู dashboard) | ดูรายงาน, analytics | Readonly |

ดังนั้น UI สร้าง API Key ต้องให้เลือก **preset สิทธิ์** ได้ชัดเจน

---

## สิ่งที่ต้องแก้ไข

### 1. เพิ่ม Permission Preset Selector ในหน้า "สร้าง API Key"

เพิ่ม dropdown/radio ให้เลือก preset **ก่อน** ส่วนเลือก tools:

| Preset | `allowed_tools` ที่ส่ง | คำอธิบาย |
|--------|----------------------|----------|
| **Readonly** (ค่าเริ่มต้น) | `["readonly"]` | ดูข้อมูลอย่างเดียว — เหมาะกับ Claude Desktop, frontend dev |
| **Developer** | `["*"]` | ทุก tool (อ่าน+เขียน) — เหมาะกับ backend dev |
| **Custom** | `["readonly", "create_unit", ...]` | เลือก tools เองทีละตัว |

**UX Flow:**
1. เลือก Preset → ถ้าเป็น Readonly หรือ Developer → ซ่อน tool checkboxes (ไม่ต้องเลือก)
2. เลือก Custom → แสดง tool checkboxes ให้เลือกทีละตัว
3. ส่ง `allowed_tools` ตาม preset ที่เลือก

### 2. อัปเดตรายการ Tool Categories

รายการ tool categories ปัจจุบันยังไม่ครบ ต้องเพิ่ม:

| Category | Tools |
|----------|-------|
| Sales | `get_daily_sales`, `get_monthly_sales`, `get_sales_by_product`, `get_sales_by_channel`, `get_sales_by_payment` |
| Dashboard | `get_dashboard_summary`, `get_top_products`, `get_hourly_sales` |
| Financial | `get_profit_loss`, `get_revenue_trend`, `get_tax_summary`, `get_payment_summary` |
| Inventory | `get_stock_balance`, `get_low_stock_alerts`, `get_stock_movement`, `get_stock_value` |
| Customers | `get_customer_summary`, `get_top_customers`, `get_new_customers`, `get_customer_purchase_history` |
| Comparison | `compare_sales_periods` |
| Products | `list_products`, `search_products` |
| **Unit of Measure** (ใหม่) | `list_units`, `search_units`, `create_unit`, `create_units`, `update_unit`, `delete_unit`, `delete_units` |
| **API Development** (ใหม่) | `get_api_catalog`, `get_api_spec` |
| **Database** (ใหม่) | `query_clickhouse`, `query_mongodb` |
| **Schema** (ใหม่) | `get_model_schema`, `get_enum_catalog` |

**หมายเหตุ:** Tools ที่มี badge `write` = เฉพาะ Developer preset ถึงจะใช้ได้:
- `create_unit`, `create_units`, `update_unit`, `delete_unit`, `delete_units`

### 3. แก้ไขการแสดงจำนวน Tools

ปัจจุบันแสดง "1 tools" เพราะนับจำนวน string ใน `allowed_tools[]`

**แก้ไข:** ถ้า `allowed_tools` มี keyword → แสดง label แทนตัวเลข:

| `allowed_tools` | แสดงผล |
|----------------|--------|
| `["readonly"]` | "Readonly (35 tools)" |
| `["*"]` | "All tools (41 tools)" |
| `["readwrite"]` | "All tools (41 tools)" |
| `["get_daily_sales", "list_units"]` | "2 tools" |
| `["readonly", "create_unit"]` | "Readonly + 1 write tool" |

### 4. แสดง Permission Badge ในรายการ Keys

ในหน้ารายการ API Keys ให้แสดง badge ระดับสิทธิ์:

| `allowed_tools` | Badge |
|----------------|-------|
| `["*"]` หรือ `["readwrite"]` | 🔧 Developer |
| `["readonly"]` | 👁 Readonly |
| อื่นๆ | 🔧 Custom |

---

## API Endpoints (ไม่เปลี่ยน)

Backend API ยังเหมือนเดิม — แค่ส่ง `allowed_tools` ต่างกัน:

### สร้าง Key
```
POST /goapi/api/mcp/keys
Content-Type: application/json

{
  "shop_id": "...",
  "name": "Claude Desktop Key",
  "description": "สำหรับ Claude Desktop ใช้ค้นหา API",
  "allowed_tools": ["readonly"],        // ← เปลี่ยนตรงนี้
  "created_by": "admin"
}
```

### สร้าง Key + Export Config (แนะนำ)
```
POST /goapi/api/mcp/keys/create-with-export
Content-Type: application/json

{
  "shop_id": "...",
  "name": "Backend Dev Key",
  "allowed_tools": ["*"],               // ← Developer preset
  "created_by": "admin"
}
```

### อัปเดต Key
```
PUT /goapi/api/mcp/keys/:id
Content-Type: application/json

{
  "allowed_tools": ["readonly"]          // ← เปลี่ยน permission
}
```

### ค่า `allowed_tools` ที่ Backend รองรับ

| ค่า | ความหมาย |
|-----|---------|
| `["readonly"]` | Readonly ทั้งหมด (35 tools) — **ค่าเริ่มต้น** ถ้าไม่ส่ง |
| `["*"]` | ทุก tool (41 tools) |
| `["readwrite"]` | เหมือน `["*"]` |
| `["readonly", "create_unit"]` | Readonly + create_unit |
| `["get_daily_sales", "list_units"]` | เฉพาะ tools ที่ระบุ |

---

## สรุปสิ่งที่ต้องทำ

1. **เพิ่ม Preset Selector** (Readonly / Developer / Custom) ในหน้าสร้าง API Key
2. **เพิ่ม Tool Categories ใหม่** (Unit of Measure, API Development, Database, Schema)
3. **แก้จำนวน Tools** จากนับ string → แสดง label ตาม keyword
4. **เพิ่ม Permission Badge** ในรายการ Keys (Readonly / Developer / Custom)
5. **Write tools ต้องมี badge** บอกว่าเป็น write tool (ใช้ได้เฉพาะ Developer/Custom)
