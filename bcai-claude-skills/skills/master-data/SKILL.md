---
name: master-data
description: จัดการ master data (units, barcodes, product groups, product categories, creditors, debtors) ผ่าน MCP tools — ใช้เมื่อต้องการดู/สร้าง/แก้ไข/ลบข้อมูลหลัก, ค้นหาสินค้า, ตรวจสอบ master data ก่อนสร้าง feature, import ข้อมูลเข้าระบบ, หรือเมื่อพูดถึงหน่วยนับ, บาร์โค้ด, เจ้าหนี้, ลูกหนี้, กลุ่มสินค้า, หมวดสินค้า
user-invocable: true
---

# Master Data — จัดการข้อมูลหลัก

## วิธีใช้
`/master-data list <type>` — แสดงรายการ
`/master-data search <keyword>` — ค้นหาสินค้า
`/master-data create <type> <data>` — สร้างใหม่
`/master-data update <type> <data>` — แก้ไข
`/master-data delete <type> <guid>` — ลบ

## ประเภท Master Data ที่รองรับ

| Type | MCP Tools | คำอธิบาย |
|------|-----------|---------|
| `units` | list/create/update/delete_unit(s) | หน่วยนับ (ชิ้น, กล่อง, โหล) |
| `barcodes` | list/create/update/delete_barcode(s) | บาร์โค้ดสินค้า |
| `product-groups` | list/create/update/delete_product_group(s) | กลุ่มสินค้า |
| `product-categories` | list/create/update/delete_product_category(ies) | หมวดสินค้า |
| `creditors` | list/create/update/delete_creditor(s) | เจ้าหนี้ |
| `debtors` | list/create/update/delete_debtor(s) | ลูกหนี้ |
| `products` | search_products | ค้นหาสินค้า (read-only) |

## ขั้นตอนการทำงาน

### List — ดูรายการ
```
/master-data list units
→ เรียก MCP tool: list_units
→ แสดงตาราง: code, name, description
```

### Search — ค้นหาสินค้า
```
/master-data search MAKITA
→ เรียก MCP tool: search_products
→ Parameters: { keyword: "MAKITA" }
→ แสดงผล: code, name, price, stock
```

### Create — สร้างใหม่
```
/master-data create unit { "code": "DOZ", "name1": "โหล", "name2": "Dozen" }
→ เรียก MCP tool: create_unit
→ ยืนยันผลลัพธ์
```

### Batch Create — สร้างหลายรายการ
```
/master-data create units [
  { "code": "DOZ", "name1": "โหล" },
  { "code": "BOX", "name1": "กล่อง" }
]
→ เรียก MCP tool: create_units (batch)
```

### ดู Schema ก่อนสร้าง
ก่อน create/update ให้ตรวจ schema:
```
Tool: get_barcode_schema / get_unit_schema / get_creditor_schema / ...
→ แสดง required fields, data types, validation rules
```

## MCP Tools ทั้งหมด

### Units (หน่วยนับ)
| Tool | หน้าที่ |
|------|--------|
| `list_units` | ดูรายการหน่วยนับ |
| `create_unit` | สร้างหน่วยนับ 1 รายการ |
| `create_units` | สร้างหน่วยนับหลายรายการ |
| `update_unit` | แก้ไขหน่วยนับ |
| `delete_unit` | ลบหน่วยนับ 1 รายการ |
| `delete_units` | ลบหน่วยนับหลายรายการ |
| `get_unit_schema` | ดู schema |

### Barcodes (บาร์โค้ด)
| Tool | หน้าที่ |
|------|--------|
| `list_barcodes` | ดูรายการบาร์โค้ด |
| `create_barcode` | สร้างบาร์โค้ด |
| `create_barcodes` | สร้างบาร์โค้ดหลายรายการ |
| `update_barcode` | แก้ไขบาร์โค้ด |
| `delete_barcode` | ลบบาร์โค้ด |
| `delete_barcodes` | ลบบาร์โค้ดหลายรายการ |
| `get_barcode_schema` | ดู schema |
| `get_ref_barcodes` | ดูบาร์โค้ดอ้างอิง (reference chain) |
| `set_ref_barcode` | ตั้งค่าบาร์โค้ดอ้างอิง |
| `create_multi_unit_barcode` | สร้างสินค้าหลายหน่วยนับ |

### Product Groups (กลุ่มสินค้า)
| Tool | หน้าที่ |
|------|--------|
| `list_product_groups` | ดูรายการกลุ่มสินค้า |
| `create_product_group` | สร้างกลุ่มสินค้า |
| `create_product_groups` | สร้างหลายรายการ |
| `update_product_group` | แก้ไข |
| `delete_product_group` | ลบ |
| `delete_product_groups` | ลบหลายรายการ |
| `get_product_group_schema` | ดู schema |

### Product Categories (หมวดสินค้า)
| Tool | หน้าที่ |
|------|--------|
| `list_product_categories` | ดูรายการหมวดสินค้า |
| `create_product_category` | สร้างหมวดสินค้า |
| `create_product_categories` | สร้างหลายรายการ |
| `update_product_category` | แก้ไข |
| `delete_product_category` | ลบ |
| `delete_product_categories` | ลบหลายรายการ |
| `get_product_category_schema` | ดู schema |

### Creditors (เจ้าหนี้)
| Tool | หน้าที่ |
|------|--------|
| `list_creditors` | ดูรายการเจ้าหนี้ |
| `create_creditor` | สร้างเจ้าหนี้ |
| `create_creditors` | สร้างหลายรายการ |
| `update_creditor` | แก้ไข |
| `delete_creditor` | ลบ |
| `delete_creditors` | ลบหลายรายการ |
| `get_creditor_schema` | ดู schema |

### Debtors (ลูกหนี้)
| Tool | หน้าที่ |
|------|--------|
| `list_debtors` | ดูรายการลูกหนี้ |
| `create_debtor` | สร้างลูกหนี้ |
| `create_debtors` | สร้างหลายรายการ |
| `update_debtor` | แก้ไข |
| `delete_debtor` | ลบ |
| `delete_debtors` | ลบหลายรายการ |
| `get_debtor_schema` | ดู schema |

### Products (สินค้า — read only)
| Tool | หน้าที่ |
|------|--------|
| `search_products` | ค้นหาสินค้า + ยอด stock |

## ตัวอย่าง
```
/master-data list units           → หน่วยนับทั้งหมด
/master-data list barcodes        → บาร์โค้ดทั้งหมด
/master-data list creditors       → เจ้าหนี้ทั้งหมด
/master-data search MAKITA        → ค้นหาสินค้า MAKITA
/master-data search สว่าน         → ค้นหาสินค้าภาษาไทย
```

## หมายเหตุ
- Create/Update/Delete ต้องมี API key ที่มีสิทธิ์เขียน
- `search_products` รองรับภาษาไทย
- Batch operations (create_units, delete_barcodes) รับ array ได้
