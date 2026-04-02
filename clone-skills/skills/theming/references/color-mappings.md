# Color Mappings Reference

## All Theme Properties (global.theme.xxx)

### Base Colors
| Property | Light | Dark | Usage |
|----------|-------|------|-------|
| `backgroundColor` | grey[50] | #1A1A2E | Scaffold background |
| `appBarColor` | #012A4A | #0D1B2A | AppBar |
| `primaryColor` | #2A6F97 | (unchanged) | Buttons, accent, info |
| `primaryLightColor` | #89C2D9 | (unchanged) | Hover, light accent |
| `secondaryColor` | white | #A9D6E5 | Secondary surfaces |
| `headTitleColor` | white | white | AppBar title |
| `columnHeaderColor` | #89C2D9 | #2D3A4A | Data list header row |
| `columnHeaderTextColor` | black | #E0E0E0 | Header text |
| `columnAlternateEvenColor` | #F3F7FA | #1E1E30 | Even rows |
| `columnAlternateOddColor` | white | #1A1A2E | Odd rows |
| `inputTextBoxForceColor` | #8A1606 | #FF6B6B | Required label |
| `inputTextBoxColor` | black | #E0E0E0 | Optional label |
| `toolBarEditModeColor` | #2A6F97 | #1B3A5C | Toolbar in edit mode |
| `buttonDangerColor` | #DC2626 | #EF4444 | Delete/danger button |

### UI Colors
| Property | Light | Dark | Usage |
|----------|-------|------|-------|
| `cardColor` | white | #252538 | Card/Container bg |
| `textColor` | #212121 | #E0E0E0 | Primary text |
| `textSecondaryColor` | grey.600 | grey.400 | Secondary text |
| `surfaceColor` | grey.100 | #252538 | Tab header, surface |
| `searchBarColor` | grey.200 | #141825 | Search bar bg (darker than surface) |
| `dividerBorderColor` | grey.300 | grey.700 | Borders, dividers |
| `dialogColor` | white | #2D2D2D | Dialog bg |
| `inputFillColor` | white | #2D2D2D | Input field bg |
| `bodyGradientColors` | [white, primaryLight, white] | [#1A1A2E, #16213E, #1A1A2E] | Body gradient |

### Icon Colors
| Property | Light | Dark | Usage |
|----------|-------|------|-------|
| `iconColor` | #424242 | #B0BEC5 | Primary icon |
| `iconSecondaryColor` | grey.500 | grey.500 | Secondary icon |

### Navigation & Bar
| Property | Light | Dark | Usage |
|----------|-------|------|-------|
| `infoBarColor` | #D4A373 (95%) | #0D1B2A (95%) | Store info bar |
| `bottomNavColor` | #D4A373 (95%) | #0D1B2A | Bottom nav |
| `bottomNavSelectedColor` | white | white | Selected nav item |
| `bottomNavUnselectedColor` | white (60%) | white (60%) | Unselected nav item |
| `onPrimaryColor` | white | white | Text on dark bg |
| `onPrimarySecondaryColor` | white70 | white70 | Secondary text on dark bg |

### Data List
| Property | Light | Dark | Usage |
|----------|-------|------|-------|
| `rowHoverColor` | blue.50 | white (6%) | Row hover |
| `rowSelectedColor` | cyan.100 | cyan (12%) | Selected (view) |
| `rowEditColor` | orange.100 | orange (15%) | Selected (edit) |

### Form
| Property | Light | Dark | Usage |
|----------|-------|------|-------|
| `formFillColor` | white | #2D2D2D | Input field bg |
| `formTextColor` | #212121 | #E0E0E0 | Field text |
| `formLabelColor` | black | #E0E0E0 | Normal label |
| `formLabelRequiredColor` | #8A1606 | #FF6B6B | Required label |
| `formBorderColor` | grey.300 | grey.600 | Field border |
| `formFocusBorderColor` | #2A6F97 | #89C2D9 | Focused border |
| `formHintColor` | grey.400 | grey.500 | Hint text |

---

## Color.fromARGB Replacement Table

| Color.fromARGB | Color | Replace With |
|----------------|-------|-------------|
| `(255, 123, 235, 157)` | light green (button) | `positiveHighlightTextColor` |
| `(255, 78, 141, 66)` / `(255, 67, 206, 74)` | dark green (button) | `positiveHighlightTextColor` |
| `(255, 164, 232, 162)` / `(255, 148, 194, 168)` | light green (card bg) | `positiveHighlightColor` |
| `(255, 241, 112, 112)` / `(255, 246, 137, 129)` | red/salmon (button) | `buttonDangerColor` |
| `(255, 238, 86, 144)` | pink (delete icon) | `negativeHighlightTextColor` |
| `(255, 235, 147, 123)` | orange (warning button) | `warningHighlightTextColor` |
| `(255, 224, 159, 67)` | orange (card bg) | `warningHighlightColor` |
| `(255, 148, 160, 194)` / `(255, 197, 212, 255)` | blue/purple (card bg) | `surfaceColor` or `infoHighlightColor` |
| `(255, 99, 120, 183)` / `(255, 29, 43, 84)` | dark blue (button) | `primaryColor` |
| `(255, 168, 171, 136)` | olive (selector) | `primaryLightColor` / `surfaceColor` |
| `(255, 208, 42, 158)` | magenta | `secondaryColor` |
| `(255, 0, 0, 0)` | black | `textColor` / `iconColor` |
| `(255, 255, 255, 255)` | white (on dark bg) | `onPrimaryColor` |
| `(255, 75, 75, 75)` / `(255, 65, 66, 74)` | dark grey (button) | `primaryColor` |
| `(255, 152, 152, 152)` | mid grey | `textSecondaryColor` |
| `(255, 231, 206, 209)` | light pink (card bg) | `surfaceColor` |

---

## Colors.* Replacement Table

| Hardcoded | Replace With | Context |
|-----------|-------------|---------|
| `Colors.white` | `cardColor` | Card/Container bg |
| `Colors.white` | `formFillColor` | Form field fill |
| `Colors.white` | `dialogColor` | Dialog bg |
| `Colors.white` (on dark bg) | `onPrimaryColor` | Text/icon on AppBar |
| `Colors.black` / `Colors.black87` | `textColor` | Primary text |
| `Colors.black54` | `textSecondaryColor` | Secondary text |
| `Colors.grey[800]` / `[700]` | `textColor` / `textSecondaryColor` | Text |
| `Colors.grey[600]` | `textSecondaryColor` | Secondary text/icon |
| `Colors.grey[500]` | `iconSecondaryColor` | Secondary icon |
| `Colors.grey[400]` | `formHintColor` | Hint text |
| `Colors.grey[300]` | `dividerBorderColor` | Borders |
| `Colors.grey[200]` | `searchBarColor` | Search bar bg |
| `Colors.grey[200]` | `surfaceColor` | Non-search surface |
| `Colors.grey[100]` / `[50]` | `surfaceColor` | Section card bg |
| `Colors.red` (button bg) | `buttonDangerColor` | Delete/danger button |
| `Colors.red` (icon/text) | `negativeHighlightTextColor` | Error icon/text |
| `Colors.blue` (button) | `primaryColor` | Primary button |
| `Colors.blue[50]` (bg) | `infoHighlightColor` | Info section bg |
| `Colors.green` (icon/text) | `positiveHighlightTextColor` | Success/active icon |
| `Colors.orange` (icon/text) | `warningHighlightTextColor` | Warning icon |
| `Colors.deepPurple` / `purple` | `secondaryColor` | Category marker |

---

## Semantic Highlight Pattern

| Hardcoded | Replace With |
|-----------|-------------|
| `Colors.green.shade50` (section bg) | `positiveHighlightColor` |
| `Colors.green.shade200` (border) | `positiveHighlightTextColor.withValues(alpha:0.3)` |
| `Colors.green.shade600/700` (text) | `positiveHighlightTextColor` |
| `Colors.blue.shade50` (bg) | `infoHighlightColor` |
| `Colors.blue.shade200` (border) | `infoHighlightTextColor.withValues(alpha:0.3)` |
| `Colors.red.shade50` (bg) | `negativeHighlightColor` |
| `Colors.orange.shade50` (bg) | `warningHighlightColor` |

---

## Common Dark Mode Pitfalls

### listObject() textStyle missing color
Screens in `screens/config/` and `screen_search/` often omit color — invisible in dark mode:
```dart
// BAD
TextStyle textStyle = TextStyle(fontSize: global.deviceConfig.listDataFontSize);
// GOOD
TextStyle textStyle = TextStyle(
  fontSize: global.deviceConfig.listDataFontSize,
  color: (selected) ? global.theme.textColor : global.theme.textSecondaryColor,
);
```

### Scaffold missing backgroundColor
Many screens omit `backgroundColor` — stays white in dark mode. Set `global.theme.backgroundColor`.

### Processing/Utility screens (rebuild, audit, GL)
| Pattern | Hardcoded | Replace |
|---------|-----------|---------|
| Action card icon | `Colors.orange[700]` | `warningHighlightTextColor` |
| Action card bg | `Colors.orange[50]` | `warningHighlightColor` |
| Success status | `Colors.green` | `positiveHighlightTextColor` |
| Error text/icon | `Colors.red[700]` | `negativeHighlightTextColor` |
| Info card bg | `Colors.blue[50]` | `infoHighlightColor` |
| Error dialog bg | `Colors.red[50]` | `negativeHighlightColor` |

### Section Card Pattern
```dart
// BAD
Container(color: Colors.grey.shade50, border: Border.all(color: Colors.grey.shade300))
// GOOD
Container(color: global.theme.surfaceColor, border: Border.all(color: global.theme.dividerBorderColor))
```
