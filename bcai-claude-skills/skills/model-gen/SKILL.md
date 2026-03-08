---
name: model-gen
description: ดู data model schema จาก backend
user_invocable: true
---

# Model Schema — ดูโครงสร้าง Data Model

## วิธีใช้
`/model-gen <model_name>`

## สิ่งที่ทำ
1. เรียก MCP tool `get_model_schema` กับชื่อ model
2. แสดง fields, types, json tags, validation rules
3. สร้าง Dart class skeleton สำหรับ frontend (ถ้าขอ)

## ตัวอย่าง
```
/model-gen ProductDoc
/model-gen SaleInvoiceDoc
/model-gen UnitDoc
```
