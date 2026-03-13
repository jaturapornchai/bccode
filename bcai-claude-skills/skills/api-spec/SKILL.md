---
name: api-spec
description: ดู API specification (request/response/parameters) ของ endpoint — ใช้เมื่อต้องการรู้ว่า API รับ-ส่งข้อมูลอะไร, ดูตัวอย่าง request/response, สร้าง Dart model จาก response, ทำ API integration, เขียน API client, หรือ debug API ที่ส่งค่าไม่ตรง. ถ้าไม่รู้ path → ใช้ `/api-search` ก่อน
user-invocable: true
---

# API Spec — ดูรายละเอียด API

## วิธีใช้
`/api-spec <method> <path>` — ดู spec ของ endpoint เฉพาะ
`/api-spec <keyword>` — ค้นหาก่อนแล้วแสดง spec

## ขั้นตอนการทำงาน

### 1. ดู API Specification
เรียก MCP tool `get_api_spec`:
```
Tool: get_api_spec
Parameters:
  - method: "GET" | "POST" | "PUT" | "DELETE"
  - path: "/goapi/api/v1/..."
```

### 2. ดูตัวอย่าง Request/Response
เรียก MCP tool `get_api_example`:
```
Tool: get_api_example
Parameters:
  - method: "POST"
  - path: "/goapi/api/v1/chatbot/chat-agent"
```

### 3. แสดงข้อมูล
```
## Endpoint: POST /goapi/api/v1/chatbot/chat-agent

### Request Headers
- Authorization: Bearer <JWT token>
- Content-Type: application/json

### Request Body
{
  "message": "string — ข้อความจาก user",
  "session_id": "string — session ID (optional)"
}

### Response (200 OK)
{
  "success": true,
  "data": {
    "response": "string — คำตอบจาก AI",
    "tools_used": ["string"] — tools ที่ AI เรียกใช้
  }
}

### curl Example
curl -X POST http://localhost:9090/goapi/api/v1/chatbot/chat-agent \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"message": "ยอดขายวันนี้"}'
```

### 4. แนะนำขั้นตอนถัดไป
- สร้าง Dart model → `/model-gen <model_name>`
- ดู enum values → `/enum-list <keyword>`
- หา API อื่น → `/api-search <keyword>`

## ตัวอย่าง
```
/api-spec POST /goapi/api/v1/chatbot/chat-agent
/api-spec GET /goapi/api/health
/api-spec chatbot    → ค้นหาก่อนแล้วแสดง spec
```

## MCP Tools ที่ใช้
| Tool | หน้าที่ |
|------|--------|
| `get_api_spec` | ดู specification (params, body, response) |
| `get_api_example` | ดูตัวอย่าง request/response จริง |
| `list_api_endpoints` | ค้นหา endpoint (ถ้าไม่รู้ path) |

## หมายเหตุ
- ถ้าไม่รู้ path → ใช้ `/api-search` ก่อน
- Tools เหล่านี้อยู่ใน MCP Dev endpoint เท่านั้น
- URL base: `http://localhost:9090/goapi`
