---
source: mongo-shop-model.go
tags: [datamodel, general-type]
---

# MongoShopModel

โครงสร้างข้อมูลร้านค้า/กิจการจาก MongoDB ประกอบด้วยรหัส holding และชื่อร้านแบบหลายภาษา

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | guidfixed | guidfixed | รหัสกิจการ (holding) — ใช้ค่า guidfixed ของร้าน |
| Names | [][[languageNameModel]] | names | name | ชื่อร้านหลายภาษา |

## ความสัมพันธ์

- Embedded: [[languageNameModel]] — รายการชื่อแยกตามภาษา
- HoldingCode: อ้างอิงรหัสกิจการ (holding/tenant) ของระบบ — map จาก field `guidfixed` ใน MongoDB
