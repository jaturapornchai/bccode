# Coding Style — บอสจืด's Conventions

## ภาษา
- **Variable/function names**: ภาษาอังกฤษเสมอ
- **Comments/logs/docstrings**: ภาษาไทยเสมอ — อธิบาย logic เป็นภาษาไทย
- **Commit messages**: ภาษาอังกฤษ (ชัดเจน กระชับ)

## Go (Backend)
- Framework: **Echo v4**
- Config: **bootstrap.json** only (ห้ามใช้ .env)
- Logger: `logger.Info()`, `logger.Error()`, `logger.Warn()`, `logger.Success()`
- Error handling: return error (ห้าม panic)
- Import path: `smlcloudplatform/internal/goapi/...`
- Database: ใช้ `mydb.DatabaseManager` (circuit breaker + retry)
- HTTP response: `c.JSON()` with proper status codes

## Flutter (Frontend)
- State management: **BLoC pattern**
- HTTP client: **Dio**
- Logger: **AppLogger** (ห้ามใช้ `print`)
- Localization: `global.language('key')` (ห้ามลืม `global.`)
- Navigation: ตาม project convention (GetX routing)
- ชื่อไฟล์: snake_case

## General (กฏเหล็ก — บังคับทุกข้อ)
- **Clean Code เสมอ**: เขียน code ให้อ่านง่าย เรียบร้อย มีโครงสร้างชัดเจน
- **หมายเหตุภาษาไทย**: ทุก function/class ต้องมี comment อธิบายเป็นภาษาไทย
- **อ่านก่อนแก้**: ต้องอ่าน code เดิมก่อนแก้ไขเสมอ
- **ทดสอบ**: build/compile ให้ผ่านก่อน deploy
- **Import ให้ครบ**: ตรวจสอบ import ทุกครั้งหลังแก้ไข
- **Upgrade ให้ทันสมัย**: ใช้ API/method ล่าสุดเสมอ ห้ามใช้ deprecated APIs (เช่น withOpacity → withValues)
- **ห้าม hardcode**: ค่าต่างๆ ต้องมาจาก config, API, หรือ language file
- **ห้าม fallback ซ่อน error**: ไม่สร้าง fallback logic ที่ซ่อนปัญหา — ถ้า fail ให้แจ้ง user ตรงๆ
- **ห้าม mock data**: ใช้ข้อมูลจริงจาก API เท่านั้น ห้ามใส่ข้อมูลจำลองเพื่อทดสอบ
- **รองรับหลายภาษา**: ทุก UI text ใช้ `global.language('key')` ห้าม hardcode ข้อความ เมื่อเพิ่ม key ใหม่ต้อง update `languages.json` ครบ 9 ภาษาเสมอ
- **ขอยืนยันก่อนแก้ data model**: ถ้าจะแก้ model, schema, หรือ database structure → ต้องถามบอสจืดก่อนเสมอ
- **UX/UI สวยทันสมัย**: ออกแบบ UI ให้ดูดี ใช้ง่าย เหมาะกับ user คนไทย ใช้ Material Design 3

## Go Patterns (เรียนรู้จาก session จริง)
- **Singleton**: ใช้ package-level `var defaultXxx` + `GetDefaultXxx()` สำหรับ shared resources
- **Interface for abstraction**: เช่น `ToolCallingProvider` interface เพื่อ support multiple AI providers
- **Fallback pattern**: `callWithFallback()` — ลอง provider ทีละตัว ถ้า fail ให้ switch ทันที
- **Export types**: ต้อง uppercase ชื่อ type ที่ใช้ข้าม package (เช่น `OAIMessage` ไม่ใช่ `oaiMessage`)
- **Handler pattern**: function `func Handler(c echo.Context) error` → register ใน `bootstrap.go`
- **JSON tags**: ใช้ `json:"field_name"` + `omitempty` สำหรับ optional fields

## Flutter Patterns (เรียนรู้จาก session จริง)
- **const removal**: ถ้าเปลี่ยนจาก hardcoded → runtime function → ต้องเปลี่ยน `const` เป็น `final`
- **Import alias**: `import '...global.dart' as global;` — ห้ามลืม `as global`
- **URL construction**: ใช้ `global.goApiUrlPath("endpoint")` ห้ามต่อ string เอง
- **Batch fix**: 100+ files → ใช้ Python script → แก้ manual cases ทีหลัง

## Docker
- Build: `docker compose build mainapi`
- Deploy: `docker compose up -d`
- Port: **8888** = จุดเข้าเดียว (MainAPI + GoAPI embedded)
- MongoDB/PostgreSQL = native on host (ไม่อยู่ใน Docker)
