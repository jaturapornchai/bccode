---
source: pdf-history-model.go
tags: [datamodel, general-type]
---

# PdfHistoryListResponse

Response สำหรับรายการประวัติ PDF (ตาม comment ในโค้ด) ห่อสถานะ, จำนวนรายการ และ array ของ [[PdfHistoryListItem]]

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Status | string | - | status | สถานะการทำงาน |
| Code | int | - | code | รหัสผลลัพธ์ |
| Count | int | - | count | จำนวนรายการที่ส่งกลับในหน้านี้ |
| Total | int64 | - | total | จำนวนรายการทั้งหมด |
| Data | [][[PdfHistoryListItem]] | - | data | รายการประวัติ PDF |

## ความสัมพันธ์

- `Data` — embed slice ของ [[PdfHistoryListItem]]
- เป็น response คู่กับ [[PdfHistoryListRequest]]
