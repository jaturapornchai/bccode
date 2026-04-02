# MCP Tools Guide สำหรับ Frontend

## Overview

Backend มี MCP (Model Context Protocol) tools 35 ตัว ให้ AI ฝั่ง frontend เรียกใช้ได้โดยไม่ต้องอ่าน backend code

- ดูข้อมูลร้านค้า (ยอดขาย, สต็อก, ลูกค้า, dashboard)
- ค้นสินค้า (Thai full-text search)
- ดู API spec, enum values, model schema สำหรับเขียนโค้ด
- Query database ตรง (PostgreSQL, MongoDB, ClickHouse)

---

## Endpoint

```
POST {BASE_URL}/goapi/mcp/invoke
```

| Environment | BASE_URL |
|------------|----------|
| Local (Docker Desktop) | `http://localhost:8888` |
| Dev VPS | `http://api.bcaicloud.com:8888` |

---

## Authentication

ทุก request ต้องมี API key:

```
Header: X-API-Key: bc_live_xxx
```

### สร้าง API key

```bash
curl -X POST {BASE_URL}/goapi/api/mcp/keys/create-with-export \
  -H "Content-Type: application/json" \
  -d '{"name":"frontend-dev","description":"For frontend AI","shop_id":"YOUR_SHOP_ID"}'
```

Response จะได้ `api_key` กลับมา — เก็บไว้ใช้ทุก request

### shop_id ไม่ต้องส่ง

API key มี `shop_id` ฝังอยู่แล้ว — **ไม่ต้องส่ง shop_id ในทุก request**

Backend auto-inject shop_id จาก API key ให้ทุก tool

---

## วิธีเรียก

```json
{
  "tool": "ชื่อ tool",
  "params": {
    // parameters ตาม tool (ไม่ต้องมี shop_id)
  }
}
```

### Response format

```json
{
  "success": true,
  "data": { ... },
  "timestamp": "2025-02-25T10:00:00Z",
  "tool": "tool_name"
}
```

Error:
```json
{
  "success": false,
  "error": "error message",
  "error_code": "ERROR_CODE",
  "timestamp": "...",
  "tool": "tool_name"
}
```

---

## ตัวอย่างการเรียก

### ค้นสินค้า
```bash
curl -X POST {BASE_URL}/goapi/mcp/invoke \
  -H "Content-Type: application/json" \
  -H "X-API-Key: bc_live_xxx" \
  -d '{"tool":"search_products","params":{"keyword":"กาแฟ"}}'
```

### ยอดขายวันนี้
```bash
curl -X POST {BASE_URL}/goapi/mcp/invoke \
  -H "Content-Type: application/json" \
  -H "X-API-Key: bc_live_xxx" \
  -d '{"tool":"get_daily_sales","params":{"date":"2025-02-25"}}'
```

### ดู API spec ของ /login
```bash
curl -X POST {BASE_URL}/goapi/mcp/invoke \
  -H "Content-Type: application/json" \
  -H "X-API-Key: bc_live_xxx" \
  -d '{"tool":"get_api_spec","params":{"path":"/login"}}'
```

### ดู enum values
```bash
curl -X POST {BASE_URL}/goapi/mcp/invoke \
  -H "Content-Type: application/json" \
  -H "X-API-Key: bc_live_xxx" \
  -d '{"tool":"list_enums","params":{"keyword":"transflag"}}'
```

### ดู Dart model schema
```bash
curl -X POST {BASE_URL}/goapi/mcp/invoke \
  -H "Content-Type: application/json" \
  -H "X-API-Key: bc_live_xxx" \
  -d '{"tool":"get_model_schema","params":{"model":"Shop"}}'
```

---

## Tools ทั้งหมด (35 ตัว)

### ค้นสินค้า

| Tool | คำอธิบาย | Parameters |
|------|----------|------------|
| `search_products` | ค้นสินค้า Thai full-text search + stock balance | `keyword` (required), `limit`, `whcode`, `locationcode`, `include_balance` |

### ยอดขาย (Sales)

| Tool | คำอธิบาย | Parameters |
|------|----------|------------|
| `get_daily_sales` | ยอดขายรายวัน | `date` (required, YYYY-MM-DD), `branch_code` |
| `get_sales_by_date_range` | ยอดขายตามช่วงวัน | `from_date`, `to_date` (required), `group_by` (day/week/month), `branch_code` |
| `get_top_selling_products` | สินค้าขายดี | `from_date`, `to_date` (required), `limit`, `branch_code` |
| `get_sales_by_seller` | ยอดขายตามช่องทาง | `from_date`, `to_date` (required) |
| `get_monthly_summary` | สรุปยอดขายรายเดือน | `year`, `month` (required) |

### Dashboard

| Tool | คำอธิบาย | Parameters |
|------|----------|------------|
| `get_dashboard_kpis` | KPI dashboard สำหรับ CEO | `period` (today/this_week/this_month/this_year) |
| `get_business_health` | คะแนนสุขภาพธุรกิจ (0-100) | - |

### การเงิน (Financial)

| Tool | คำอธิบาย | Parameters |
|------|----------|------------|
| `get_profit_analysis` | วิเคราะห์กำไร | `from_date`, `to_date` (required) |
| `get_accounts_receivable` | ลูกหนี้ค้างชำระ (aging) | `limit` |
| `get_accounts_payable` | เจ้าหนี้ค้างชำระ (aging) | `limit` |
| `get_cash_flow` | กระแสเงินสด | `from_date`, `to_date` (required) |

### สต็อก (Inventory)

| Tool | คำอธิบาย | Parameters |
|------|----------|------------|
| `get_inventory_value` | มูลค่าสินค้าคงคลัง | `whcode` |
| `get_low_stock_alerts` | แจ้งเตือนสต็อกต่ำ | `threshold`, `limit` |
| `get_dead_stock` | สินค้าไม่เคลื่อนไหว | `days_no_movement`, `limit` |
| `get_inventory_turnover` | อัตราหมุนเวียนสินค้า | `from_date`, `to_date` (required), `limit` |

### ลูกค้า (Customer)

| Tool | คำอธิบาย | Parameters |
|------|----------|------------|
| `get_top_customers` | ลูกค้า top revenue | `from_date`, `to_date` (required), `limit` |
| `get_customer_growth` | การเติบโตของลูกค้า | `from_date`, `to_date` (required) |
| `get_customer_segments` | วิเคราะห์ RFM segmentation | - |

### เปรียบเทียบ (Comparison)

| Tool | คำอธิบาย | Parameters |
|------|----------|------------|
| `get_yoy_comparison` | เทียบปีต่อปี | `year`, `month` |
| `get_mom_comparison` | เทียบเดือนต่อเดือน | `year`, `month` |

### Database Query

| Tool | คำอธิบาย | Parameters |
|------|----------|------------|
| `get_database_schema` | ดูโครงสร้าง tables | `table_name` |
| `execute_query` | รัน SQL SELECT (readonly) | `query` (required), `limit` |
| `get_table_sample` | ดูตัวอย่างข้อมูล | `table_name` (required), `limit` |

### MongoDB

| Tool | คำอธิบาย | Parameters |
|------|----------|------------|
| `query_mongodb` | Query collection | `collection` (required), `filter`, `limit`, `database` |
| `list_mongodb_collections` | ดู collections ทั้งหมด | `database` |
| `aggregate_mongodb` | Aggregation pipeline | `collection`, `pipeline` (required), `limit`, `database` |

### ClickHouse (OLAP)

| Tool | คำอธิบาย | Parameters |
|------|----------|------------|
| `query_clickhouse` | รัน SQL SELECT (readonly) | `query` (required), `limit`, `database` |
| `list_clickhouse_tables` | ดู tables ทั้งหมด | `database` |

### API Reference (สำหรับ AI เขียนโค้ด)

| Tool | คำอธิบาย | Parameters |
|------|----------|------------|
| `list_api_endpoints` | ดู API ทั้งหมด (1,427 endpoints) | `keyword`, `method`, `category`, `source` (goapi/mainapi), `limit` |
| `get_api_spec` | ดู spec ละเอียด (params, request/response schema) | `path` (required), `method` |
| `get_api_example` | ดูตัวอย่าง curl + Dart + TypeScript | `path` (required), `method` |

### Enum & Model (สำหรับ AI เขียนโค้ด)

| Tool | คำอธิบาย | Parameters |
|------|----------|------------|
| `list_enums` | ดู enum values ทั้งหมด (30 groups) | `category`, `keyword` |
| `get_model_schema` | ดู struct → JSON schema + Dart class + TS interface | `model`, `category`, `keyword` |

---

## ดู Tools ทั้งหมด (JSON)

```
GET {BASE_URL}/goapi/mcp/tools
```

---

## API Key Management

| Action | Method | Path |
|--------|--------|------|
| สร้าง key | POST | `/goapi/api/mcp/keys/create-with-export` |
| ดู keys | GET | `/goapi/api/mcp/keys` |
| ลบ key | DELETE | `/goapi/api/mcp/keys/{id}` |

---

## Claude Code Skills (Slash Commands)

MCP tools เรียกผ่าน curl/HTTP ได้โดยตรง แต่ถ้าใช้ **Claude Code** (AI agent) จะมี **Skills** ที่ครอบ MCP tools ไว้เป็น slash commands ใช้ง่ายกว่ามาก

### Shared Skills Repo

```
c:\bcdev\clone-skills\
├── skills/
│   ├── api-search/SKILL.md   # /api-search — ค้นหา API endpoints
│   ├── api-spec/SKILL.md     # /api-spec — ดู API specification
│   ├── model-gen/SKILL.md    # /model-gen — Generate model class
│   ├── enum-list/SKILL.md    # /enum-list — ดู enum values
│   └── mcp-check/SKILL.md    # /mcp-check — ตรวจสอบ MCP status
├── rules/
│   ├── mcp-first.md           # MCP-First development rules
│   └── erp-conventions.md     # ERP system conventions
└── docs/
    └── mcp-tools-guide.md     # ไฟล์นี้ (copy)
```

GitHub: `https://github.com/jaturapornchai/bcaicloudskill`

### Skills ที่มี

| Skill | คำสั่ง | ใช้ทำอะไร | MCP Tool ที่ใช้ |
|-------|--------|----------|----------------|
| API Search | `/api-search login` | ค้นหา API endpoints ของ backend | `list_api_endpoints` |
| API Spec | `/api-spec login` | ดู request/response spec + curl + code snippet | `get_api_spec` + `get_api_example` + `get_model_schema` |
| Model Gen | `/model-gen Shop` | Generate model class (Dart/TS/Go) จาก MCP schema | `get_model_schema` + `get_database_schema` + `get_table_sample` |
| Enum List | `/enum-list transflag` | ดู enum values / constants | `list_enums` |
| MCP Check | `/mcp-check` | ตรวจสอบ MCP server + ทดสอบ tools | ทดสอบ 5 tools พร้อมกัน |
| Master Data | `/master-data product` | ดูโครงสร้างข้อมูลหลัก (สินค้า, ลูกค้า, คลัง, สาขา ฯลฯ) | `get_model_schema` + `get_database_schema` + `get_table_sample` + `list_api_endpoints` |

### วิธี Setup Skills

#### วิธีที่ 1: Symlink (แนะนำ)

```bash
# Windows (Run as Administrator)
mklink /D "C:\bcdev\bcaiaccount\.claude\skills" "C:\bcdev\clone-skills\skills"
mklink /D "C:\bcdev\bcaiaccount\.claude\rules" "C:\bcdev\clone-skills\rules"

# macOS / Linux
ln -s /path/to/clone-skills/skills /path/to/myproject/.claude/skills
ln -s /path/to/clone-skills/rules /path/to/myproject/.claude/rules
```

#### วิธีที่ 2: User-Level (ใช้ได้ทุก project)

```bash
# Windows
xcopy /E /I "C:\bcdev\clone-skills\skills" "%USERPROFILE%\.claude\skills"
xcopy /E /I "C:\bcdev\clone-skills\rules" "%USERPROFILE%\.claude\rules"

# macOS / Linux
cp -r clone-skills/skills ~/.claude/skills
cp -r clone-skills/rules ~/.claude/rules
```

### ตัวอย่างการใช้ Skills

```
# ค้นหา API เกี่ยวกับ product
/api-search product

# ดู spec ของ login API
/api-spec login

# Generate Dart model จาก Shop schema
/model-gen Shop

# ดู enum transflag (ประเภทเอกสาร)
/enum-list transflag

# เช็คว่า MCP server ทำงานอยู่ไหม
/mcp-check
```

### Workflow แนะนำ

1. เริ่ม dev session → `/mcp-check` (ยืนยัน MCP ทำงาน)
2. หา API → `/api-search {keyword}`
3. ดู spec ละเอียด → `/api-spec {path}`
4. สร้าง model → `/model-gen {model-name}`
5. ดู enum values → `/enum-list {category}`

---

## MCP-First Rules (กฏสำหรับ Frontend)

1. **ใช้ MCP ก่อนเสมอ** — ตรวจสอบ backend ผ่าน MCP tools ก่อนเขียน/แก้ frontend
2. **ทดสอบ MCP ทุก session** — รัน `/mcp-check` ตอนเริ่ม dev
3. **MCP สำหรับ AI เท่านั้น** — Flutter app เรียก REST API ปกติ (ห้ามเรียก MCP endpoint โดยตรง)
4. **MCP invoke format** — ใช้ key `"params"` เสมอ (ไม่ใช่ `"parameters"` หรือ `"arguments"`)

### ห้ามทำ

- ห้ามแก้ backend source code
- ห้ามสร้าง/แก้ DB schema โดยตรง
- ห้าม frontend app เรียก MCP endpoint

### ถ้า API ไม่มี

1. ตรวจ MCP tools ก่อน (`/api-search`)
2. ถ้าไม่มี → เขียน API request prompt ใน `prompts/api_requests/{feature}.md`
3. ส่ง prompt ให้ทีม backend
4. สร้าง frontend ด้วย mock data ไว้ก่อนได้

---

## หมายเหตุ

- ทุก tool เป็น **readonly** — ไม่แก้ไขข้อมูล
- Database query tools อนุญาตเฉพาะ `SELECT` / `SHOW`
- Data isolation: ทุก query filter by `shop_id` จาก API key อัตโนมัติ
- Rate limit: ยังไม่มี (อนาคตจะเพิ่ม)
- Skills repo อยู่ที่ `c:\bcdev\clone-skills\` (local) หรือ GitHub
- Skills update: `cd clone-skills && git pull`
