# Schema Migration — กฎเมื่อ DB Schema เปลี่ยน

## กฏหลัก

เมื่อ backend มีการเปลี่ยน database schema ไม่ว่าจะเป็น:
- เพิ่ม/ลบ/เปลี่ยนชื่อ column
- เปลี่ยน data type หรือ nullable
- เพิ่ม/ลบ table หรือ collection

**ต้องแจ้ง frontend dev เสมอ**

## ทำไมต้องแจ้ง?

- Dart model อาจ parse ผิด (field หาย → null crash)
- `fromJson` ที่ไม่ handle null จะ throw exception ใน production
- Frontend อาจส่ง field ที่ backend ไม่รับแล้ว

## Migration Impact Report

ทุกครั้งที่ schema เปลี่ยน ต้องสรุป:

```markdown
## Migration Impact Report

### Changes
- [เพิ่ม/ลบ/เปลี่ยน] column `{name}` ใน table `{table}`
- Type: `{old_type}` → `{new_type}`
- Nullable: `{old}` → `{new}`

### Frontend Impact
- [ ] Dart model `{ModelName}` ต้อง update field `{name}`
- [ ] `fromJson()` ต้อง handle null/missing field
- [ ] Repository response parsing ต้องปรับ

### Backward Compatibility
- [ ] Frontend เก่ายังใช้ได้ไหม?
- [ ] ต้อง deploy frontend ก่อน/หลัง backend?
```

## 3-Step Process

1. **สร้าง Impact Report** → แจ้ง frontend dev
2. **Deploy backend** → ตรวจสอบว่า API response ยังถูกต้อง
3. **Deploy frontend** → update Dart models ให้ตรง

## ข้อควรระวัง

| Action | ความเสี่ยง | วิธีจัดการ |
|--------|-----------|-----------|
| เพิ่ม column | ต่ำ | Frontend เก่ายัง work (field ใหม่ถูก ignore) |
| ลบ column | สูง | Frontend จะ crash ถ้า field หาย → ต้องแก้ frontend ก่อน |
| เปลี่ยนชื่อ column | สูง | เหมือนลบ+เพิ่ม → ต้องแก้ทั้ง 2 ฝั่ง |
| เปลี่ยน type | สูง | parse ผิด → ต้องแก้ทั้ง 2 ฝั่ง |
| เปลี่ยน nullable | กลาง | `fromJson` อาจ throw → เพิ่ม null check |
