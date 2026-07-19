---
source: mongo-debtor-model.go
tags: [datamodel, general-type]
---

# ProcessMongoDebtorModel

โมเดลข้อมูลลูกหนี้ (Debtor) ที่ใช้ในขั้นตอน process ข้อมูลจาก MongoDB ระบุตัวตนด้วย HoldingCode + GuidFixed + Code พร้อมข้อมูลภาษี ประเภทลูกค้า เครดิต และที่อยู่สำหรับวางบิล

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | holdingcode | holdingcode | รหัสกลุ่มกิจการ (tenant) |
| GuidFixed | string | guidfixed | guidfixed | GUID ประจำ record |
| Code | string | code | code | รหัสลูกหนี้ |
| Names | [][[ProcessMongoDebtorNameModel]] | names | names | รายการชื่อลูกหนี้ (รองรับหลายชื่อ/หลายภาษา) |
| TaxID | string | taxid | taxid | เลขประจำตัวผู้เสียภาษี |
| PersonalType | int8 | personaltype | personaltype | ประเภทบุคคล (บุคคลธรรมดา/นิติบุคคล) |
| CustomerType | int32 | customertype | customertype | ประเภทลูกค้า |
| BranchNumber | string | branchnumber | branchnumber | เลขที่สาขา (สาขาผู้เสียภาษี) |
| FundCode | string | fundcode | fundcode | รหัสกองทุน/แหล่งเงิน |
| CreditDay | int32 | creditday | creditday | จำนวนวันเครดิต |
| Email | string | email | email | อีเมล |
| AddressBilling | map[string]any | addressforbilling | addressforbilling | ที่อยู่สำหรับวางบิล (โครงสร้างอิสระ) |

## ความสัมพันธ์

- ฝัง (embed) `[]` [[ProcessMongoDebtorNameModel]] ใน field `Names`
- `HoldingCode` — อ้างอิงกลุ่มกิจการ/tenant ของข้อมูล
- `Code` — รหัสอ้างอิงลูกหนี้ในระบบ
