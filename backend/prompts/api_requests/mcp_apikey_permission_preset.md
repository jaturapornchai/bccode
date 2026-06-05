# MCP API Key — Permission Preset UI

## บริบท

Backend ปรับระบบ MCP API Key permission แล้ว — แยก tools เป็น 2 กลุ่ม:
- **Readonly (35 tools):** ดูข้อมูลอย่างเดียว (sales, dashboard, inventory, units ฯลฯ)
- **Write (6 tools):** สร้าง/แก้ไข/ลบข้อมูล (createunit, createunits, updateunit, deleteunit, deleteunits)

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

| Preset | `allowedtools` ที่ส่ง | คำอธิบาย |
|--------|----------------------|----------|
| **Readonly** (ค่าเริ่มต้น) | `["readonly"]` | ดูข้อมูลอย่างเดียว — เหมาะกับ Claude Desktop, frontend dev |
| **Developer** | `["*"]` | ทุก tool (อ่าน+เขียน) — เหมาะกับ backend dev |
| **Custom** | `["readonly", "createunit", ...]` | เลือก tools เองทีละตัว |

**UX Flow:**
1. เลือก Preset → ถ้าเป็น Readonly หรือ Developer → ซ่อน tool checkboxes (ไม่ต้องเลือก)
2. เลือก Custom → แสดง tool checkboxes ให้เลือกทีละตัว
3. ส่ง `allowedtools` ตาม preset ที่เลือก

### 2. อัปเดตรายการ Tool Categories

รายการ tool categories ปัจจุบันยังไม่ครบ ต้องเพิ่ม:

| Category | Tools |
|----------|-------|
| Sales | `getdailysales`, `getmonthlysales`, `getsalesbyproduct`, `getsalesbychannel`, `getsalesbypayment` |
| Dashboard | `getdashboardsummary`, `gettopproducts`, `gethourlysales` |
| Financial | `getprofitloss`, `getrevenuetrend`, `gettaxsummary`, `getpaymentsummary` |
| Inventory | `getstockbalance`, `getlowstockalerts`, `getstockmovement`, `getstockvalue` |
| Customers | `getcustomersummary`, `gettopcustomers`, `getnewcustomers`, `getcustomerpurchasehistory` |
| Comparison | `comparesalesperiods` |
| Products | `listproducts`, `searchproducts` |
| **Unit of Measure** (ใหม่) | `listunits`, `searchunits`, `createunit`, `createunits`, `updateunit`, `deleteunit`, `deleteunits` |
| **API Development** (ใหม่) | `getapicatalog`, `getapispec` |
| **Database** (ใหม่) | `queryclickhouse`, `querymongodb` |
| **Schema** (ใหม่) | `getmodelschema`, `getenumcatalog` |

**หมายเหตุ:** Tools ที่มี badge `write` = เฉพาะ Developer preset ถึงจะใช้ได้:
- `createunit`, `createunits`, `updateunit`, `deleteunit`, `deleteunits`

### 3. แก้ไขการแสดงจำนวน Tools

ปัจจุบันแสดง "1 tools" เพราะนับจำนวน string ใน `allowedtools[]`

**แก้ไข:** ถ้า `allowedtools` มี keyword → แสดง label แทนตัวเลข:

| `allowedtools` | แสดงผล |
|----------------|--------|
| `["readonly"]` | "Readonly (35 tools)" |
| `["*"]` | "All tools (41 tools)" |
| `["readwrite"]` | "All tools (41 tools)" |
| `["getdailysales", "listunits"]` | "2 tools" |
| `["readonly", "createunit"]` | "Readonly + 1 write tool" |

### 4. แสดง Permission Badge ในรายการ Keys

ในหน้ารายการ API Keys ให้แสดง badge ระดับสิทธิ์:

| `allowedtools` | Badge |
|----------------|-------|
| `["*"]` หรือ `["readwrite"]` | 🔧 Developer |
| `["readonly"]` | 👁 Readonly |
| อื่นๆ | 🔧 Custom |

---

## API Endpoints (ไม่เปลี่ยน)

Backend API ยังเหมือนเดิม — แค่ส่ง `allowedtools` ต่างกัน:

### สร้าง Key
```
POST /goapi/api/mcp/keys
Content-Type: application/json

{
  "holdingcode": "...",
  "name": "Claude Desktop Key",
  "description": "สำหรับ Claude Desktop ใช้ค้นหา API",
  "allowedtools": ["readonly"],        // ← เปลี่ยนตรงนี้
  "createdby": "admin"
}
```

### สร้าง Key + Export Config (แนะนำ)
```
POST /goapi/api/mcp/keys/create-with-export
Content-Type: application/json

{
  "holdingcode": "...",
  "name": "Backend Dev Key",
  "allowedtools": ["*"],               // ← Developer preset
  "createdby": "admin"
}
```

### อัปเดต Key
```
PUT /goapi/api/mcp/keys/:id
Content-Type: application/json

{
  "allowedtools": ["readonly"]          // ← เปลี่ยน permission
}
```

### ค่า `allowedtools` ที่ Backend รองรับ

| ค่า | ความหมาย |
|-----|---------|
| `["readonly"]` | Readonly ทั้งหมด (35 tools) — **ค่าเริ่มต้น** ถ้าไม่ส่ง |
| `["*"]` | ทุก tool (41 tools) |
| `["readwrite"]` | เหมือน `["*"]` |
| `["readonly", "createunit"]` | Readonly + createunit |
| `["getdailysales", "listunits"]` | เฉพาะ tools ที่ระบุ |

---

## สรุปสิ่งที่ต้องทำ

1. **เพิ่ม Preset Selector** (Readonly / Developer / Custom) ในหน้าสร้าง API Key
2. **เพิ่ม Tool Categories ใหม่** (Unit of Measure, API Development, Database, Schema)
3. **แก้จำนวน Tools** จากนับ string → แสดง label ตาม keyword
4. **เพิ่ม Permission Badge** ในรายการ Keys (Readonly / Developer / Custom)
5. **Write tools ต้องมี badge** บอกว่าเป็น write tool (ใช้ได้เฉพาะ Developer/Custom)
