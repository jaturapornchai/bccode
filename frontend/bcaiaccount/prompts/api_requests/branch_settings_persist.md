# API Request: Branch Settings Persistence (base_currency, language, timezone, yeartype)

## Status: RESOLVED
**Backend มี fields ครบแล้ว** — Go struct ของ branch model มี fields ทั้งหมดพร้อม bson tags

```go
BaseCurrency     string `json:"base_currency" bson:"base_currency"`
Language         string `json:"language" bson:"language"`
Timezone         string `json:"timezone" bson:"timezone"`
DateFormat       string `json:"date_format" bson:"date_format"`
YearType         string `json:"yeartype" bson:"yeartype"`
TimezoneLabel    string `json:"timezonelabel" bson:"timezonelabel"`
TimezoneOffset   string `json:"timezoneoffset" bson:"timezoneoffset"`
DecimalQuantity  int8   `json:"decimal_quantity" bson:"decimal_quantity"`
DecimalPrice     int8   `json:"decimal_price" bson:"decimal_price"`
DecimalDocument  int8   `json:"decimal_document" bson:"decimal_document"`
```

## สาเหตุของปัญหาเดิม
- Dialog "ตั้งค่าเริ่มต้นสาขา" แสดงทุกครั้ง เพราะ initial setup save ไม่สำเร็จ (405 Method Not Allowed)
- สาเหตุ: guidfixed ว่าง → PUT ไปที่ `/organization/branch` แทน `/organization/branch/{id}`

## การแก้ไข (Frontend)
1. เพิ่ม guard `if (branch.guidfixed.isNotEmpty)` ก่อน PUT
2. เพิ่ม local cache (SharedPreferences) เป็น fallback
3. ครั้งถัดไป: โหลดจาก cache ถ้า API ยังไม่มีค่า → ไม่แสดง dialog ซ้ำ

## Endpoints ที่เกี่ยวข้อง (ใช้งานได้แล้ว)
- **PUT** `/organization/branch/{guid}` — save fields ครบ
- **GET** `/organization/branch/list` — return fields ครบ
- **GET** `/organization/branch/{guid}` — return fields ครบ
