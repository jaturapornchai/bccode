---
source: query-result-model.go
tags: [datamodel, general-type]
---

# ResultFromQueryResponse

Response model สำหรับ endpoint `/resultfromquery` — บอกสถานะการรัน query, GUID ของผลลัพธ์ที่บันทึก และจำนวนแถว

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Status | string | - | status | สถานะ (success หรือ error) |
| GUID | string | - | guid | GUID ของผลลัพธ์ที่บันทึก |
| Count | int | - | count | จำนวนแถวที่บันทึก |
| Message | string | - | message | ข้อความสำเร็จ/ผิดพลาด |

## ความสัมพันธ์

- เป็นผลตอบกลับของ [[ResultFromQueryRequest]] — `GUID` ใช้ต่อใน [[ResultGetRequest]] และ [[ResultToPDFRequest]]
