---
name: theming
description: >
  Dark/Light theme standards for Flutter ERP screens. Color properties, hardcode scanning,
  ThemeRefreshMixin, data list row colors, form styling.
  Triggers: "theme", "dark mode", "color", "backgroundColor", "textColor", "hardcoded colors".
user-invocable: true
---

# Theming — Dark/Light Mode Standards

## Usage
- `/theming` — Check and fix colors to match theme standards
- `/theming <screen-name>` — Check specific screen
- `/theming check` — Scan entire project for hardcoded colors

## Architecture

### Key Files
| File | Purpose |
|------|---------|
| `lib/model/theme_model.dart` | Color properties model |
| `lib/global.dart` > `themeSelect()` | Light mode colors (brand-based) |
| `lib/global.dart` > `applyThemeMode()` | Dark mode overrides |
| `lib/global.dart` > `isDarkMode()` | Current mode check |
| `lib/main_app.dart` | ThemeData + InputDecorationTheme |

### Flow
```
app start -> themeSelect(brand) -> applyThemeMode() -> displaySettingsNotifier -> MaterialApp rebuild
user toggle -> saveThemeSettings() ---^
```

### 3 Theme Modes
- `auto` — follows system (PlatformDispatcher.platformBrightness)
- `light` — forced light
- `dark` — forced dark

### Brand Theme (orthogonal to dark/light)
- Mode 0 = Default (blue #2A6F97)
- Mode 1 = DoHome (dark blue #235396)
- Dark mode override runs after brand — changes bg/text, not primary

---

## Color Properties

See [references/color-mappings.md](references/color-mappings.md) for the full property table with light/dark values.

### Quick Reference — Property Selection
| Need | Use |
|------|-----|
| Container/Card background | `cardColor` |
| Search bar background | `searchBarColor` |
| Body gradient | `bodyGradientColors` |
| Dialog background | `dialogColor` |
| Primary text | `textColor` |
| Secondary text (hint, subtitle) | `textSecondaryColor` |
| Text on AppBar/info bar | `onPrimaryColor` |
| General icon | `iconColor` |
| Secondary icon | `iconSecondaryColor` |
| Border/divider | `dividerBorderColor` |
| Input field border | `formBorderColor` |
| Input field fill | `formFillColor` |
| Data list hover | `rowHoverColor` |
| Data list selected (view) | `rowSelectedColor` |
| Data list selected (edit) | `rowEditColor` |

---

## Rules

### No hardcoded colors
```dart
// BAD
color: Colors.white,
color: Color.fromARGB(255, 148, 160, 194),
color: Color(0xFF94A0C2),

// GOOD
color: global.theme.cardColor,
color: global.theme.textColor,
```

See [references/color-mappings.md](references/color-mappings.md) for the full replacement table (Colors.*, Color.fromARGB, etc.).

### Accent color on dark backgrounds
`appBarColor` is #0D1B2A in dark mode — too dark for accent use. Switch based on mode:
```dart
final accentColor = global.isDarkMode()
    ? global.theme.primaryLightColor  // #89C2D9
    : global.theme.appBarColor;       // #012A4A
```

### ThemeRefreshMixin — required on all StatefulWidgets
Without it, screens won't rebuild on theme change:
```dart
class _MyScreenState extends State<MyScreen> with global.ThemeRefreshMixin { }
```

### No `const` with theme colors
`global.theme.*` is not compile-time constant:
```dart
// BAD — compile error
const BoxDecoration(color: global.theme.cardColor)
// GOOD
BoxDecoration(color: global.theme.cardColor)
```

### No Color[] operator
`global.theme.*` is `Color`, not `MaterialColor` — no `[shade]` subscript.

### onPrimaryColor is text-only
White in both modes. For fill/background use `formFillColor`, `cardColor`, or `dialogColor`.

### floatingLabelStyle backgroundColor
Hardcoding `Colors.white` shows a white rectangle behind label in dark mode. Use `global.theme.cardColor`.

---

## _getContainerColor (Data List Rows)

```dart
Color? _getContainerColor(String itemGuid, int index) {
  if (selectGuid.isNotEmpty && selectGuid == itemGuid) {
    return isEditMode // or isSaveAllow — depends on screen
        ? global.theme.rowEditColor
        : global.theme.rowSelectedColor;
  }
  if (_hoverIndex == index) return global.theme.rowHoverColor;
  return (index % 2 == 0)
      ? global.theme.columnAlternateEvenColor
      : global.theme.columnAlternateOddColor;
}
```

Pitfall: `screenEvent == edit` does not work — `switchToEdit()` doesn't set screenEvent. Use `isEditMode` or `isSaveAllow`.

---

## Search Bar
Use `searchBarColor` (darker than surfaceColor), `filled: false`, add `textColor` and `formHintColor`. See `/data-list` skill for full snippet.

## Edit Form Body
Use `cardColor` background, `formBorderColor` for enabled border, `inputTextBoxForceColor` for required labels.

## Adding New Theme Colors
1. Add property in `lib/model/theme_model.dart`
2. Set light in `global.dart` > `themeSelect()`, dark in `applyThemeMode()`
3. Use via `global.theme.newProperty`

Dark mode: dark grey bg (#1A1A2E) not pure black, text #E0E0E0 not pure white, min 4.5:1 contrast.

## Excluded Files/Patterns
**Directories:** `pdfgen/`, `usersystem/`, `global.dart` (theme definitions themselves)
**Allowed:** `Colors.black.withValues(alpha: 0.01-0.5)` (shadow), `Colors.transparent`, `PdfColors.*`, color picker data, SnackBar white icons.
**bclms** uses `AppTheme`/`Theme.of(context)` — different system.

---

## Checklist

- [ ] Scaffold has `backgroundColor: global.theme.backgroundColor`
- [ ] AppBar has `backgroundColor: global.theme.appBarColor`
- [ ] No `Colors.*` as background (use `cardColor`/`surfaceColor`)
- [ ] No `Colors.black`/`Colors.grey[600-800]` as text (use `textColor`)
- [ ] Search bar uses `searchBarColor`
- [ ] Edit form uses `cardColor` background
- [ ] Form fields use `formFillColor`/`formBorderColor`/`formHintColor`
- [ ] Row colors use `rowHoverColor`/`rowSelectedColor`/`rowEditColor`
- [ ] No `const` wrapping widgets with `global.theme`
- [ ] Tested in both Light and Dark mode
- [ ] State class has `ThemeRefreshMixin`

## Batch Fix Workflow (`/theming check`)

1. **Scan**: `grep -r "Colors\." --include="*.dart" lib/screens/ | grep -v "pdfgen\|usersystem\|global.dart" | wc -l`
2. **Prioritize**: white bg (highest) > black text > grey borders > semantic colors
3. **Fix by folder**: config/ > master/ > transaction/ > report/ > gl/ > screen_search/
4. **Verify**: `dart analyze lib/ 2>&1 | grep -c "error"` must be 0

## References
- [Color Mappings](references/color-mappings.md) — full property table + replacement mappings
