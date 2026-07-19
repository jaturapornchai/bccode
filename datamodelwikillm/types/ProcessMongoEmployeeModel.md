---
source: mongo-employee-model.go
tags: [datamodel, general-type]
---

# ProcessMongoEmployeeModel

โมเดลข้อมูลพนักงานที่ใช้ในขั้นตอน process ของ goapi ประกอบด้วยรหัส ชื่อ รหัสผ่าน และรหัส holding พร้อมฟิลด์สำหรับระบบอนุมัติ (ตำแหน่ง แผนก บทบาทการอนุมัติ และวงเงินอนุมัติสูงสุด)

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Code | string | code | code | รหัสพนักงาน |
| Name | string | name | name | ชื่อพนักงาน |
| Password | string | password | password | รหัสผ่าน |
| HoldingCode | string | holdingcode | holdingcode | รหัส holding (กลุ่มกิจการ/tenant) |
| Position | string | position | position | ตำแหน่งงาน เช่น "หัวหน้าแผนก", "ผู้จัดการฝ่าย", "ผู้อำนวยการ" |
| Department | string | department | department | แผนก เช่น "IT", "บัญชี", "จัดซื้อ" |
| ApprovalRole | int | approvalrole | approvalrole | บทบาทการอนุมัติ: 0=ไม่มีสิทธิ์, 1-4=ระดับผู้อนุมัติ |
| MaxApprovalAmount | float64 | maxapprovalamount | maxapprovalamount | วงเงินอนุมัติสูงสุด (บาท) — 0 = ไม่จำกัด |

## ความสัมพันธ์

- ไม่มี struct ซ้อนภายใน
- `HoldingCode` — อ้างอิงรหัส holding (ขอบเขต tenant/กลุ่มกิจการ) ที่พนักงานสังกัด
