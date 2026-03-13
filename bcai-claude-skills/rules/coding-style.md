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
- **ห้าม mock data**: ใช้ข้อมูลจริงจาก API เท่านั้น ห้ามใส่ข้อมูลจำลองเพื่อทดสอบ (รวม bcdashboard — ถ้า API ยังไม่พร้อมให้แจ้ง error ตรงๆ ไม่ต้อง fallback mock)
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

## ห้าม Hardcode สี (Theme Colors Rule) — บังคับทุกข้อ

**ทุกสีที่แสดงบน UI ต้องมาจาก `global.theme.*` เท่านั้น — ห้าม hardcode `Colors.*` หรือ `Color(0xFF...)` โดยตรง**

เหตุผล: ระบบรองรับ Dark Mode / Light Mode — ถ้า hardcode สีจะเปลี่ยนธีมไม่ได้

### ห้ามทำ (NEVER):
```dart
// ❌ ห้าม — hardcode สีโดยตรง
backgroundColor: Colors.white,
color: Colors.grey.shade200,
fillColor: Colors.white,
style: TextStyle(color: Colors.black87),
floatingLabelStyle: TextStyle(backgroundColor: Colors.white),
decoration: BoxDecoration(color: Color(0xFFF5F7FA)),
```

### ต้องทำแทน (DO THIS):
```dart
// ✅ ถูกต้อง — ใช้สีจาก global.theme
backgroundColor: global.theme.cardColor,
color: global.theme.dividerBorderColor,
fillColor: global.theme.formFillColor,
style: TextStyle(color: global.theme.formTextColor),
floatingLabelStyle: TextStyle(backgroundColor: global.theme.cardColor),
decoration: BoxDecoration(color: global.theme.surfaceColor),
```

### Mapping สี:

| Hardcode (ห้ามใช้) | ใช้แทน |
|---------------------|--------|
| `Colors.white` (background) | `global.theme.cardColor` หรือ `global.theme.backgroundColor` |
| `Colors.white` (form fill) | `global.theme.formFillColor` |
| `Colors.white` (text on dark bg) | `global.theme.onPrimaryColor` (text/icon บน AppBar, ปุ่มสี, gradient header) |
| `Colors.white` (icon on button) | `global.theme.onPrimaryColor` (Icon บน ElevatedButton สี) |
| `Colors.white` (button foreground) | `global.theme.onPrimaryColor` (foregroundColor ของปุ่มสี) |
| `Colors.white.withValues(alpha: x)` (overlay) | ปล่อยได้ถ้าเป็น semi-transparent overlay บน gradient/dark bg |
| `Colors.grey.shade50-200` (bg) | `global.theme.surfaceColor` |
| `Colors.grey.shade200-400` (border) | `global.theme.dividerBorderColor` |
| `Colors.grey.shade500-700` (icon) | `global.theme.iconSecondaryColor` |
| `Colors.grey.shade800-900` (text) | `global.theme.textColor` |
| `Colors.black87` (text) | `global.theme.textColor` |
| `Colors.black54` (secondary text) | `global.theme.textSecondaryColor` |
| `Color(0xFFF5F7FA)` (light bg) | `global.theme.surfaceColor` |
| `Colors.blue.shade50-100` (header) | `global.theme.columnHeaderColor` |
| `Colors.blue.shade900` (header text) | `global.theme.columnHeaderTextColor` |
| `Colors.red` / `Colors.red[50]` (error bg) | `global.theme.negativeHighlightColor` |
| `Colors.red[700-900]` (error text) | `global.theme.negativeHighlightTextColor` |
| `Colors.green` / `Colors.green[50]` (success bg) | `global.theme.positiveHighlightColor` |
| `Colors.green[700-900]` (success text) | `global.theme.positiveHighlightTextColor` |
| `Colors.orange` / `Colors.orange[50]` (warning bg) | `global.theme.warningHighlightColor` |
| `Colors.orange[700-900]` (warning text) | `global.theme.warningHighlightTextColor` |
| `Colors.blue` / `Colors.blue[50]` (info bg) | `global.theme.infoHighlightColor` |
| `Colors.blue[700-900]` (info text) | `global.theme.infoHighlightTextColor` |

### ข้อยกเว้น (ไม่ต้องเปลี่ยน):
- `Colors.white` ใน **QR Code** — QR ต้อง background ขาวเพื่อ scan ได้
- `Colors.black.withValues(alpha: ...)` ใน **boxShadow/barrier** — shadow ต้องเป็นสีเข้มเสมอ
- `Colors.transparent` — ไม่ต้องเปลี่ยน
- **PdfColors.\*** — PDF generation ใช้สีคงที่ (print ≠ หน้าจอ)
- **Brand colors** ที่ตายตัว (Google login=white button, LINE login=`0xFF06C755`, AI provider icons)
- **Chart/series colors** ใน `chartColors` map — data visualization identifiers
- **Financial category colors** (positiveColor, negativeColor ใน menu_screen) — fixed semantic values สำหรับ data
- **Login/usersystem screens** — มีดีไซน์เฉพาะ
- **chatbot/alert_agent/knowledge_base** — UI แยกอิสระ
- **Color picker initial values** (`Color colorSelected = Colors.white`) — เป็นข้อมูล ไม่ใช่ UI
- **SnackBar icon colors** (`Icon(Icons.check, color: Colors.white)`) — SnackBar มี bg เข้มเสมอ

### Pitfall: onPrimaryColor ห้ามใช้เป็น background/fillColor
`onPrimaryColor` = white ทั้ง dark+light mode → ใช้เป็น text/icon บนพื้นเข้มเท่านั้น
```dart
// BAD — form field ขาวจ้าใน dark mode
fillColor: global.theme.onPrimaryColor,
// GOOD
fillColor: global.theme.formFillColor,
```

### ThemeRefreshMixin — บังคับใช้ทุก StatefulWidget:
```dart
// ทุก State class ต้องใช้ mixin นี้เพื่อ rebuild เมื่อเปลี่ยนธีม
class _MyScreenState extends State<MyScreen> with global.ThemeRefreshMixin {
  // ...
}
```

### ตรวจสอบก่อน commit:
- `dart analyze lib/` ต้องไม่มี error
- ค้นหา `Colors.white`, `Colors.grey`, `Colors.black` ใน code ใหม่ — ต้องไม่มี (ยกเว้น exceptions ข้างบน)
- ค้นหา `Color(0xFF` — ต้องไม่มี hardcode สี light/dark ใน code ใหม่

## Docker
- Build: `docker compose build mainapi`
- Deploy: `docker compose up -d`
- Port: **8888** = จุดเข้าเดียว (MainAPI + GoAPI embedded)
- MongoDB/PostgreSQL = native on host (ไม่อยู่ใน Docker)
