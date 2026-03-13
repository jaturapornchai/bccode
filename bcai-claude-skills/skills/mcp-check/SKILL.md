---
name: mcp-check
description: ตรวจสอบสถานะ MCP server, tools ที่ใช้ได้, API key — ใช้เมื่อต้องการรู้ว่า MCP ทำงานอยู่ไหม, มี tools อะไรบ้าง, ตรวจสอบ connection ก่อนเริ่มงาน, debug MCP ที่ไม่ตอบ, หรือเมื่อ MCP tool call ล้มเหลว. ใช้เป็นขั้นตอนแรกเมื่อ MCP มีปัญหา
user-invocable: true
---

# MCP Check — ตรวจสอบ MCP Status

## วิธีใช้
`/mcp-check` — ตรวจสอบ MCP server ทั้ง General และ Dev
`/mcp-check tools` — แสดงรายการ tools ที่ใช้ได้

## ขั้นตอนการทำงาน

### 1. ตรวจ Health
```bash
# General endpoint
GET http://localhost:9090/goapi/mcp/health

# Dev endpoint
GET http://localhost:9090/goapi/mcp/dev/health
```

### 2. ตรวจ Tools ที่ใช้ได้
```bash
# General tools (business data — safe for AI agents)
GET http://localhost:9090/goapi/mcp/tools

# Dev tools (ทั้งหมด — รวม database access)
GET http://localhost:9090/goapi/mcp/dev/tools
```

### 3. แสดงผล
```
## MCP Server Status

| Endpoint | Status | URL |
|----------|--------|-----|
| General | ✅ Online | http://localhost:9090/goapi/mcp/health |
| Dev | ✅ Online | http://localhost:9090/goapi/mcp/dev/health |

## Available Tools

### General Tools (25 ตัว — ปลอดภัยสำหรับ AI)
| Category | จำนวน | ตัวอย่าง |
|----------|--------|---------|
| Sales | 5 | get_daily_sales, get_top_selling_products |
| Dashboard | 3 | get_dashboard_kpis, get_business_health |
| Inventory | 5 | get_inventory_value, get_low_stock_alerts |
| Finance | 5 | get_profit_analysis, get_cash_flow |
| Customer | 3 | get_top_customers, get_customer_segments |
| Master Data | 4 | list_units, search_products |

### Dev-Only Tools (16 ตัว — database access)
| Category | จำนวน | ตัวอย่าง |
|----------|--------|---------|
| PostgreSQL | 4 | get_database_schema, execute_query |
| MongoDB | 3 | query_mongodb, aggregate_mongodb |
| ClickHouse | 3 | query_clickhouse, list_clickhouse_tables |
| API/Schema | 6 | list_api_endpoints, get_model_schema |
```

### 4. ตรวจ API Key (ถ้ามี)
- API Key format: `bc_live_{32chars}`
- Auth: Header `X-API-Key` หรือ Query `?api_key=`

## Troubleshooting

### MCP ไม่ตอบ
1. ตรวจว่า backend server รันอยู่: `GET http://localhost:9090/goapi/api/health`
2. ตรวจ port — ใช้ **9090** (ไม่ใช่ 8888 หรือ 9091)
3. GoAPI รวมเข้า MainAPI แล้ว — ใช้ prefix `/goapi` ทุกครั้ง

### MCP tool ไม่ทำงาน
1. ตรวจว่าใช้ถูก endpoint (General vs Dev)
2. Dev tools ใช้ได้เฉพาะ `/goapi/mcp/dev/` endpoint
3. ตรวจ API key permission

## Endpoints สรุป
| Endpoint | Port | Path | ใช้สำหรับ |
|----------|------|------|----------|
| Health (General) | 9090 | `/goapi/mcp/health` | ตรวจสถานะ |
| Health (Dev) | 9090 | `/goapi/mcp/dev/health` | ตรวจสถานะ dev |
| Tools (General) | 9090 | `/goapi/mcp/tools` | ดู business tools |
| Tools (Dev) | 9090 | `/goapi/mcp/dev/tools` | ดูทุก tools |
| SSE (General) | 9090 | `/goapi/mcp/sse` | Claude Desktop/AI |
| SSE (Dev) | 9090 | `/goapi/mcp/dev/sse` | Dev AI clients |
