# API Request: Language Endpoint — Public (No Auth Required)

## สิ่งที่ต้องการ
ปัจจุบัน endpoint `GET /api/language/:lang` ต้องการ JWT token (ตอบ 401 ถ้าไม่มี)
ต้องการให้เปลี่ยนเป็น **public endpoint** ที่เรียกได้โดยไม่มี token

## Endpoint ที่ต้องการ
- **Method:** GET
- **Path:** `/api/language/:lang`
- **GoAPI full path:** `GET /goapi/api/language/:lang`
- **Auth:** ❌ ไม่ต้อง (ต้องลบ/ข้าม auth middleware สำหรับ route นี้)

## Request
```
GET /goapi/api/language/th
GET /goapi/api/language/en
```

| Parameter | In   | Type   | Required | Description          |
|-----------|------|--------|----------|----------------------|
| lang      | path | string | ✅        | รหัสภาษา เช่น th, en |

## Expected Response: 200 OK
```json
{
  "common.save": "บันทึก",
  "common.cancel": "ยกเลิก",
  "common.search": "ค้นหา",
  "login.title": "เข้าสู่ระบบ"
}
```
- Response เป็น flat object `{ "code": "text" }` — ไม่มี nested
- Key ที่ไม่มีใน `:lang` จะไม่ถูก include (กรองออกอัตโนมัติ)

## Error Responses
| Status | เมื่อไร |
|--------|---------|
| 400    | ไม่ส่ง `:lang` |
| 500    | โหลดไฟล์ภาษาไม่ได้ (internal error) |

## Use Case
- **Login Page**: โหลดข้อความภาษาก่อน login — ผู้ใช้ยังไม่มี token เลย
- **First-time user**: ตั้งค่า server URL แล้วต้องการแสดงภาษาที่ถูกต้องทันที
- **หน้า login_password_screen.dart**: เรียก `languageSelect()` ใน `_navigateToShopScreen()` ซึ่งอยู่ใน login flow ก่อนที่จะเข้า shop screen

## สิ่งที่ต้องแก้ใน Backend
1. **ลบ/ข้าม auth middleware** สำหรับ route `GET /api/language/:lang`
2. ให้ route นี้ถูก register ก่อน middleware ตรวจ token หรือใช้ public route group

## หมายเหตุสำหรับ Flutter
- Flutter code ใน `global.dart:loadLanguageFile()` เรียก endpoint นี้แบบ plain `http.get(url)` ไม่ส่ง token
- เมื่อ backend เปิดเป็น public แล้ว Flutter code นั้นจะทำงานได้ถูกต้องทันที
