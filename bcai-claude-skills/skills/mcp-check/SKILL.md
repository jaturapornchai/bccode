---
name: mcp-check
description: ตรวจสอบสถานะ MCP server และ tools ที่ใช้ได้
user_invocable: true
---

# MCP Check — ตรวจสอบ MCP Status

## วิธีใช้
`/mcp-check`

## สิ่งที่ทำ
1. ตรวจสอบ MCP server health (`GET /goapi/mcp/health`)
2. ดูรายการ tools ทั้งหมดที่ใช้ได้
3. แสดงจำนวน tools แยกตาม category
4. ตรวจสอบ API key ว่าใช้ได้หรือไม่

## Endpoints ที่ตรวจ
- Health: `http://localhost:8888/goapi/mcp/health`
- Tools: `http://localhost:8888/goapi/mcp/tools`
