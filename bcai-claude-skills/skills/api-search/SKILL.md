---
name: api-search
description: ค้นหา API endpoints จาก backend ผ่าน MCP
user_invocable: true
---

# API Search — ค้นหา API Endpoint

## วิธีใช้
`/api-search <keyword>`

## สิ่งที่ทำ
1. เรียก MCP tool `get_api_catalog` เพื่อดูรายการ API ทั้งหมด
2. Filter ตาม keyword ที่ให้มา
3. แสดงผลเป็นตาราง: method, path, description

## ตัวอย่าง
```
/api-search chatbot
/api-search mcp
/api-search stock
```

## หมายเหตุ
- ใช้ได้เฉพาะเมื่อ MCP server ทำงานอยู่
- ต้องมี MCP API key ใน config
