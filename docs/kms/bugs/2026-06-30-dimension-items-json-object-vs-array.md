---
tags: [bug, bc-account, frontend, go, postgres]
date: 2026-06-30
---

# มิติสินค้า (productdimension) บันทึกไม่ได้ — items json ส่ง `{}` แทน `[]`

## Symptom
หน้าจอ **มิติสินค้า** (`/productdimension`) กด "เพิ่ม" → กรอกชื่อ → "บันทึก" แล้ว record ไม่ถูกสร้าง (รายการว่าง). UI ไม่ขึ้น error ชัด. พบโดย Stagehand UAT (create=false) — เป็นจอเดียวใน 8 product-master ที่ fail.

## Root cause
POST `/dimension` ตอบ **400**:
```
json: cannot unmarshal object into Go struct field Dimension.items of type []models.DimensionItem
```
ฟอร์มมี `jsonField("items", "รายการมิติย่อย")` (textarea JSON). Field json ที่เป็น **array** ต้อง default เป็น `"[]"` แต่ logic มี whitelist เฉพาะ `permissioncodes/approvalcodes/allowedtools` → `[]`, ที่เหลือ (รวม `items`) default เป็น `"{}"` (object). backend struct `Dimension.items` เป็น slice → unmarshal object ลง slice ไม่ได้ → 400.

whitelist ซ้ำกัน 2 ที่ใน `frontend/src/app/system-settings/system-settings-screen.tsx`:
- **~line 14976** — seed default ตอนเปิดฟอร์ม "เพิ่ม"
- **~line 16617** — `parseJsonField()` ตอน submit (textarea ว่าง)

## Fix
เพิ่ม `items` เข้า array-branch ทั้ง 2 ที่ (`field.key === "items"` / `key === "items"`) → seed + serialize เป็น `[]`.
ปลอดภัย: key `"items"` มีจอเดียว (dimension); json field อื่น (paymentrounding/pointconfig/permission*) ไม่กระทบ.

## Regression test
Stagehand UAT `D:\bccode\scratch\stagehand-uat\uat-dim.js` (DeepSeek): login → /productdimension → add+ชื่อ+save → edit → delete+confirm. ผ่าน `{create:true,edit:true,confirmModal:true,deleted:true}` หลังแก้. ครบทั้ง 8 จอผ่าน UAT (`uat-all.js`).

related: [[stagehand-uat-harness]]
