---
source: query-result-model.go
tags: [datamodel, general-type]
---

# ErrorResponse

Response model สำหรับกรณีผิดพลาด — `Status` เป็น "error" เสมอ (สร้างผ่าน helper `NewErrorResponse`)

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Status | string | - | status | สถานะ (เป็น 'error' เสมอ) |
| Message | string | - | message | ข้อความผิดพลาด |
| Detail | string | - | detail | รายละเอียดข้อผิดพลาดเพิ่มเติม (optional) |

## ความสัมพันธ์

- ไม่มี struct ฝัง — ใช้เป็น error response กลางของกลุ่ม endpoint query result (`/resultfromquery`, `/resultget`, `/resulttopdf`)
