---
source: report-model.go
tags: [datamodel, general-type]
---

# ReportConditionProductBalanceModel

เงื่อนไขสำหรับรายงานยอดคงเหลือสินค้า (product balance) ระบุวันเวลาสิ้นสุดที่ใช้คำนวณยอด

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| EndDateTime | time.Time | enddatetime | enddatetime | วันเวลาสิ้นสุดของข้อมูลที่ใช้คำนวณยอดคงเหลือ |

## ความสัมพันธ์

- ไม่มี struct ซ้อนหรือ reference-by-code ที่ชัดเจน
