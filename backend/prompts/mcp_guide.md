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
  -d '{"name":"frontend-dev","description":"For frontend AI","holdingcode":"YOUR_HOLDING_CODE"}'
```

Response จะได้ `apikey` กลับมา — เก็บไว้ใช้ทุก request

### holdingcode ไม่ต้องส่ง

API key มี `holdingcode` ฝังอยู่แล้ว — **ไม่ต้องส่ง holdingcode ในทุก request**

Backend auto-inject holdingcode จาก API key ให้ทุก tool

---

## วิธีเรียก

```json
{
  "tool": "ชื่อ tool",
  "params": {
    // parameters ตาม tool (ไม่ต้องมี holdingcode)
  }
}
```

### Response format

```json
{
  "success": true,
  "data": { ... },
  "timestamp": "2025-02-25T10:00:00Z",
  "tool": "toolname"
}
```

Error:
```json
{
  "success": false,
  "error": "error message",
  "errorcode": "ERROR_CODE",
  "timestamp": "...",
  "tool": "toolname"
}
```

---

## ตัวอย่างการเรียก

### ค้นสินค้า
```bash
curl -X POST {BASE_URL}/goapi/mcp/invoke \
  -H "Content-Type: application/json" \
  -H "X-API-Key: bc_live_xxx" \
  -d '{"tool":"searchproducts","params":{"keyword":"กาแฟ"}}'
```

### ยอดขายวันนี้
```bash
curl -X POST {BASE_URL}/goapi/mcp/invoke \
  -H "Content-Type: application/json" \
  -H "X-API-Key: bc_live_xxx" \
  -d '{"tool":"getdailysales","params":{"date":"2025-02-25"}}'
```

### ดู API spec ของ /login
```bash
curl -X POST {BASE_URL}/goapi/mcp/invoke \
  -H "Content-Type: application/json" \
  -H "X-API-Key: bc_live_xxx" \
  -d '{"tool":"getapispec","params":{"path":"/login"}}'
```

### ดู enum values
```bash
curl -X POST {BASE_URL}/goapi/mcp/invoke \
  -H "Content-Type: application/json" \
  -H "X-API-Key: bc_live_xxx" \
  -d '{"tool":"listenums","params":{"keyword":"transflag"}}'
```

### ดู Dart model schema
```bash
curl -X POST {BASE_URL}/goapi/mcp/invoke \
  -H "Content-Type: application/json" \
  -H "X-API-Key: bc_live_xxx" \
  -d '{"tool":"getmodelschema","params":{"model":"Shop"}}'
```

---

## Tools ทั้งหมด (35 ตัว)

### ค้นสินค้า

| Tool | คำอธิบาย | Parameters |
|------|----------|------------|
| `searchproducts` | ค้นสินค้า Thai full-text search + stock balance | `keyword` (required), `limit`, `whcode`, `locationcode`, `includebalance` |

### ยอดขาย (Sales)

| Tool | คำอธิบาย | Parameters |
|------|----------|------------|
| `getdailysales` | ยอดขายรายวัน | `date` (required, YYYY-MM-DD), `branchcode` |
| `getsalesbydaterange` | ยอดขายตามช่วงวัน | `fromdate`, `todate` (required), `groupby` (day/week/month), `branchcode` |
| `gettopsellingproducts` | สินค้าขายดี | `fromdate`, `todate` (required), `limit`, `branchcode` |
| `getsalesbyseller` | ยอดขายตามช่องทาง | `fromdate`, `todate` (required) |
| `getmonthlysummary` | สรุปยอดขายรายเดือน | `year`, `month` (required) |

### Dashboard

| Tool | คำอธิบาย | Parameters |
|------|----------|------------|
| `getdashboardkpis` | KPI dashboard สำหรับ CEO | `period` (today/thisweek/thismonth/thisyear) |
| `getbusinesshealth` | คะแนนสุขภาพธุรกิจ (0-100) | - |

### การเงิน (Financial)

| Tool | คำอธิบาย | Parameters |
|------|----------|------------|
| `getprofitanalysis` | วิเคราะห์กำไร | `fromdate`, `todate` (required) |
| `getaccountsreceivable` | ลูกหนี้ค้างชำระ (aging) | `limit` |
| `getaccountspayable` | เจ้าหนี้ค้างชำระ (aging) | `limit` |
| `getcashflow` | กระแสเงินสด | `fromdate`, `todate` (required) |

### สต็อก (Inventory)

| Tool | คำอธิบาย | Parameters |
|------|----------|------------|
| `getinventoryvalue` | มูลค่าสินค้าคงคลัง | `whcode` |
| `getlowstockalerts` | แจ้งเตือนสต็อกต่ำ | `threshold`, `limit` |
| `getdeadstock` | สินค้าไม่เคลื่อนไหว | `daysnomovement`, `limit` |
| `getinventoryturnover` | อัตราหมุนเวียนสินค้า | `fromdate`, `todate` (required), `limit` |

### ลูกค้า (Customer)

| Tool | คำอธิบาย | Parameters |
|------|----------|------------|
| `gettopcustomers` | ลูกค้า top revenue | `fromdate`, `todate` (required), `limit` |
| `getcustomergrowth` | การเติบโตของลูกค้า | `fromdate`, `todate` (required) |
| `getcustomersegments` | วิเคราะห์ RFM segmentation | - |

### เปรียบเทียบ (Comparison)

| Tool | คำอธิบาย | Parameters |
|------|----------|------------|
| `getyoycomparison` | เทียบปีต่อปี | `year`, `month` |
| `getmomcomparison` | เทียบเดือนต่อเดือน | `year`, `month` |

### Database Query

| Tool | คำอธิบาย | Parameters |
|------|----------|------------|
| `getdatabaseschema` | ดูโครงสร้าง tables | `tablename` |
| `executequery` | รัน SQL SELECT (readonly) | `query` (required), `limit` |
| `gettablesample` | ดูตัวอย่างข้อมูล | `tablename` (required), `limit` |

### MongoDB

| Tool | คำอธิบาย | Parameters |
|------|----------|------------|
| `querymongodb` | Query collection | `collection` (required), `filter`, `limit`, `database` |
| `listmongodbcollections` | ดู collections ทั้งหมด | `database` |
| `aggregatemongodb` | Aggregation pipeline | `collection`, `pipeline` (required), `limit`, `database` |

### ClickHouse (OLAP)

| Tool | คำอธิบาย | Parameters |
|------|----------|------------|
| `queryclickhouse` | รัน SQL SELECT (readonly) | `query` (required), `limit`, `database` |
| `listclickhousetables` | ดู tables ทั้งหมด | `database` |

### API Reference (สำหรับ AI เขียนโค้ด)

| Tool | คำอธิบาย | Parameters |
|------|----------|------------|
| `listapiendpoints` | ดู API ทั้งหมด (1,427 endpoints) | `keyword`, `method`, `category`, `source` (goapi/mainapi), `limit` |
| `getapispec` | ดู spec ละเอียด (params, request/response schema) | `path` (required), `method` |
| `getapiexample` | ดูตัวอย่าง curl + Dart + TypeScript | `path` (required), `method` |

### Enum & Model (สำหรับ AI เขียนโค้ด)

| Tool | คำอธิบาย | Parameters |
|------|----------|------------|
| `listenums` | ดู enum values ทั้งหมด (30 groups) | `category`, `keyword` |
| `getmodelschema` | ดู struct → JSON schema + Dart class + TS interface | `model`, `category`, `keyword` |

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

## Claude Code / Project-local Usage

MCP tools เรียกผ่าน curl/HTTP ได้โดยตรง และให้ใช้เอกสาร project-local เป็นหลัก:

- Backend rules: `D:\bcdev\backend\CLAUDE.md`
- Frontend rules: `D:\bcdev\frontend\bcaiaccount\CLAUDE.md`
- API request prompts: `D:\bcdev\backend\prompts\api_requests\`
- MCP guide: ไฟล์นี้

### Common MCP workflows

| Need | MCP Tool |
|------|----------|
| ค้นหา API endpoints | `listapiendpoints` |
| ดู request/response spec | `getapispec`, `getapiexample` |
| สร้าง model จาก schema | `getmodelschema`, `getdatabaseschema`, `gettablesample` |
| ดู enum values/constants | `listenums` |
| ตรวจ MCP server | `GET /goapi/mcp/health`, `GET /goapi/mcp/tools` |

### Workflow แนะนำ

1. เริ่ม dev session → ตรวจ `GET {BASE_URL}/goapi/mcp/health`
2. หา API → เรียก `listapiendpoints`
3. ดู spec ละเอียด → เรียก `getapispec`
4. สร้าง model → เรียก `getmodelschema`
5. ดู enum values → เรียก `listenums`

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
- Data isolation: ทุก query filter by `holdingcode` จาก API key อัตโนมัติ
- Rate limit: ยังไม่มี (อนาคตจะเพิ่ม)
- ใช้ project-local docs/source เป็นหลัก ไม่พึ่ง shared local skills folder
