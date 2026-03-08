# BC AI Cloud — Flutter App Rules

## Overview
Flutter multi-platform app (Web, Windows, Android, iOS) for BC AI Cloud accounting system.

- **Package:** `smlaicloud`
- **State Management:** BLoC pattern
- **HTTP Client:** Dio
- **Auth:** Firebase Auth (Google Sign-In, LINE Login) + backend JWT

## Project Structure
```
├── lib/
│   ├── main_bcaidev.dart      # Entry: DEV flavor
│   ├── main_bcaiuat.dart      # Entry: UAT flavor
│   ├── main_bcaiprod.dart     # Entry: PROD flavor
│   ├── flavors.dart           # Flavor enum (bcaidev, bcaiuat, bcaiprod)
│   ├── environment.dart       # Environment config (URLs, settings)
│   ├── app_const.dart         # URL constants
│   ├── global.dart            # Global state
│   ├── api/client.dart        # Dio HTTP client
│   ├── bloc/                  # BLoC state management
│   ├── model/                 # Data models
│   ├── repositories/          # API repositories
│   ├── screens/               # Feature screens
│   ├── usersystem/            # Login, shop selection
│   └── utils/                 # Utilities
├── web/config.json            # Backend URLs (goapi_url)
├── android/                   # Android platform
├── windows/                   # Windows platform
└── scripts/                   # Deploy scripts
```

## Backend URLs (config.json)
```json
{
  "goapi_url": "http://localhost:9090/goapi"
}
```
- **MainAPI (port 9090):** Cloud platform API (auth, CRUD, organizations)
- **GoAPI:** รวมเข้า MainAPI แล้ว — ใช้ path prefix `/goapi` (เช่น `http://localhost:9090/goapi/api/health`)

## Critical Rules

### AI Communication Language
- **AI ต้องสื่อสารกับผู้ใช้ (User) เป็นภาษาไทยเป็นหลักในทุกๆ การตอบสนองและการแจ้งเตือน** (AI must communicate with the user primarily in Thai for all responses and notifications).

### ห้าม Frontend เรียก AI API โดยตรง (Security Rule)

**Flutter app ห้ามเรียก AI API key ของทุก provider โดยตรง — ต้องผ่าน Backend เท่านั้น**

เพื่อความปลอดภัย:
- **ห้าม** hardcode AI API key ใดๆ ใน Flutter code (OpenRouter, Groq, DeepSeek, Gemini, OpenAI ฯลฯ)
- **ห้าม** เรียก AI API endpoint ตรง (เช่น `api.groq.com`, `openrouter.ai/api`, `api.deepseek.com`, `generativelanguage.googleapis.com`)
- **ต้อง** เรียกผ่าน backend API เสมอ (เช่น `/goapi/api/v1/chatbot/chat-gemini`)
- AI API key ทั้งหมดเก็บที่ backend (ตั้งค่าผ่านหน้า Setup → Integrations)
- Backend เป็นคนเรียก AI provider แล้ว return ผลกลับมาให้ Flutter

**เหตุผล:** API key ที่ฝังใน frontend สามารถถูกดึงออกได้ง่าย (decompile, network sniff) ทำให้ถูกขโมยใช้งานได้


### MCP-First Standard

**MCP มีไว้สำหรับ AI tools เท่านั้น** (Claude Code, Claude Desktop, Cursor, VSCode + MCP extension)
เพื่อให้ AI สามารถ:
- ดู database schema, sample data, business logic ของ backend ได้
- query ข้อมูลจาก PostgreSQL, MongoDB, ClickHouse เพื่อวิเคราะห์
- เชื่อมต่อจากทุก AI client (Claude Desktop, VSCode, Cursor) ผ่าน SSE protocol มาตรฐาน
- ใช้เป็น source of truth ว่า backend มี API อะไรบ้าง

**MCP ไม่ใช่สำหรับ Flutter app เรียกโดยตรง** — Flutter ใช้ REST API ปกติผ่าน Dio client

**กฏสำหรับ AI:**
1. **อ่าน MCP Tools ก่อนเสมอ** — `GET /goapi/mcp/tools` จะบอกว่า backend มี API อะไรบ้าง, parameter อะไร, return อะไร
2. **ใช้ MCP เป็น source of truth** — ถ้า MCP tools list ไม่มี endpoint ที่ต้องการ แปลว่า backend ยังไม่มี → ส่ง API Specification Prompt ใน `prompts/api_requests/{feature}.md`
3. **อ่าน backend source code ได้ถ้าจำเป็น** — ใช้ MCP เป็นหลัก แต่ถ้า MCP ไม่เพียงพอ สามารถ read ไฟล์ใน `d:\bcdev\backend\` ได้โดยตรง อ่านได้อย่างเดียว ห้ามแก้ไข
4. **ตรวจสอบ backend ก่อนเขียน/แก้ frontend เสมอ** — ใช้ MCP หรืออ่าน source code เพื่อยืนยัน field, type, response format ให้ถูกต้องก่อนเสมอ
5. **ทดสอบ MCP ทุกครั้ง** — ใช้ MCP tool เพื่อยืนยันว่า MCP server ใช้งานได้จริงก่อนเริ่ม dev ถ้า MCP ไม่ตอบ ให้แจ้ง user ก่อนดำเนินการต่อ
6. **MCP ใช้เพื่อ dev เท่านั้น** — Flutter app ต้องเรียก REST API ผ่าน Dio client ทั้งหมด ห้าม Flutter เรียก MCP endpoint โดยตรงในทุกกรณี
7. **แก้ backend ได้ถ้าจำเป็น** — AI สามารถอ่าน+แก้ backend code ได้โดยตรง แต่ต้องระวังเรื่องโครงสร้างข้อมูลที่เชื่อมกับ frontend
8. **โครงสร้าง DB ต้องสร้างผ่าน backend เท่านั้น** — ห้าม AI สร้าง/แก้ schema, migration, collection, table, index ใน MongoDB, PostgreSQL, ClickHouse โดยตรง ต้องให้ backend จัดการเสมอ ถ้าต้องการ schema ใหม่ → เขียน prompt ใน `prompts/api_requests/`

### BACKEND ACCESS: อ่าน+แก้ได้ (Cross-Project)

**AI สามารถอ่านและแก้ไข source code ได้ทั้ง 2 project:**
- **Backend** `D:\bcdev\backend` — อ่าน+แก้ได้
- **Frontend** `D:\bcdev\bcaiaccount` — อ่าน+แก้ได้
- ตรวจ code ข้ามไปมาได้เสมอ เช่น ตรวจ Backend response แล้วแก้ Frontend model ให้ตรงกัน

**ลำดับความสำคัญ:**
1. ใช้ MCP tools เป็นหลัก (เร็วกว่า, ได้ข้อมูล runtime จริง)
2. ถ้า MCP ไม่เพียงพอหรือไม่ชัดเจน → อ่าน/แก้ `d:\bcdev\backend\` ได้โดยตรง
3. แก้ไข backend ได้ถ้าจำเป็น แต่ต้องระวังเรื่องโครงสร้างข้อมูลที่เชื่อมกัน

**ทำได้ (ALLOWED):**
- read/browse/search/edit files ใน `d:\bcdev\backend\` เพื่อทำความเข้าใจและแก้ไข logic, model, handler
- อ้างอิง backend source code เพื่อยืนยัน field, type, response format
- แก้ไข backend code เมื่อจำเป็นเพื่อให้ frontend-backend ทำงานประสานกัน

**ข้อควรระวัง:**
- แก้ Backend response → **ต้องตรวจว่า Frontend ยังใช้ได้**
- เพิ่ม field ใหม่ → ปลอดภัย (Frontend เก่ายัง work)
- ลบ/เปลี่ยนชื่อ field → **อันตราย** → ต้องแก้ทั้ง 2 ฝั่ง
- import หรือ reference backend codebase จาก Flutter app → **ห้าม**

**ต้องทำ (MUST):**
- ใช้ MCP tools เพื่อสืบค้นข้อมูล database schema, sample data, business logic
- ใช้ MCP tools เพื่อทดสอบ queries และตรวจสอบข้อมูล
- ใช้ MCP tools เพื่อดู response format ก่อนสร้าง Dart model
- ตรวจสอบโครงสร้างข้อมูลทั้ง 2 ฝั่งให้ตรงกันเสมอ

**เมื่อ API ที่ต้องการยังไม่มี:**
- สร้าง API ใหม่ใน backend ได้โดยตรง หรือเขียน **API Specification Prompt** ไว้ใน `prompts/api_requests/`
- Format ของ API Specification Prompt ดูหัวข้อ "API Specification Prompt Format" ด้านล่าง

### MCP Endpoints

**Base URL:** `http://localhost:9090/goapi`

MCP แยกเป็น 2 ชุด endpoint:

#### General (สำหรับ Frontend dev / ใช้งานทั่วไป)
| Endpoint | Description |
|----------|-------------|
| `GET /goapi/mcp/health` | ตรวจสอบสถานะ MCP server |
| `GET /goapi/mcp/tools` | รายการ General tools (business data เท่านั้น) |
| `GET /goapi/mcp/sse` | SSE endpoint (สำหรับ Claude Desktop / AI Agent) |
| `POST /goapi/mcp/message` | ส่ง message ผ่าน SSE session |
| `POST /goapi/mcp/invoke` | เรียกใช้ MCP tool โดยตรง |

เห็นเฉพาะ tools ที่เกี่ยวกับ business data (sales, inventory, customers, products, barcodes, units, product groups, product categories)

#### Dev (สำหรับ Backend dev / database access)
| Endpoint | Description |
|----------|-------------|
| `GET /goapi/mcp/dev/health` | ตรวจสอบสถานะ MCP dev server |
| `GET /goapi/mcp/dev/tools` | รายการ ทุก tools (General + Dev) |
| `GET /goapi/mcp/dev/sse` | SSE endpoint สำหรับ dev |

Dev-only tools (16 ตัว — ซ่อนจาก General):
`get_database_schema`, `execute_query`, `get_table_sample`, `query_mongodb`, `list_mongodb_collections`, `aggregate_mongodb`, `query_clickhouse`, `list_clickhouse_tables`, `execute_pg_command`, `execute_ch_command`, `list_api_endpoints`, `get_api_spec`, `get_api_example`, `list_enums`, `get_model_schema`, `rebuild_products`

#### Config
- **API Key Format:** `bc_live_{32chars}` — ขอจาก `POST /api/mcp/keys`
- **Auth Methods:** Header `X-API-Key`, Query `?api_key=`, Bearer token
- **Frontend dev** → ใช้ `/goapi/mcp/sse`
- **Backend dev** → ใช้ `/goapi/mcp/dev/sse`
- General endpoint ปลอดภัยสำหรับ AI — ไม่มี tools ที่แก้ไข database โดยตรง
- Dev endpoint มี `execute_pg_command` / `execute_ch_command` ที่ทำ INSERT/UPDATE/DELETE ได้

### GoAPI URL Mapping (เดิม → ใหม่)

GoAPI รวมเข้า MainAPI แล้ว — port เดียว (9090) เติม `/goapi` prefix หน้า path เดิม

| เดิม | ใหม่ |
|------|------|
| `goapi:9091/api/health` | `localhost:9090/goapi/api/health` |
| `goapi:9091/get` | `localhost:9090/goapi/get` |
| `goapi:9091/genpdf` | `localhost:9090/goapi/genpdf` |
| `goapi:9091/mcp/sse` | `localhost:9090/goapi/mcp/sse` |
| `goapi:9091/api/product/search` | `localhost:9090/goapi/api/product/search` |
| `goapi:9091/s3/file/*` | `localhost:9090/goapi/s3/file/*` |
| `goapi:9091/image/upload` | `localhost:9090/goapi/image/upload` |
| `goapi:9091/api/lineoa/*` | `localhost:9090/goapi/api/lineoa/*` |
| `goapi:9091/api/approval/*` | `localhost:9090/goapi/api/approval/*` |
| `goapi:9091/api/setup/*` | `localhost:9090/goapi/api/setup/*` |
| `goapi:9091/clickhouse/*` | `localhost:9090/goapi/clickhouse/*` |
| `goapi:9091/api/v1/chatbot/*` | `localhost:9090/goapi/api/v1/chatbot/*` |
| `goapi:9091/api/v1/unified/*` | `localhost:9090/goapi/api/v1/unified/*` |
| `goapi:9091/version` | `localhost:9090/goapi/version` |

**mainapi routes ที่ไม่เปลี่ยน** (ไม่มี `/goapi` prefix):
- `/login`, `/register`, `/refresh` — auth
- `/shop`, `/product/*`, `/warehouse` — CRUD
- `/healthz` — health check
- `/swagger/*` — API docs

### MCP Tools ที่ใช้ได้

**เรียก `GET /goapi/mcp/tools` (General) หรือ `GET /goapi/mcp/dev/tools` (Dev) เพื่อดูรายการล่าสุดเสมอ**

#### General Tools (เห็นทั้ง General + Dev endpoint)

##### ค้นหาสินค้า
- `search_products` — ค้นหาสินค้า + ยอด stock (รองรับภาษาไทย)

##### ข้อมูลการขาย
- `get_daily_sales` — ยอดขายรายวัน
- `get_sales_by_date_range` — ยอดขายช่วงวันที่ (group by day/week/month)
- `get_top_selling_products` — สินค้าขายดี
- `get_sales_by_seller` — ยอดขายตามพนักงาน
- `get_monthly_summary` — สรุปรายเดือน เทียบเดือนก่อน

##### Dashboard & KPI
- `get_dashboard_kpis` — KPI รวม (sales, orders, profit, customers, inventory)
- `get_dashboard_summary` — Dashboard overview
- `get_business_health` — คะแนนสุขภาพธุรกิจ (0-100) + คำแนะนำ

##### การเงิน
- `get_profit_analysis` — วิเคราะห์กำไร แยกตามหมวด
- `get_financial_summary` — สรุปการเงิน
- `get_accounts_receivable` — ลูกหนี้ + aging buckets (30/60/90+ วัน)
- `get_accounts_payable` — เจ้าหนี้ + aging buckets
- `get_cash_flow` — กระแสเงินสด (รายรับ/รายจ่าย/สุทธิ)

##### สินค้าคงคลัง
- `get_inventory_value` — มูลค่าสินค้าคงคลัง แยกตามหมวด/คลัง
- `get_inventory_status` — สถานะสต็อก
- `get_low_stock_alerts` — สินค้าใกล้หมด/หมดสต๊อก
- `get_dead_stock` — สินค้าไม่เคลื่อนไหว (dead stock)
- `get_inventory_turnover` — อัตราหมุนเวียนสินค้า

##### ลูกค้า
- `get_top_customers` — ลูกค้ารายใหญ่ (sort by amount/orders/profit)
- `get_customer_growth` — การเติบโตของลูกค้า
- `get_customer_segments` — วิเคราะห์ลูกค้า RFM (Recency, Frequency, Monetary)

##### เปรียบเทียบ
- `get_yoy_comparison` — เทียบปีต่อปี
- `get_mom_comparison` — เทียบเดือนต่อเดือน

##### บาร์โค้ด / หน่วยนับ / กลุ่มสินค้า
- `list_barcodes` — ดูรายการบาร์โค้ด
- `create_barcode` — สร้างบาร์โค้ด
- `get_ref_barcodes` — ดูบาร์โค้ดอ้างอิง (reference chain)
- `set_ref_barcode` — ตั้งค่าบาร์โค้ดอ้างอิง
- `create_multi_unit_barcode` — สร้างสินค้าหลายหน่วยนับ

#### Dev-Only Tools (เห็นเฉพาะ Dev endpoint — 16 ตัว)

##### Database
- `get_database_schema` — ดูโครงสร้าง tables, columns, relationships
- `execute_query` — รัน SELECT query (readonly)
- `get_table_sample` — ดูข้อมูลตัวอย่างจาก table
- `execute_pg_command` — รัน INSERT/UPDATE/DELETE (⚠️ แก้ไขข้อมูลได้)

##### MongoDB
- `query_mongodb` — query MongoDB collection (readonly)
- `list_mongodb_collections` — ดู collections ทั้งหมด
- `aggregate_mongodb` — รัน aggregation pipeline (readonly)

##### ClickHouse
- `query_clickhouse` — รัน SELECT/SHOW query (readonly)
- `list_clickhouse_tables` — ดู tables + engine info
- `execute_ch_command` — รัน INSERT/UPDATE/DELETE (⚠️ แก้ไขข้อมูลได้)

##### API / Schema
- `list_api_endpoints` — ดูรายการ API endpoints ทั้งหมด
- `get_api_spec` — ดู API specification
- `get_api_example` — ดูตัวอย่าง API request/response
- `list_enums` — ดู enum values
- `get_model_schema` — ดู model schema
- `rebuild_products` — rebuild products index

### วิธีใช้ MCP ใน development

#### เมื่อต้องการรู้ database schema
```
ใช้ tool: get_database_schema
แทนการไปอ่าน migration files หรือ model files ใน backend
```

#### เมื่อต้องการรู้ว่า API return อะไร
```
ใช้ tool: execute_query หรือ get_table_sample
เพื่อดูข้อมูลจริงแทนการอ่าน handler code ใน backend
```

#### เมื่อต้องสร้าง model ใหม่ใน Flutter
```
1. ใช้ get_database_schema เพื่อดู columns + data types
2. ใช้ get_table_sample เพื่อดูข้อมูลตัวอย่าง
3. สร้าง Dart model ตาม schema ที่ได้
```

#### เมื่อต้องสร้าง report/dashboard ใน Flutter
```
1. ใช้ get_dashboard_kpis หรือ tool ที่เกี่ยวข้อง
2. ดู response format ที่ได้
3. สร้าง model + UI ตาม response structure
```

#### เมื่อ API ที่ต้องการยังไม่มี
```
1. ตรวจสอบก่อนว่า MCP tools ที่มีอยู่ตอบโจทย์ได้ไหม
2. ถ้าไม่มี → เขียน API Specification Prompt
3. บันทึกไว้ที่ prompts/api_requests/{feature_name}.md
4. แจ้ง user ว่าต้องส่ง prompt ไปให้ทีม backend สร้าง API ก่อน
5. สร้าง Flutter code ด้าน frontend ไว้ก่อนได้ (mock data / placeholder)
```

### API Specification Prompt Format

เมื่อต้องการ API ใหม่ที่ backend ยังไม่มี ให้สร้างไฟล์ใน `prompts/api_requests/` ตาม format นี้:

```markdown
# API Request: {ชื่อ feature}

## สิ่งที่ต้องการ
{อธิบายสั้นๆ ว่าต้องการ API อะไร ทำไม}

## Endpoint ที่ต้องการ
- **Method:** GET/POST/PUT/DELETE
- **Path:** /api/{resource}
- **Auth:** Bearer token (JWT)

## Request
{request body / query params ที่ต้องการส่ง}
```json
{
  "field1": "type + description",
  "field2": "type + description"
}
```

## Expected Response
{response format ที่ frontend ต้องการรับ}
```json
{
  "success": true,
  "data": {
    "field1": "type + description",
    "field2": "type + description"
  }
}
```

## Use Case
{อธิบาย use case จากฝั่ง frontend — หน้าจอไหน, ทำอะไร}

## Database Tables ที่เกี่ยวข้อง
{ถ้าใช้ MCP get_database_schema ดูแล้ว ใส่ข้อมูลที่ได้มา}

## MCP Tool ที่อยากได้ (optional)
{ถ้าต้องการเพิ่มเป็น MCP tool ด้วย ระบุชื่อ + parameters}
- **Tool name:** {tool_name}
- **Parameters:** {param1 (type), param2 (type)}
```

### Development Workflow ข้าม Project

```
┌─────────────────┐     MCP Tools      ┌─────────────────┐
│   Flutter App    │ ◄──────────────►   │   Backend API   │
│  (bcaiaccount)   │   SSE + JSON-RPC   │   (Main  9090)  │
│                  │                    │  GoAPI = /goapi  │
│  AI อ่าน+แก้ได้   │                    │                  │
│                  │                    │  AI อ่าน+แก้ได้   │
│                  │                    │  + ใช้ MCP ด้วย  │
└─────────────────┘                    └─────────────────┘
        │                                       ▲
        │  ถ้า API ไม่มี                          │
        ▼                                       │
┌─────────────────┐                             │
│ prompts/         │  เขียน spec หรือ             │
│ api_requests/    │  แก้ backend โดยตรง ─────────┘
│ {feature}.md     │
└─────────────────┘
```

## Flavors
- `bcaidev` — Development (localhost)
- `bcaiuat` — UAT (uat server)
- `bcaiprod` — Production

## Code Conventions
- ใช้ภาษาไทยใน log messages และ comments ได้
- Logger: `AppLogger.info()`, `AppLogger.error()`, `AppLogger.debug()`, `AppLogger.warning()`
- State management: BLoC pattern with events/states
- HTTP: Dio client via `api/client.dart`
- Auth token: stored in `global.appConfig.getString("token")`

### ห้าม Mock Data / Hardcoded Test Data

**ห้ามใส่ mock data, sample data, หรือ hardcoded test data ใน production code**

**ห้ามทำ (NEVER):**
- ห้าม hardcode ชื่อลูกค้า, ชื่อสินค้า, ราคา, ที่อยู่ เป็นตัวอย่างใน UI code
- ห้ามใส่ข้อมูลจำลอง เช่น `"สิธิพัช สามวง"`, `"บะหมี่กึ่งสำเร็จรูป"`, `"1,800.00 บาท"` ใน widget code
- ห้าม hardcode รายการ mock items ใน ListView, GridView หรือ component ใดๆ
- ห้ามใช้ค่าตัวเลขจำลอง เช่น ราคา, จำนวน, ยอดรวม ที่ไม่ได้มาจาก API

**ต้องทำแทน (DO THIS INSTEAD):**
- ใช้ข้อมูลจาก API/Backend เสมอ — เรียกผ่าน Repository + BLoC
- ถ้า API ยังไม่พร้อม → แสดง Empty State (เช่น "ไม่มีข้อมูล") หรือ Loading State
- ถ้าต้องการ placeholder text → ใช้ `language("key")` ที่เป็น generic เช่น `language("no_data")`
- ถ้าต้องการ UI preview สำหรับ dev → ใช้ flag/env check แล้วแสดงเฉพาะ DEV mode

**เหตุผล:**
- Mock data ทำให้ลืมเชื่อม API จริง
- ข้อมูลจำลองภาษาไทยทำให้ localization ยากขึ้น (ต้อง skip ทีละตัว)
- Production app ไม่ควรมี test data ปนอยู่

### Localization (Multi-language Support)

**ทุก UI text ที่ user เห็นต้องใช้ `language("key")` หรือ `global.language("key")`**

**กฏ:**
- ห้าม hardcode ภาษาไทยใน UI code (ยกเว้น login flow ที่ยัง load language ไม่ได้)
- ใช้ `global.language("key")` สำหรับ screens, components, widgets
- ใช้ `global.language("key")` สำหรับ PDF generation (`pdfgen/`)
- Language keys อยู่ใน `D:\bcdev\backend\assets\language\languages.json` (shared กับ backend)
- ถ้าต้องการ key ใหม่ → เพิ่มใน `languages.json` พร้อม 9 ภาษา (th, en, cn, ja, km, ko, lo, my, vi)
- Login flow files ที่ยกเว้น: `login_password_screen.dart`, `login_screen.dart`, `registration.dart`, `select_language_screen.dart`, `login_line_screen.dart`, `select_shop.dart`, `server_config_screen.dart`

**ข้อยกเว้นที่ไม่ต้อง localize:**
- Log messages (`AppLogger.*`, `debugPrint`, `print()`)
- Comments (`//`, `/* */`)
- Month/Day names ใน arrays (ใช้ใน DatePicker)
- NumberToWord constants (ศูนย์, หนึ่ง, ร้อย, พัน...)
- Error messages ใน `Exception()`, `throw`

## กฏสำคัญ: สรุปว่าต้องแก้ Backend หรือไม่

**ทุกครั้งที่ทำงาน feature หรือ fix bug เสร็จ ต้องสรุปให้ user ทราบเสมอว่า:**

> **ต้องแก้ Backend หรือไม่?**
> - **ไม่ต้องแก้** — ถ้า frontend แก้เองครบ + backend มี API/fields พร้อมแล้ว
> - **ต้องแก้** — ถ้า API ยังไม่มี หรือ backend ต้องเพิ่ม field/endpoint/logic ใหม่
>   → ให้เขียน prompt ไว้ใน `prompts/api_requests/{feature}.md`
>   → แจ้ง user ว่าต้องส่ง prompt ไปให้ทีม backend

**ห้ามจบงานโดยไม่บอก** — user ต้องรู้ว่ายังมีงาน backend ค้างหรือไม่

## Build Commands
```bash
# Web DEV
flutter build web -t lib/main_bcaidev.dart --release

# Web UAT
flutter build web -t lib/main_bcaiuat.dart --release

# Web PROD
flutter build web -t lib/main_bcaiprod.dart --release

# Windows DEV
flutter build windows -t lib/main_bcaidev.dart --release

# Run DEV
flutter run -t lib/main_bcaidev.dart
```

## BC AI Skills (Project-Specific AI Skills)

**อ่านและปฏิบัติตาม skills ใน `D:\bcdev\bcaiskill\bcaiskill\`:**

| Skill | คำอธิบาย | อ่านเมื่อ |
|-------|---------|----------|
| `bcai-core` | Business rules, document flows, DB schema (PM-owned) | ทุกครั้งก่อนเริ่มงาน |
| `bcai-flutter-frontend` | Flutter patterns (BLoC, Dio, widgets) | ทำงานฝั่ง Flutter |
| `bcai-go-backend` | Go backend patterns (Echo, PostgreSQL, MCP tools) | ทำงานฝั่ง Backend |

**บังคับอ่านก่อนเริ่มงาน:**
1. `bcai-core/SKILL.md` + `bcai-core/references/business-rules.md` — กฏ + checklist
2. ถ้าทำ Flutter → อ่าน `bcai-flutter-frontend/SKILL.md`
3. ถ้าทำ Backend → อ่าน `bcai-go-backend/SKILL.md`
4. ถ้าทำ MCP tools → อ่าน `bcai-go-backend/references/mcp-tools.md`

**References ที่สำคัญ:**
- `bcai-core/references/document-flows.md` — State machine ของทุก document type
- `bcai-core/references/cross-system-workflow.md` — MCP-first, API spec format
- `bcai-core/references/schema-migration.md` — กฎเมื่อ DB schema เปลี่ยน
- `bcai-core/references/database-schema.md` — โครงสร้าง tables ทั้งหมด
- `bcai-flutter-frontend/references/bloc-pattern.md` — BLoC code templates
- `bcai-go-backend/references/mcp-tools.md` — วิธีสร้าง MCP tool (35+ tools)

## Jead Skill (Global AI Rules)

**อ่านและปฏิบัติตาม rules ทั้งหมดใน `D:\bcdev\jead-skill\`:**

- `CLAUDE.md` — global rules สำหรับ AI ทุกตัว
- `identity.md` — ตัวตนและสไตล์ของ Jead (ต้องปฏิบัติตามเสมอ)
- `rules/*.md` — กฏการทำงาน (coding-style, mcp-first, erp-conventions, security)
- `skills/*/SKILL.md` — slash commands ที่ใช้ได้ (/api-search, /api-spec, /model-gen, /enum-list, /mcp-check, /master-data)

**Auto-Update Rule:** เมื่อเรียนรู้สิ่งใหม่จาก Jead (กฏใหม่, preferences, patterns) → ต้อง update ไฟล์ใน `D:\bcdev\jead-skill\` ทันที
