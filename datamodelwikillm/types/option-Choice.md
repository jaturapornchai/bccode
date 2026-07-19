---
source: backend/internal/product/option/models/option.go
tags: [datamodel, general-type]
---

# Choice

โครงสร้าง `Choice` จากโมดูล mainapi `option` มีฟิลด์ตามซอร์ส `option.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| OptCode | string | - | - | ค่าของ OptCode ตามฟิลด์ `OptCode` ในซอร์ส |
| Code | string | code | code | รหัสรายการ |
| SuggestCode | string | suggestcode,omitempty | suggestcode,omitempty | ค่าของ SuggestCode ตามฟิลด์ `suggestcode` ในซอร์ส |
| Barcode | string | barcode | barcode | บาร์โค้ดสินค้า |
| Price | float64 | price | price | ราคา |
| Qty | float64 | qty | qty | จำนวน |
| QtyMax | float64 | qtymax | qtymax | ค่าของ QtyMax ตามฟิลด์ `qtymax` ในซอร์ส |
| Name | [[Name\|models.Name]] | inline | - | โครงสร้างฝังแบบ inline |
| Names | *[][[NameX\|models.NameX]] | names | names | รายการชื่อหลายภาษา |
| ItemUnit | string | itemunit | itemunit,omitempty | ค่าของ ItemUnit ตามฟิลด์ `itemunit` ในซอร์ส |
| Selected | bool | selected | selected | ค่าของ Selected ตามฟิลด์ `selected` ในซอร์ส |
| Default | bool | default | default | ค่าของ Default ตามฟิลด์ `default` ในซอร์ส |
| IncudeOptions | [][[option-IncudeChoice\|IncudeChoice]] | choicedetails,omitempty | choicedetails,omitempty | ค่าของ IncudeOptions ตามฟิลด์ `choicedetails` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[Name]], [[NameX]], [[option-IncudeChoice|IncudeChoice]]
