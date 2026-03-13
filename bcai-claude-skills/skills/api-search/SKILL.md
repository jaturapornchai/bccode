---
name: api-search
description: ค้นหา API endpoints จาก backend ผ่าน MCP — ใช้เมื่อต้องการรู้ว่า backend มี API อะไรบ้าง, หา endpoint สำหรับ feature, ตรวจสอบ API ก่อนเขียน frontend, ถามว่า "มี API สำหรับ ... ไหม", หรือเมื่อพูดถึง endpoint, route, handler ใดๆ ในระบบ. ใช้ skill นี้ก่อน `/api-spec` เสมอถ้าไม่รู้ path
user-invocable: true
---

# API Search — ค้นหา API Endpoint

## วิธีใช้
`/api-search <keyword>` — ค้นหา API ที่มี keyword ใน path หรือ description
`/api-search` — แสดง API ทั้งหมด

## ขั้นตอนการทำงาน

### 1. ค้นหา API endpoints
เรียก MCP tool `list_api_endpoints` ผ่าน MCP dev endpoint:
```
Tool: list_api_endpoints
Parameters: (ไม่มี — return ทั้งหมด)
```

### 2. Filter ตาม keyword
- ค้นหาใน path, method, description, tags
- Case-insensitive matching
- ถ้าไม่มี keyword → แสดงทั้งหมดแบ่งตาม category

### 3. แสดงผล
แสดงเป็นตาราง:
| Method | Path | Description |
|--------|------|-------------|
| GET | /goapi/api/health | Health check |
| POST | /goapi/api/v1/chatbot/chat-agent | AI chatbot |

### 4. แนะนำขั้นตอนถัดไป
- ถ้าต้องการรายละเอียด → `/api-spec <method> <path>`
- ถ้าต้องการดูตัวอย่าง → ใช้ MCP tool `get_api_example`
- ถ้าต้องการดู model → `/model-gen <model_name>`

## ตัวอย่าง
```
/api-search chatbot     → หา API เกี่ยวกับ chatbot
/api-search product     → หา API เกี่ยวกับสินค้า
/api-search mcp         → หา MCP endpoints
/api-search stock       → หา API เกี่ยวกับสต็อก
/api-search sale        → หา API เกี่ยวกับการขาย
```

## MCP Tools ที่ใช้
| Tool | Endpoint | หน้าที่ |
|------|----------|--------|
| `list_api_endpoints` | Dev only | ดูรายการ API ทั้งหมด |
| `get_api_spec` | Dev only | ดูรายละเอียด API |
| `get_api_example` | Dev only | ดูตัวอย่าง request/response |

## หมายเหตุ
- ใช้ได้เฉพาะเมื่อ MCP server ทำงานอยู่ (`GET http://localhost:9090/goapi/mcp/dev/health`)
- `list_api_endpoints` อยู่ใน Dev endpoint เท่านั้น (ไม่ใช่ General)
- URL base: `http://localhost:9090/goapi`
