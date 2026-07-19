---
source: backend/internal/goapi/models/global-model.go
tags: [datamodel, general-type]
---

# productOrderOptionModel

โครงสร้างตัวเลือกสินค้า (option เช่น สี ขนาด) เก็บรหัส GUID ชื่อหลายภาษา สถานะการตัดสต๊อก และรายการรายละเอียดตัวเลือก

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Guidcode | string | - | guidcode | รหัส GUID ของตัวเลือก |
| Names | [][[languageNameModel]] | - | names | ชื่อตัวเลือกหลายภาษา |
| Isstock | bool | - | isstock | ตัวเลือกนี้มีผลต่อสต๊อกหรือไม่ |
| Optiondetails | [][[productOrderOptionDetailModel]] | - | optiondetails | รายการรายละเอียดของตัวเลือก (เช่น ขาว/ดำ) |

## ความสัมพันธ์
- ฝัง [[languageNameModel]] (Names), [[productOrderOptionDetailModel]] (Optiondetails)
- ถูกฝังใน [[productOrderModel]] (Options)
