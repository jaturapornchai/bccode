# MCP-First Development — กฏการใช้ MCP

## หลักการ
MCP (Model Context Protocol) เป็น bridge ให้ AI ฝั่ง frontend เข้าถึง backend โดยไม่ต้องอ่าน code โดยตรง

## กฏสำคัญ

### Cross-Project Access (สำคัญ)
**AI สามารถอ่าน+แก้ source code ได้ทั้ง 2 project:**
- **Backend** `D:\bcdev\backend` — อ่าน+แก้ได้
- **Frontend** `D:\bcdev\bcaiaccount` — อ่าน+แก้ได้
- ตรวจ code ข้ามไปมาได้เสมอ เช่น ตรวจ Backend response แล้วแก้ Frontend model ให้ตรงกัน

### กฏเฉพาะ Project ทดสอบ (bcaitest01)
**เมื่อทำงานใน project `bcaitest01`:**
- **Backend** `D:\bcdev\backend` — **อ่านได้เสมอ ห้ามแก้ไข**
- **Frontend** `D:\bcdev\bcaiaccount` — **อ่านได้เสมอ ห้ามแก้ไข**
- **ต้องตรวจ backend+frontend เสมอ** เมื่อทดสอบหรือพบปัญหา
- **เมื่อพบ bug จาก MCP tools** → ตรวจ source code เพื่อหาสาเหตุ แล้ว **รายงานให้ Jead** (ห้ามแก้เอง)

### MCP ยังคงสำคัญสำหรับ:
1. **ค้นหา API ผ่าน MCP tools**: `get_api_catalog`, `get_api_spec`
2. **ค้นหา schema/enums ผ่าน MCP**: `get_model_schema`, `get_enum_catalog`
3. **ถ้าต้องการ API ใหม่**: เขียน prompt ใน `prompts/api_requests/{feature}.md`

### Backend AI (Claude Code ใน backend)
1. **อ่าน+แก้ backend code ได้**
2. **เมื่อได้ prompt จาก frontend** → สร้าง API ตาม spec + เพิ่ม MCP tool ถ้าจำเป็น
3. **MCP tools ต้องครอบคลุม**: ทุก API ที่ frontend ต้องใช้

### MCP Endpoints — แยก General vs Dev
MCP แยกเป็น 2 ชุด endpoint:

| | General (Frontend dev) | Dev (Backend dev) |
|---|---|---|
| **SSE** | `GET /goapi/mcp/sse` | `GET /goapi/mcp/dev/sse` |
| **Tools** | `GET /goapi/mcp/tools` | `GET /goapi/mcp/dev/tools` |
| **Health** | `GET /goapi/mcp/health` | `GET /goapi/mcp/dev/health` |

- **General** — เห็นเฉพาะ business tools (sales, inventory, products, barcodes, units ฯลฯ)
- **Dev** — เห็นทุก tools รวม database access, API catalog, model schema, enums
- **Dev-only tools (16 ตัว):** `get_database_schema`, `execute_query`, `get_table_sample`, `query_mongodb`, `list_mongodb_collections`, `aggregate_mongodb`, `query_clickhouse`, `list_clickhouse_tables`, `execute_pg_command`, `execute_ch_command`, `list_api_endpoints`, `get_api_spec`, `get_api_example`, `list_enums`, `get_model_schema`, `rebuild_products`
- API Keys: `POST/GET/PUT/DELETE /goapi/api/mcp/keys` (ใช้ key เดียวกันทั้ง 2 endpoint)

### MCP Config Locations
- **Claude Code**: `~/.claude.json` → `mcpServers` → bc-erp (SSE)
  - Frontend dev ใช้ `/goapi/mcp/sse`
  - Backend dev ใช้ `/goapi/mcp/dev/sse`
- **Claude Desktop**: `%APPDATA%/Claude/claude_desktop_config.json` → `mcp-remote` global
- **Frontend project**: `.mcp.json` ใน project root (ใช้ General endpoint)

## prompts/api_requests/ Pattern
เมื่อ frontend ต้องการ API ใหม่:
1. Frontend AI เขียน spec ใน `prompts/api_requests/{feature}.md`
2. Backend AI อ่าน spec แล้วสร้าง API
3. Backend AI เพิ่ม MCP tool ถ้าจำเป็น
4. Frontend AI ใช้ MCP ตรวจสอบ API ใหม่แล้ว implement
