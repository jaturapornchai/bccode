---
name: master-data
description: จัดการ master data (units, products) ผ่าน MCP tools
user_invocable: true
---

# Master Data — จัดการข้อมูลหลัก

## วิธีใช้
`/master-data <action> <type> [params]`

## Actions
- `list` — แสดงรายการ (เช่น `/master-data list units`)
- `search` — ค้นหา (เช่น `/master-data search products MAKITA`)

## สิ่งที่ทำ
1. เรียก MCP tools ที่เหมาะสม (`list_units`, `search_products`)
2. แสดงผลเป็นตาราง
3. ถ้าเป็น write operation → แจ้งว่าต้องใช้ Developer API key

## MCP Tools ที่ใช้
- `list_units` / `search_units` / `create_unit` / `update_unit` / `delete_unit`
- `search_products` / `list_products`
