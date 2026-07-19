---
source: backend/internal/goapi/models/global-model.go
tags: [datamodel, general-type]
---

# productOrderOptionDetailModel

โครงสร้างรายละเอียดของตัวเลือกสินค้า (เช่น ค่า "ขาว" ของตัวเลือก "สี") เก็บชื่อหลายภาษา รูปภาพ และตัวเลือกอื่นที่เชื่อมโยงกัน มีฟิลด์สถานะภายใน (Selected, Isenable) ที่ไม่ส่งออกทาง JSON

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Guidcode | string | - | guidcode | รหัส GUID ของรายละเอียดตัวเลือก |
| Names | [][[languageNameModel]] | - | names | ชื่อรายละเอียดตัวเลือกหลายภาษา |
| Image | string | - | image | รูปภาพประกอบตัวเลือก |
| Includeoptions | [][[productOrderOptionDetailIncludeModel]] | - | includeoptions | ตัวเลือกอื่นที่เชื่อมโยง/รวมกับค่านี้ |
| Selected | bool | - | - (json:"-") | สถานะถูกเลือก (ใช้ภายใน ไม่ serialize) |
| Isenable | bool | - | - (json:"-") | สถานะเปิดใช้งาน (ใช้ภายใน ไม่ serialize) |

## ความสัมพันธ์
- ฝัง [[languageNameModel]] (Names), [[productOrderOptionDetailIncludeModel]] (Includeoptions)
- ถูกฝังใน [[productOrderOptionModel]] (Optiondetails)
