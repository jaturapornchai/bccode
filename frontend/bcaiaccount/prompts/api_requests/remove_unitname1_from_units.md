# API Request: ลบ field `unitname1` ออกจาก units collection

## สิ่งที่ต้องการ
ลบ field `unitname1` ออกจาก `units` MongoDB collection และ Go model เพราะไม่ได้ใช้งานจริง — ชื่อหน่วยนับใช้ `names` array (multi-language) แทน

## สถานะปัจจุบัน
- `units` collection มี field `unitname1` แต่ค่าว่างเปล่า (`""`) ทุก document
- ชื่อหน่วยนับจริงเก็บใน `names` array: `[{code:"th", name:"หัว"}, {code:"en", name:""}]`
- Flutter `UnitModel` ไม่มี `unitname1` อยู่แล้ว — ใช้ `names` ถูกต้อง

## สิ่งที่ต้องทำ

### 1. ลบ `unitname1` จาก Go struct
- หาไฟล์ model ของ unit (เช่น `models/unit.go` หรือที่เกี่ยวข้อง)
- ลบ field `unitname1` ออกจาก struct

### 2. ลบ `unitname1` จาก MongoDB documents ที่มีอยู่
```javascript
// MongoDB migration script
db.units.updateMany({}, { $unset: { unitname1: "" } })
```

### 3. ตรวจสอบ handler/service ที่อ้างถึง `unitname1`
- ค้นหา `unitname1` ใน codebase ทั้งหมด
- ลบ reference ที่เกี่ยวข้องออก

## Use Case
ลบ field ที่ไม่ได้ใช้เพื่อลดความสับสน — `names` array เป็น source of truth สำหรับชื่อหน่วยนับ
