# Coding Style — Jead's Conventions

## ภาษา
- **Variable/function names**: ภาษาอังกฤษเสมอ
- **Comments/logs**: ภาษาไทยได้
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

## General
- **Pragmatic approach**: ทำให้ work ก่อน refactor ทีหลัง
- **อย่า over-engineer**: ทำแค่ที่ขอ ไม่เพิ่มของที่ไม่จำเป็น
- **อ่านก่อนแก้**: ต้องอ่าน code เดิมก่อนแก้ไขเสมอ
- **ทดสอบ**: build/compile ให้ผ่านก่อน deploy
- **Import ให้ครบ**: ตรวจสอบ import ทุกครั้งหลังแก้ไข

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
