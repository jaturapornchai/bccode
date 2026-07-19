---
source: query-result-model.go
tags: [datamodel, general-type]
---

# ResultGetResponse

Response model สำหรับ endpoint `/resultget` — คืนข้อมูลผลลัพธ์ query พร้อมข้อมูลการแบ่งหน้า

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Status | string | - | status | สถานะ (success หรือ error) |
| Data | []map[string]interface{} | - | data | ข้อมูลผลลัพธ์ query |
| Pagination | [[PaginationInfo]] | - | pagination | ข้อมูลการแบ่งหน้า |
| Message | string | - | message | ข้อความเพิ่มเติม (optional) |

## ความสัมพันธ์

- ฝัง [[PaginationInfo]] เป็นข้อมูลการแบ่งหน้า
- เป็นผลตอบกลับของ [[ResultGetRequest]]
