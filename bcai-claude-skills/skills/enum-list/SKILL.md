---
name: enum-list
description: ดูรายการ enums ที่ backend กำหนด
user_invocable: true
---

# Enum Catalog — ดูรายการ Enums

## วิธีใช้
`/enum-list [keyword]`

## สิ่งที่ทำ
1. เรียก MCP tool `get_enum_catalog`
2. แสดงรายการ enums ทั้งหมด (หรือ filter ตาม keyword)
3. แสดง values + descriptions ของแต่ละ enum

## ตัวอย่าง
```
/enum-list
/enum-list payment
/enum-list doc_type
```
