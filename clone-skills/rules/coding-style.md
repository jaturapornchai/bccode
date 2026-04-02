# Coding Style Conventions

## Language
- **Variable/function names**: Always English
- **Comments/logs/docstrings**: Always Thai — explain logic in Thai
- **Commit messages**: English (clear, concise)

## Go (Backend)
- Framework: **Echo v4**
- Config: **bootstrap.json** only (never use .env)
- Logger: `logger.Info()`, `logger.Error()`, `logger.Warn()`, `logger.Success()`
- Error handling: Return error (never panic)
- Import path: `smlcloudplatform/internal/goapi/...`
- Database: Use `mydb.DatabaseManager` (circuit breaker + retry)
- HTTP response: `c.JSON()` with proper status codes

## Flutter (Frontend)
- State management: **BLoC pattern**
- HTTP client: **Dio**
- Logger: **AppLogger** (never use `print`)
- Localization: `global.language('key')` (always include `global.`)
- Navigation: GetX routing
- File naming: snake_case

## General (Mandatory Rules)
- **Write clean code**: Readable, well-structured, organized
- **Thai comments**: Every function/class must have Thai comments
- **Read before edit**: Always read existing code before modifying
- **Build first**: Compile/build must pass before deploy
- **Check imports**: Verify all imports after every edit
- **Use latest APIs**: Never use deprecated APIs (e.g., `withOpacity` → `withValues`)
- **No hardcode**: All values from config, API, or language file
- **No hidden fallbacks**: Never create fallback logic that hides errors — report failures directly to user
- **No mock data**: Use real API data only. If API not ready, show error (including bcdashboard)
- **Multi-language**: All UI text via `global.language('key')`. When adding new keys, update `languages.tsv` for all 9 languages (1 key = 1 line, Tab-separated: `key\tth\ten\tcn\tja\tkm\tko\tlo\tmy\tvi`)
- **Confirm before changing data model**: Ask Jead before modifying model, schema, or database structure
- **Modern UX/UI**: Material Design 3, user-friendly for Thai users

## Go Patterns
- **Singleton**: Package-level `var defaultXxx` + `GetDefaultXxx()` for shared resources
- **Interface for abstraction**: e.g., `ToolCallingProvider` interface for multiple AI providers
- **Fallback pattern**: `callWithFallback()` — try providers sequentially, switch on failure
- **Export types**: Uppercase type names used across packages (e.g., `OAIMessage` not `oaiMessage`)
- **Handler pattern**: `func Handler(c echo.Context) error` → register in `bootstrap.go`
- **JSON tags**: Use `json:"field_name"` + `omitempty` for optional fields

## Date Picker — Use CustomDatePicker Only
Never use Flutter's `showDatePicker()` — it lacks Buddhist Era (B.E.) support.
Always use `CustomDatePicker` which supports B.E./C.E. based on `global.profileData.yeartype`.
```dart
CustomDatePicker(
  labelText: global.language('field_name'),
  useIconSelectDate: true,
  initialDate: DateTime.now(),
  firstDate: DateTime(2020),
  lastDate: DateTime.now().add(const Duration(days: 365)),
  onDateSelected: (date) {
    if (date != null) {
      setState(() { controller.text = date.toIso8601String(); });
    }
  },
  decoration: const InputDecoration(),
)
```

## Flutter Patterns
- **const removal**: When changing hardcoded → runtime function, change `const` to `final`
- **Import alias**: `import '...global.dart' as global;` — never omit `as global`
- **URL construction**: Use `global.goApiUrlPath("endpoint")` — never concatenate strings
- **Batch fix**: 100+ files → use Python script → fix manual cases after

## Theme Colors — No Hardcoded Colors (Mandatory)

**All UI colors must come from `global.theme.*` — never hardcode `Colors.*` or `Color(0xFF...)`.**

Reason: System supports Dark/Light mode — hardcoded colors break theme switching.

### Never Do:
```dart
// ❌ Hardcoded colors
backgroundColor: Colors.white,
color: Colors.grey.shade200,
fillColor: Colors.white,
style: TextStyle(color: Colors.black87),
floatingLabelStyle: TextStyle(backgroundColor: Colors.white),
decoration: BoxDecoration(color: Color(0xFFF5F7FA)),
```

### Do Instead:
```dart
// ✅ Theme colors
backgroundColor: global.theme.cardColor,
color: global.theme.dividerBorderColor,
fillColor: global.theme.formFillColor,
style: TextStyle(color: global.theme.formTextColor),
floatingLabelStyle: TextStyle(backgroundColor: global.theme.cardColor),
decoration: BoxDecoration(color: global.theme.surfaceColor),
```

### Color Mapping:

| Hardcoded (forbidden) | Use instead |
|----------------------|-------------|
| `Colors.white` (background) | `global.theme.cardColor` or `global.theme.backgroundColor` |
| `Colors.white` (form fill) | `global.theme.formFillColor` |
| `Colors.white` (text on dark bg) | `global.theme.onPrimaryColor` |
| `Colors.white` (icon on button) | `global.theme.onPrimaryColor` |
| `Colors.white` (button foreground) | `global.theme.onPrimaryColor` |
| `Colors.white.withValues(alpha: x)` (overlay) | OK if semi-transparent overlay on gradient/dark bg |
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

### Exceptions (no change needed):
- `Colors.white` in **QR Code** — white background required for scanning
- `Colors.black.withValues(alpha: ...)` in **boxShadow/barrier** — shadows must be dark
- `Colors.transparent` — no change needed
- **PdfColors.\*** — PDF generation uses fixed colors (print ≠ screen)
- **Brand colors** (Google login=white, LINE login=`0xFF06C755`, AI provider icons)
- **Chart/series colors** in `chartColors` map — data visualization identifiers
- **Financial category colors** (positiveColor, negativeColor in menu_screen) — fixed semantic values
- **Login/usersystem screens** — custom design
- **chatbot/alert_agent/knowledge_base** — independent UI
- **Color picker initial values** (`Color colorSelected = Colors.white`) — data, not UI
- **SnackBar icon colors** (`Icon(Icons.check, color: Colors.white)`) — SnackBar always has dark bg

### Pitfall: onPrimaryColor is not for backgrounds
`onPrimaryColor` = white in both dark+light mode. Use only for text/icon on dark surfaces.
```dart
// BAD — form field glaring white in dark mode
fillColor: global.theme.onPrimaryColor,
// GOOD
fillColor: global.theme.formFillColor,
```

### ThemeRefreshMixin — Required for every StatefulWidget:
```dart
class _MyScreenState extends State<MyScreen> with global.ThemeRefreshMixin {
  // ...
}
```

### Pre-commit checks:
- `dart analyze lib/` must have no errors
- Search for `Colors.white`, `Colors.grey`, `Colors.black` in new code — must not exist (except above exceptions)
- Search for `Color(0xFF` — no hardcoded light/dark colors in new code

## Docker
- Build: `docker compose build mainapi`
- Deploy: `docker compose up -d`
- Port: **8888** = single entry point (MainAPI + GoAPI embedded)
- MongoDB/PostgreSQL = native on host (not in Docker)
