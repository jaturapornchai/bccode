---
name: api-spec
description: ดู API specification (request/response) ของ endpoint
user_invocable: true
---

# API Spec — ดูรายละเอียด API

## วิธีใช้
`/api-spec <method> <path>`

## สิ่งที่ทำ
1. เรียก MCP tool `get_api_spec` กับ method + path ที่ระบุ
2. แสดง: request body, response body, parameters, headers
3. แสดงตัวอย่าง curl command

## ตัวอย่าง
```
/api-spec POST /goapi/api/v1/chatbot/chat-agent
/api-spec GET /goapi/api/health
```

## หมายเหตุ
- ถ้าไม่รู้ path → ใช้ `/api-search` ก่อน
