---
name: theming
description: มาตรฐาน Dark/Light Theme สำหรับ Flutter — สี, ฟอร์ม, data list, navigation — ใช้เมื่อสร้างหน้าจอใหม่ (ตรวจสีให้ถูก theme), แก้ dark mode ที่แสดงผลผิด, ค้นหา hardcoded colors (Colors.white/black/grey, Color.fromARGB, Color(0x...)), เพิ่มสีใหม่ใน ThemeModel, ตรวจสอบว่าจอรองรับ dark/light mode, หรือเมื่อพูดถึง theme, dark mode, สี, color, backgroundColor, textColor ใดๆ ในระบบ Flutter
user-invocable: true
---

# Theming — มาตรฐานระบบธีม Dark/Light Mode

## วิธีใช้
`/theming` — ตรวจสอบและแก้ไขสีให้ตรงตาม theme มาตรฐาน
`/theming <screen-name>` — ตรวจจอเฉพาะ
`/theming check` — scan ทั้ง project หาสี hardcode

## เมื่อไหร่ที่ใช้ Skill นี้
- สร้างหน้าจอใหม่ — ใช้สีจาก global.theme เท่านั้น
- แก้ไขสีให้ถูก dark/light mode
- ตรวจหาสี hardcode ที่ยังเหลืออยู่
- เพิ่มสีมาตรฐานใหม่เข้า ThemeModel

## สถาปัตยกรรม Theme ของ Project

### ไฟล์หลัก
| ไฟล์ | หน้าที่ |
|------|--------|
| `lib/model/theme_model.dart` | โมเดลเก็บ properties สีทั้งหมด |
| `lib/global.dart` → `themeSelect()` | ตั้งค่าสี Light mode (brand-based) |
| `lib/global.dart` → `applyThemeMode()` | Override สี Dark mode |
| `lib/global.dart` → `isDarkMode()` | ตรวจว่าตอนนี้ dark หรือไม่ |
| `lib/global.dart` → `saveThemeSettings()` | บันทึก theme ลง SharedPreferences |
| `lib/main_app.dart` | ThemeData (dark/light) + InputDecorationTheme |

### Flow การทำงาน
```
app start → themeSelect(brand) → applyThemeMode() → displaySettingsNotifier → MaterialApp rebuild
                                    ↑
user toggle → saveThemeSettings() ──┘
```

### 3 โหมด Theme
- `auto` — ตามระบบ (PlatformDispatcher.instance.platformBrightness)
- `light` — บังคับสว่าง
- `dark` — บังคับมืด

### Brand Theme (orthogonal กับ dark/light)
- Mode 0 = Default (สีน้ำเงิน #2A6F97)
- Mode 1 = DoHome (สีน้ำเงินเข้ม #235396)
- Dark mode override ทำงานหลัง brand → เปลี่ยนสีพื้น/ตัวอักษร ไม่แตะ primary

---

## 1. Color Properties ทั้งหมด (global.theme.xxx)

### Base Colors (มีมาก่อน)
| Property | Light | Dark | ใช้กับ |
|----------|-------|------|--------|
| `backgroundColor` | grey[50] | #1A1A2E | Scaffold background |
| `appBarColor` | #012A4A | #0D1B2A | AppBar |
| `primaryColor` | #2A6F97 | (ไม่เปลี่ยน) | ปุ่ม, accent, info icon/text |
| `primaryLightColor` | #89C2D9 | (ไม่เปลี่ยน) | hover, light accent |
| `secondaryColor` | white | #A9D6E5 | secondary surfaces, purple/accent marker |
| `headTitleColor` | white | white | AppBar title |
| `columnHeaderColor` | #89C2D9 | #2D3A4A | header แถว data list |
| `columnHeaderTextColor` | black | #E0E0E0 | ตัวอักษร header |
| `columnAlternateEvenColor` | #F3F7FA | #1E1E30 | แถวคู่ |
| `columnAlternateOddColor` | white | #1A1A2E | แถวคี่ |
| `inputTextBoxForceColor` | #8A1606 | #FF6B6B | label บังคับ |
| `inputTextBoxColor` | black | #E0E0E0 | label ไม่บังคับ |
| `toolBarEditModeColor` | #2A6F97 | #1B3A5C | toolbar ตอน edit |
| `buttonDangerColor` | #DC2626 | #EF4444 | ปุ่ม delete/danger (ElevatedButton bg) |

### UI Colors (เพิ่มใหม่)
| Property | Light | Dark | ใช้กับ |
|----------|-------|------|--------|
| `cardColor` | white | #252538 | Card, Container background |
| `textColor` | #212121 | #E0E0E0 | ตัวอักษรหลัก |
| `textSecondaryColor` | grey.600 | grey.400 | ตัวอักษรรอง |
| `surfaceColor` | grey.100 | #252538 | tab header, container surface |
| `searchBarColor` | grey.200 | #141825 | search bar background (ทึมกว่า surfaceColor) |
| `dividerBorderColor` | grey.300 | grey.700 | เส้นขอบ, divider |
| `dialogColor` | white | #2D2D2D | Dialog background |
| `inputFillColor` | white | #2D2D2D | Input field background |
| `bodyGradientColors` | [white, primaryLight, white] | [#1A1A2E, #16213E, #1A1A2E] | Body gradient |

### Icon Colors
| Property | Light | Dark | ใช้กับ |
|----------|-------|------|--------|
| `iconColor` | #424242 | #B0BEC5 | icon หลัก |
| `iconSecondaryColor` | grey.500 | grey.500 | icon รอง |

### Navigation & Bar
| Property | Light | Dark | ใช้กับ |
|----------|-------|------|--------|
| `infoBarColor` | #D4A373 (95%) | #0D1B2A (95%) | แถบข้อมูลร้าน |
| `bottomNavColor` | #D4A373 (95%) | #0D1B2A | bottom nav |
| `bottomNavSelectedColor` | white | white | nav item ที่เลือก |
| `bottomNavUnselectedColor` | white (60%) | white (60%) | nav item ที่ไม่เลือก |
| `onPrimaryColor` | white | white | ตัวอักษรบนพื้นเข้ม |
| `onPrimarySecondaryColor` | white70 | white70 | ตัวอักษรรองบนพื้นเข้ม |

### Data List
| Property | Light | Dark | ใช้กับ |
|----------|-------|------|--------|
| `rowHoverColor` | blue.50 | white (6%) | hover แถว |
| `rowSelectedColor` | cyan.100 | cyan (12%) | เลือกดู |
| `rowEditColor` | orange.100 | orange (15%) | กดแก้ไข |

### Form
| Property | Light | Dark | ใช้กับ |
|----------|-------|------|--------|
| `formFillColor` | white | #2D2D2D | พื้น input field |
| `formTextColor` | #212121 | #E0E0E0 | ตัวอักษรใน field |
| `formLabelColor` | black | #E0E0E0 | label ปกติ |
| `formLabelRequiredColor` | #8A1606 | #FF6B6B | label บังคับ |
| `formBorderColor` | grey.300 | grey.600 | เส้นขอบ field |
| `formFocusBorderColor` | #2A6F97 | #89C2D9 | เส้นขอบ focus |
| `formHintColor` | grey.400 | grey.500 | hint text |

---

## 2. กฎสำคัญ (MUST follow)

### ห้าม hardcode สี (ครอบคลุมทุกรูปแบบ)
```dart
// BAD — ทุกแบบห้ามหมด
color: Colors.white,
color: Colors.black,
color: Colors.grey[600],
color: Color.fromARGB(255, 148, 160, 194),   // ← ห้าม!
color: Color.fromRGBO(148, 160, 194, 1.0),   // ← ห้าม!
color: Color(0xFF94A0C2),                     // ← ห้าม!
fillColor: Colors.white,
hintStyle: TextStyle(color: Colors.grey[400]),

// GOOD — ใช้ global.theme.* เท่านั้น
color: global.theme.cardColor,
color: global.theme.textColor,
color: global.theme.textSecondaryColor,
fillColor: global.theme.formFillColor,
hintStyle: TextStyle(color: global.theme.formHintColor),
```

### ตาราง Color.fromARGB ที่พบบ่อย → แก้เป็น theme
| Color.fromARGB เดิม | สี | ใช้แทน |
|---------------------|-----|--------|
| `(255, 123, 235, 157)` | เขียวอ่อน (button) | `positiveHighlightTextColor` |
| `(255, 78, 141, 66)` / `(255, 67, 206, 74)` | เขียวเข้ม (button) | `positiveHighlightTextColor` |
| `(255, 164, 232, 162)` / `(255, 148, 194, 168)` | เขียวอ่อน (card bg) | `positiveHighlightColor` |
| `(255, 241, 112, 112)` / `(255, 246, 137, 129)` | แดง/salmon (button) | `buttonDangerColor` |
| `(255, 238, 86, 144)` | ชมพู (delete icon) | `negativeHighlightTextColor` |
| `(255, 235, 147, 123)` | ส้ม (warning button) | `warningHighlightTextColor` |
| `(255, 224, 159, 67)` | ส้ม (card bg) | `warningHighlightColor` |
| `(255, 148, 160, 194)` / `(255, 197, 212, 255)` | ฟ้า/ม่วงอ่อน (card bg) | `surfaceColor` หรือ `infoHighlightColor` |
| `(255, 99, 120, 183)` / `(255, 29, 43, 84)` | น้ำเงินเข้ม (button) | `primaryColor` |
| `(255, 168, 171, 136)` | มะกอก (selector button) | `primaryLightColor` / `surfaceColor` |
| `(255, 208, 42, 158)` | ม่วง/magenta | `secondaryColor` |
| `(255, 0, 0, 0)` | ดำ (text/icon) | `textColor` / `iconColor` |
| `(255, 255, 255, 255)` | ขาว (บนพื้นเข้ม) | `onPrimaryColor` |
| `(255, 75, 75, 75)` / `(255, 65, 66, 74)` | เทาเข้ม (button) | `primaryColor` |
| `(255, 152, 152, 152)` | เทากลาง | `textSecondaryColor` |
| `(255, 231, 206, 209)` | ชมพูอ่อน (card bg) | `surfaceColor` |

### เลือกสีถูกประเภท
| ต้องการ | ใช้ |
|---------|-----|
| พื้นหลัง Container/Card | `global.theme.cardColor` |
| พื้นหลัง search bar | `global.theme.searchBarColor` |
| พื้นหลัง body gradient | `global.theme.bodyGradientColors` |
| พื้นหลัง dialog | `global.theme.dialogColor` |
| ตัวอักษรหลัก | `global.theme.textColor` |
| ตัวอักษรรอง (hint, subtitle) | `global.theme.textSecondaryColor` |
| ตัวอักษรบน AppBar/info bar | `global.theme.onPrimaryColor` |
| Icon ทั่วไป | `global.theme.iconColor` |
| Icon รอง | `global.theme.iconSecondaryColor` |
| เส้นขอบ/divider | `global.theme.dividerBorderColor` |
| เส้นขอบ input field | `global.theme.formBorderColor` |
| พื้น input field | `global.theme.formFillColor` |
| Hover แถว data list | `global.theme.rowHoverColor` |
| เลือกดูแถว | `global.theme.rowSelectedColor` |
| แก้ไขแถว | `global.theme.rowEditColor` |

### ห้ามใช้สีเข้มเป็น text/accent บน dark background
```dart
// BAD — สีเข้มมองไม่เห็นใน dark mode
color: Colors.black87,
color: Colors.black54,
color: Colors.black,
color: Colors.grey[700],
color: Colors.grey[800],
color: global.theme.appBarColor,  // #0D1B2A ใน dark = มืดมาก!

// GOOD — ใช้ global.theme ที่ปรับตาม mode แล้ว
color: global.theme.textColor,           // หลัก
color: global.theme.textSecondaryColor,  // รอง (hint, subtitle)
```

### Accent Color ที่ปรับตาม theme (เช่น widget ที่ใช้สี appBarColor เป็น accent)
```dart
// BAD — appBarColor มืดเกินใน dark mode
final accentColor = global.theme.appBarColor;

// GOOD — สลับสีสว่างใน dark mode
final accentColor = global.isDarkMode()
    ? global.theme.primaryLightColor  // #89C2D9 (ฟ้าอ่อน)
    : global.theme.appBarColor;       // #012A4A (น้ำเงินเข้ม)
```

### ตารางเทียบสี (ครบทุก shade) — MUST ใช้ทุกครั้ง

| เดิม (hardcode) | ใช้แทน | ใช้กับ |
|-----------------|--------|--------|
| `Colors.white` | `global.theme.cardColor` | พื้น Card/Container ทั่วไป |
| `Colors.white` (form fill / fillColor) | `global.theme.formFillColor` | พื้น input field |
| `Colors.white` (dialog) | `global.theme.dialogColor` | พื้น Dialog |
| `Colors.white` บนพื้นเข้ม | `global.theme.onPrimaryColor` | text/icon บน AppBar, colored button |
| `Colors.black` | `global.theme.textColor` | ตัวอักษรหลัก |
| `Colors.black87` | `global.theme.textColor` | ตัวอักษรหลัก |
| `Colors.black54` | `global.theme.textSecondaryColor` | ตัวอักษรรอง |
| `Colors.grey[800]` | `global.theme.textColor` | ตัวอักษรหลัก |
| `Colors.grey[700]` | `global.theme.textSecondaryColor` | icon/label รอง |
| `Colors.grey[600]` | `global.theme.textSecondaryColor` | ตัวอักษรรอง / icon |
| `Colors.grey[500]` | `global.theme.iconSecondaryColor` | icon รอง |
| `Colors.grey[400]` | `global.theme.formHintColor` | hint text |
| `Colors.grey[300]` | `global.theme.dividerBorderColor` | เส้นขอบ / divider |
| `Colors.grey[200]` | `global.theme.searchBarColor` | พื้น search bar |
| `Colors.grey[200]` | `global.theme.surfaceColor` | พื้น surface (ไม่ใช่ search bar) |
| `Colors.grey[100]` | `global.theme.surfaceColor` | พื้น section card / description box |
| `Colors.grey[50]` | `global.theme.surfaceColor` | พื้น section card อ่อน |
| `Color(0xFFF8FAFC)` | `global.theme.surfaceColor` | custom light grey |
| `Colors.red` (button bg) | `global.theme.buttonDangerColor` | ElevatedButton ลบ/danger |
| `Colors.red` (icon/text) | `global.theme.negativeHighlightTextColor` | icon ลบ, ตัวอักษร error |
| `Colors.blue` (button bg) | `global.theme.primaryColor` | ElevatedButton primary |
| `Colors.blue` (icon/text/accent) | `global.theme.primaryColor` | icon info, accent |
| `Colors.blue[50]` (bg) | `global.theme.infoHighlightColor` | section bg info |
| `Colors.blue[200]` (border) | `global.theme.infoHighlightTextColor.withValues(alpha:0.3)` | section border info |
| `Colors.blue[700/800]` | `global.theme.primaryColor` | text accent |
| `Colors.green` (icon/text) | `global.theme.positiveHighlightTextColor` | icon สำเร็จ, active |
| `Colors.orange` (icon/text/bg) | `global.theme.warningHighlightTextColor` | warning icon/snackbar |
| `Colors.deepPurple` / `Colors.purple` | `global.theme.secondaryColor` | decoration, category marker |

### Section Card Pattern (สำคัญมาก — วนซ้ำบ่อย)

Section card = Container ที่รวม fields หลายตัว เช่น "ข้อมูลพนักงาน", "LINE OA", "approval section"

```dart
// BAD — hardcoded grey → dark mode เป็นสีเทาอ่อน มองยาก
Container(
  decoration: BoxDecoration(
    color: Colors.grey.shade50,          // ← ผิด
    border: Border.all(color: Colors.grey.shade300),  // ← ผิด
  ),
  child: Column(children: [
    Icon(Icons.badge, color: Colors.grey.shade700),   // ← ผิด
    Text('หัวข้อ', style: TextStyle(color: Colors.grey.shade700)), // ← ผิด
    // description box
    Container(color: Colors.grey.shade100, ...) // ← ผิด
  ])
)

// GOOD — ใช้ global.theme → dark/light mode ถูกทั้งคู่
Container(
  decoration: BoxDecoration(
    color: global.theme.surfaceColor,
    border: Border.all(color: global.theme.dividerBorderColor),
  ),
  child: Column(children: [
    Icon(Icons.badge, color: global.theme.textSecondaryColor),
    Text('หัวข้อ', style: TextStyle(color: global.theme.textSecondaryColor)),
    // description box
    Container(color: global.theme.surfaceColor, ...)
  ])
)
```

### Semantic Highlight Background Pattern

สำหรับ section ที่ต้องการสีสะท้อนความหมาย (success/info/warning/error):

```dart
// BAD — shade50 มองไม่เห็นใน dark mode
color: Colors.green.shade50,
border: Border.all(color: Colors.green.shade200),
text: TextStyle(color: Colors.green.shade600),

// GOOD — ใช้ positiveHighlightColor ที่ปรับตาม dark/light แล้ว
color: global.theme.positiveHighlightColor,
border: Border.all(color: global.theme.positiveHighlightTextColor.withValues(alpha: 0.3)),
text: TextStyle(color: global.theme.positiveHighlightTextColor),
```

| เดิม (hardcode) | ใช้แทน |
|-----------------|--------|
| `Colors.green.shade50` (section bg) | `global.theme.positiveHighlightColor` |
| `Colors.green.shade200` (section border) | `global.theme.positiveHighlightTextColor.withValues(alpha:0.3)` |
| `Colors.green.shade600/700` (section text) | `global.theme.positiveHighlightTextColor` |
| `Colors.blue.shade50` (section bg) | `global.theme.infoHighlightColor` |
| `Colors.blue.shade200` (section border) | `global.theme.infoHighlightTextColor.withValues(alpha:0.3)` |
| `Colors.red.shade50` (section bg) | `global.theme.negativeHighlightColor` |
| `Colors.orange.shade50` / `Colors.amber.shade50` (section bg) | `global.theme.warningHighlightColor` |

**ข้อยกเว้น — ใช้ Colors.green ตรงๆ ได้ (semantic icon/status):**
- icon สถานะ: `const Icon(Icons.check_circle, color: Colors.green)`
- avatar placeholder เล็กๆ: `color: Colors.green.shade100` ใน CircleAvatar
- badge/chip สี: `backgroundColor: Colors.green`

### ไฟล์/สีที่ยกเว้น (ห้ามเปลี่ยน)

**Directories ยกเว้น:**
- `pdfgen/` — PDF generation ใช้สีคงที่ (print สีต่างจากหน้าจอ)
- `usersystem/` — Login/register screens มี design แยก
- `global.dart` — Theme definition เอง ต้องใช้ Colors จริง

**Patterns ยกเว้น (ห้ามแก้):**
- `Colors.black.withValues(alpha: 0.01-0.5)` — shadow/barrier/overlay
- `Colors.transparent` — ไม่ต้องเปลี่ยน
- `PdfColors.*` — PDF generation (ไม่ใช่ UI)
- `Color colorSelected = Colors.white` — color picker initial value (data ไม่ใช่ UI)
- Commented-out code — ไม่ต้องแก้

**Patterns ที่ต้องแก้ (พบบ่อยมาก):**
- `Colors.white` เป็น `foregroundColor` บน button → `global.theme.onPrimaryColor`
- `Colors.white` เป็น icon color บน AppBar/colored bg → `global.theme.onPrimaryColor`
- `Colors.white` เป็น text บน gradient header → `global.theme.onPrimaryColor`
- `Colors.white.withValues(alpha: 0.7)` บน header → `global.theme.onPrimaryColor.withValues(alpha: 0.7)`
- `const Icon(..., color: Colors.white)` → `Icon(..., color: global.theme.onPrimaryColor)` (ลบ const!)

**ไม่มีข้อยกเว้น — ทุก color ต้องผ่าน global.theme เสมอ**

| เดิม (hardcode) | ใช้แทน | ใช้กับ |
|-----------------|--------|--------|
| `Colors.red` (ปุ่มลบ/danger bg) | `global.theme.negativeHighlightTextColor` | ปุ่ม delete, error button bg |
| `Colors.red` (icon/text) | `global.theme.negativeHighlightTextColor` | icon ลบ, ตัวอักษร error |
| `Colors.green` (สถานะ active) | `global.theme.positiveHighlightTextColor` | icon สำเร็จ, active status |
| `Colors.blue` (info/primary) | `global.theme.primaryColor` | ปุ่ม primary, accent |
| `Colors.orange` (warning) | `global.theme.warningHighlightTextColor` | icon คำเตือน |
| `Colors.deepPurple` / `Colors.purple` | `global.theme.secondaryColor` | category, tag |

### ถ้าจำเป็นต้องใช้สีต่างกันตาม theme
```dart
// สำหรับสีที่ไม่มีใน global.theme
color: global.isDarkMode() ? Colors.green[300] : Colors.green[700],
```

### ThemeRefreshMixin — บังคับทุก StatefulWidget

ทุก `StatefulWidget` ที่แสดงสีจาก `global.theme.*` **ต้อง** ใช้ `ThemeRefreshMixin`:
```dart
class _MyScreenState extends State<MyScreen> with global.ThemeRefreshMixin {
  // ThemeRefreshMixin จะ listen displaySettingsNotifier
  // เมื่อเปลี่ยนธีม → setState() อัตโนมัติ → rebuild ด้วยสีใหม่
}
```

**Flow:**
```
user เปลี่ยนธีม → saveThemeSettings() → applyThemeMode()
→ displaySettingsNotifier.value++ → ThemeRefreshMixin.setState()
→ ทุกจอ rebuild ด้วยสี global.theme ใหม่
```

ถ้าไม่ใส่ mixin → จอจะไม่ rebuild เมื่อเปลี่ยนธีม → สียังเป็นโหมดเก่า

### ระวัง const (Pitfall สำคัญ)
`global.theme.*` ไม่ใช่ compile-time constant — ห้ามใช้ใน `const` constructor:
```dart
// BAD — compile error: invalid_constant
const BoxDecoration(color: global.theme.cardColor)
const TextStyle(color: global.theme.textColor)
const Icon(Icons.search, color: global.theme.iconColor)

// GOOD — ลบ const ออก
BoxDecoration(color: global.theme.cardColor)
TextStyle(color: global.theme.textColor)
Icon(Icons.search, color: global.theme.iconColor)
```

**ข้อควรระวัง:** Widget ที่เป็น child ของ `const` parent ก็ต้องลบ `const` ด้วย

### ระวัง Color[] operator (Pitfall)
`global.theme.*` เป็น `Color` ไม่ใช่ `MaterialColor` — ห้ามใช้ `[shade]`:
```dart
// BAD — Color ไม่มี [] operator
global.theme.primaryColor[700]
global.theme.primaryColor[50]

// GOOD
global.theme.primaryColor       // ใช้สีหลักตรงๆ
global.theme.primaryLightColor  // ใช้สีอ่อน
```

### onPrimaryColor ห้ามใช้เป็น fillColor/background (Pitfall สำคัญมาก)

`onPrimaryColor` = white ทั้ง dark/light mode — เป็นสีสำหรับ text/icon บนพื้นเข้มเท่านั้น
ถ้าเอาไปใส่เป็น fillColor ของ form field → dark mode จะได้ช่อง input สีขาว!

```dart
// BAD — onPrimaryColor = white ทั้ง 2 mode → form field ขาวจ้าใน dark mode
fillColor: readOnly ? global.theme.surfaceColor : global.theme.onPrimaryColor,

// GOOD — formFillColor = white (light) / #2D2D2D (dark)
fillColor: readOnly ? global.theme.surfaceColor : global.theme.formFillColor,
```

**ตารางเลือกให้ถูก:**
| ต้องการ | ใช้ | ห้ามใช้ |
|---------|-----|---------|
| พื้น form field | `formFillColor` | `onPrimaryColor` |
| พื้น Card/Container | `cardColor` | `onPrimaryColor` |
| พื้น Dialog | `dialogColor` | `onPrimaryColor` |
| text/icon บน AppBar | `onPrimaryColor` | ✅ ถูกแล้ว |
| text/icon บนปุ่มสี | `onPrimaryColor` | ✅ ถูกแล้ว |
| icon ใน SnackBar | `Colors.white` ได้ | SnackBar มีพื้นสีเข้มเสมอ |

### Colors.white ใน SnackBar / Notification (ไม่ต้องแก้)

Icon ใน SnackBar เช่น `Icon(Icons.check, color: Colors.white)` **ไม่ต้องแก้**
เพราะ SnackBar มี backgroundColor เข้มเสมอ (ทั้ง dark/light) → white icon ถูกต้อง

### floatingLabelStyle backgroundColor (Pitfall สำคัญ — Dark Mode)

`floatingLabelStyle` ใน `TextFormField` มี `backgroundColor` เพื่อสร้างช่องว่างบน border
ถ้า hardcode `Colors.white` → dark mode จะเห็นสี่เหลี่ยมขาวบน label:
```dart
// BAD — ใน dark mode จะเห็นแถบสีขาวหลัง label
floatingLabelStyle: TextStyle(backgroundColor: Colors.white)

// GOOD — ใช้สี card ที่ปรับตาม theme
floatingLabelStyle: TextStyle(backgroundColor: global.theme.cardColor)
```

**ไฟล์ shared ที่ต้องตรวจ:**
- `lib/utils/textfield_custom.dart` — component หลักที่ทุก form ใช้
- `lib/utils/date_picker.dart`
- `lib/utils/time_picker.dart`

---

## 2.5 Common Dark Mode Pitfalls (พบบ่อย — ต้อง scan ทุกครั้ง)

### listObject() textStyle ไม่มี color (พบบ่อยมาก)
ไฟล์ใน `screens/config/` และ `screen_search/` มักมี `listObject()` ที่สร้าง textStyle โดยไม่ใส่ color → dark mode มองไม่เห็น

```dart
// BAD — ไม่มี color → ใช้ default black → มองไม่เห็นใน dark mode
TextStyle textStyle = TextStyle(
  fontWeight: (selected) ? FontWeight.bold : FontWeight.normal,
  fontSize: global.deviceConfig.listDataFontSize,
);

// GOOD — ต้องมี color เสมอ
TextStyle textStyle = TextStyle(
  fontWeight: (selected) ? FontWeight.bold : FontWeight.normal,
  fontSize: global.deviceConfig.listDataFontSize,
  color: (selected) ? global.theme.textColor : global.theme.textSecondaryColor,
);
```

**ไฟล์ที่ต้องตรวจ:** ทุกไฟล์ใน `screens/config/` และ `screen_search/` ที่มี `listObject()` function

### Text widget ใน list ไม่มี style (พบบ่อยใน screen_search)
```dart
// BAD — Text ไม่มี style → มองไม่เห็นใน dark mode
Text(value.code ?? "")
Text(global.packName(value.names!))

// GOOD — ใส่ style ที่มี color
Text(value.code ?? "", style: textStyle)
Text(global.packName(value.names!), style: textStyle)
```

### Drag-and-drop / Selection hardcoded colors
หน้าจอที่มี drag-and-drop tree (เช่น product_category_screen) มักใช้สีแบบนี้:

```dart
// BAD — hardcode สี state ต่างๆ
Color color = Colors.white;                    // default row
if (isDragTarget) color = Colors.green;        // drag target
if (isSelected) color = Colors.blue;           // selected
if (isInvalidDrag) color = Colors.red;         // invalid drag

// GOOD — ใช้ theme highlight colors
Color color = global.theme.cardColor;                      // default
if (isDragTarget) color = global.theme.positiveHighlightColor;   // drag target
if (isSelected) color = global.theme.infoHighlightColor;         // selected
if (isInvalidDrag) color = global.theme.negativeHighlightColor;  // invalid
```

### Processing/Utility Screen (rebuild, audit, GL process)
หน้าจอประมวลผล มักมี pattern เฉพาะที่ต้องแก้:

```dart
// BAD — Action card สี hardcode
Container(
  color: Colors.orange[50],
  child: Column(children: [
    Icon(Icons.build, color: Colors.orange[700]),
    Text('ประมวลผล', style: TextStyle(color: Colors.orange[800])),
  ]),
)

// GOOD — ใช้ warning highlight
Container(
  color: global.theme.warningHighlightColor,
  child: Column(children: [
    Icon(Icons.build, color: global.theme.warningHighlightTextColor),
    Text('ประมวลผล', style: TextStyle(color: global.theme.warningHighlightTextColor)),
  ]),
)
```

**Pattern ที่พบบ่อยใน processing screens:**

| Pattern | เดิม | ใช้แทน |
|---------|------|--------|
| Action card icon | `Colors.orange[700]` | `warningHighlightTextColor` |
| Action card bg | `Colors.orange[50]` | `warningHighlightColor` |
| Success status icon | `Colors.green` | `positiveHighlightTextColor` |
| Error status icon/text | `Colors.red` / `Colors.red[700]` | `negativeHighlightTextColor` |
| Info card bg | `Colors.blue[50]` | `infoHighlightColor` |
| Progress indicator | `Colors.orange[700]` | `warningHighlightTextColor` |
| Error dialog bg | `Colors.red[50]` | `negativeHighlightColor` |
| Error dialog border | `Colors.red[200]` | `negativeHighlightTextColor.withValues(alpha:0.3)` |
| Log error text | `Colors.red[700]` | `negativeHighlightTextColor` |
| AppBar deepOrange/custom | `Colors.deepOrange` | `global.theme.appBarColor` |

### Scaffold backgroundColor (พบบ่อย — ลืมตั้ง)
หลายจอไม่ set backgroundColor ให้ Scaffold → dark mode ยังเป็นพื้นขาว

```dart
// BAD — ไม่ตั้ง backgroundColor
Scaffold(
  appBar: AppBar(...),
  body: ...
)

// GOOD — ตั้ง backgroundColor เสมอ
Scaffold(
  backgroundColor: global.theme.backgroundColor,
  appBar: AppBar(backgroundColor: global.theme.appBarColor),
  body: ...
)
```

### วิธี scan หา pattern เหล่านี้
```bash
# หา listObject ที่ textStyle ไม่มี color
grep -n "TextStyle textStyle" lib/screens/config/*.dart lib/screen_search/*.dart
# แล้วตรวจว่าแต่ละตัวมี color: หรือไม่

# หา Text widget ที่ไม่มี style ใน listObject
grep -n "Text(" lib/screen_search/*.dart | grep -v "style:"

# หา Scaffold ที่ไม่มี backgroundColor
grep -n "Scaffold(" lib/screens/**/*.dart
# แล้วตรวจว่ามี backgroundColor: หรือไม่

# หา processing screens ที่ยังมี hardcode
grep -rn "Colors\.\(orange\|deepOrange\|red\|green\|blue\)" lib/screens/config/rebuild*.dart lib/screens/audit*.dart lib/screens/gl/*.dart
```

---

## 3. _getContainerColor มาตรฐาน (Data List)

ทุก data list screen ต้องใช้ pattern นี้ — **ตรวจว่า screen ใช้ตัวแปรไหน**:

```dart
Color? _getContainerColor(String itemGuid, int index) {
  // แถวที่เลือก — แยก edit กับ view
  if (selectGuid.isNotEmpty && selectGuid == itemGuid) {
    // ⚠️ ใช้ตัวแปรที่ screen นั้นมีจริง:
    // - ถ้ามี bool isEditMode → ใช้ isEditMode
    // - ถ้าไม่มี isEditMode แต่มี isSaveAllow → ใช้ isSaveAllow
    // - ห้ามใช้ screenEvent == edit (screenEvent ไม่ถูก set เป็น edit ใน switchToEdit)
    return isEditMode  // หรือ isSaveAllow
        ? global.theme.rowEditColor      // ส้ม = กำลังแก้ไข
        : global.theme.rowSelectedColor; // ฟ้า = เลือกดู
  }
  // hover
  if (_hoverIndex == index) {
    return global.theme.rowHoverColor;
  }
  // alternate rows
  return (index % 2 == 0)
      ? global.theme.columnAlternateEvenColor
      : global.theme.columnAlternateOddColor;
}
```

**⚠️ Pitfall สำคัญ:** `screenEvent == global.ScreenEventEnum.edit` ใน `_getContainerColor` จะ **ไม่ทำงาน** เพราะ `switchToEdit()` ไม่ set `screenEvent = edit` — ใช้ `isEditMode` หรือ `isSaveAllow` แทนเสมอ

---

## 4. Search Bar มาตรฐาน

Search bar ใช้ `searchBarColor` เป็น background (ทึมกว่า surfaceColor เพื่อให้เห็นชัดใน dark mode):

### แบบ Flat (แนะนำ — ใช้ใน barcode screen, config screens)
```dart
Container(
  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
  color: global.theme.searchBarColor,  // ← ใช้ searchBarColor เสมอ
  child: Row(
    children: [
      Expanded(
        child: TextField(
          style: TextStyle(color: global.theme.textColor),
          decoration: InputDecoration(
            border: InputBorder.none,
            filled: false,  // ← false เพราะ Container เป็นพื้นแล้ว
            hintText: global.language('search'),
            hintStyle: TextStyle(color: global.theme.formHintColor),
            prefixIcon: Icon(Icons.search, size: 20, color: global.theme.formHintColor),
          ),
        ),
      ),
      // font size + line spacing buttons (ถ้ามี)
    ],
  ),
),
```

### แบบ Rounded (ใช้ใน popup/dialog selection)
```dart
Container(
  padding: const EdgeInsets.all(5),
  color: global.theme.searchBarColor,
  child: Padding(
    padding: const EdgeInsets.only(left: 10, right: 10),
    child: TextFormField(
      style: TextStyle(color: global.theme.textColor),
      decoration: InputDecoration(
        isDense: true,
        border: InputBorder.none,
        filled: false,
        hintText: global.language('search'),
        hintStyle: TextStyle(color: global.theme.formHintColor),
        prefixIcon: Icon(Icons.search, size: 20, color: global.theme.formHintColor),
      ),
    ),
  ),
),
```

**สำคัญ:**
- ใช้ `searchBarColor` ไม่ใช่ `surfaceColor` — searchBarColor ทึมกว่า เห็นชัดใน dark mode
- `filled: false` เพราะ outer Container มีสีพื้นแล้ว (ถ้า filled: true จะซ้อนสี)
- ต้องมี `style: TextStyle(color: global.theme.textColor)` → dark mode เห็นตัวอักษร
- ต้องมี `hintStyle: TextStyle(color: global.theme.formHintColor)` → hint ไม่หาย

### Pitfall: Search bar ไม่เปลี่ยนตามธีม
ถ้า search bar ไม่เปลี่ยนสีเมื่อสลับ dark/light mode:
1. ตรวจว่า StatefulWidget ใช้ `ThemeRefreshMixin` หรือยัง
2. ตรวจว่าไม่มี hardcoded color ใน Container/BoxDecoration
3. ตรวจว่าไม่มี `const` ครอบ Container ที่ใช้ global.theme

---

## 5. Edit Form Body มาตรฐาน

```dart
SingleChildScrollView(
  child: Container(
    color: global.theme.cardColor,  // ไม่ใช่ Colors.white
    width: double.infinity,
    padding: const EdgeInsets.all(10),
    child: Column(
      children: [
        TextFormField(
          decoration: InputDecoration(
            border: const OutlineInputBorder(),
            enabledBorder: OutlineInputBorder(
              borderSide: BorderSide(color: global.theme.formBorderColor),
            ),
            floatingLabelBehavior: FloatingLabelBehavior.always,
            labelText: global.language("field_name"),
            labelStyle: TextStyle(color: global.theme.inputTextBoxForceColor),
          ),
        ),
      ],
    ),
  ),
),
```

---

## 6. เพิ่มสีใหม่เข้า Theme

เมื่อต้องการเพิ่มสีมาตรฐานใหม่:

1. **เพิ่ม property** ใน `lib/model/theme_model.dart`
2. **ตั้งค่า Light** ใน `global.dart` → `themeSelect()` (ท้ายฟังก์ชัน)
3. **ตั้งค่า Dark** ใน `global.dart` → `applyThemeMode()` (ใน `if (isDarkMode())`)
4. **ใช้** ผ่าน `global.theme.newProperty`

### หลักเลือกสี Dark Mode
- Background: ใช้ dark grey (#1A1A2E, #252538) ไม่ใช่ pure black
- Text: ใช้ #E0E0E0 ไม่ใช่ pure white (ลดแสงจ้า)
- Contrast ratio: อย่างน้อย 4.5:1 (WCAG AA)
- Accent/hover: ใช้ alpha transparency แทนสีทึบ → เนียนกับพื้นมากกว่า

---

## 7. Checklist ตรวจจอใหม่

### Scaffold & AppBar
- [ ] `Scaffold` มี `backgroundColor: global.theme.backgroundColor`
- [ ] `AppBar` มี `backgroundColor: global.theme.appBarColor` (ไม่ใช่ Colors.deepOrange หรือสี custom)
- [ ] AppBar title ใช้ `headTitleColor` / `onPrimaryColor`

### Background & Surface
- [ ] ไม่มี `Colors.white` เป็น background (ใช้ `cardColor`/`surfaceColor`)
- [ ] ไม่มี `Colors.grey.shade50/100` เป็น section card bg (ใช้ `surfaceColor`)
- [ ] ไม่มี `Colors.grey.shade300` เป็น section border (ใช้ `dividerBorderColor`)
- [ ] ไม่มี `Colors.green.shade50` เป็น section bg (ใช้ `positiveHighlightColor`)
- [ ] ไม่มี `Colors.blue.shade50` เป็น section bg (ใช้ `infoHighlightColor`)
- [ ] ไม่มี `Colors.orange.shade50` เป็น section bg (ใช้ `warningHighlightColor`)
- [ ] ไม่มี `Colors.red.shade50` เป็น section bg (ใช้ `negativeHighlightColor`)
- [ ] Search bar ใช้ `searchBarColor` (ไม่ใช่ surfaceColor)
- [ ] Edit form body ใช้ `cardColor`
- [ ] Dialog ใช้ `dialogColor`

### Text Colors (สำคัญมาก — dark mode จะมองไม่เห็น)
- [ ] ไม่มี `Colors.black` / `Colors.black87` เป็น text color (ใช้ `textColor`)
- [ ] ไม่มี `Colors.black54` เป็น text color (ใช้ `textSecondaryColor`)
- [ ] ไม่มี `Colors.grey[700]` / `Colors.grey[800]` เป็น text (ใช้ `textColor`)
- [ ] ไม่มี `Colors.grey[600]` เป็น text (ใช้ `textSecondaryColor`)
- [ ] ไม่ใช้ `appBarColor` เป็น accent/text color (ใช้ isDarkMode() สลับ)

### Icon & Border
- [ ] ไม่มี `Colors.grey[xxx]` เป็น icon color (ใช้ `iconColor`/`iconSecondaryColor`)
- [ ] Border/divider ใช้ `dividerBorderColor` หรือ `formBorderColor`

### Data List
- [ ] Row colors ใช้ `rowHoverColor`/`rowSelectedColor`/`rowEditColor`

### Form
- [ ] Form fields ใช้ `formFillColor`/`formBorderColor`/`formHintColor`
- [ ] `enabledBorder` ไม่ใช่ const (ถ้าใช้ global.theme)

### Processing/Utility Screens (rebuild, audit, GL)
- [ ] Action card icon/text ไม่ใช้ `Colors.orange[700]` (ใช้ `warningHighlightTextColor`)
- [ ] Progress indicator ไม่ใช้ `Colors.orange` (ใช้ `warningHighlightTextColor`)
- [ ] Status icon สำเร็จ/ผิดพลาด ใช้ highlight colors (ไม่ใช่ `Colors.green`/`Colors.red`)
- [ ] Error dialog/box ใช้ `negativeHighlightColor`/`negativeHighlightTextColor`
- [ ] Log text ใช้ `textColor` / `negativeHighlightTextColor` (ไม่ใช่ `Colors.red[700]`)

### Compile & Test
- [ ] ไม่มี `const` ครอบ widget ที่ใช้ global.theme
- [ ] ทดสอบทั้ง Light และ Dark mode

---

## 8. Batch Fix Workflow (`/theming check`)

วิธี scan ทั้ง project แล้วแก้ hardcoded colors:

### Step 1: Scan หา Colors. ทั้งหมด
```bash
# นับจำนวน hardcoded colors ใน screens/
grep -r "Colors\." --include="*.dart" lib/screens/ | grep -v "pdfgen\|usersystem\|global.dart" | wc -l
```

### Step 2: จัดกลุ่มตาม priority
1. **สูงสุด** — `Colors.white` เป็น background (dark mode ขาวจ้า)
2. **สูง** — `Colors.black`, `Colors.grey[600-800]` เป็น text (dark mode มองไม่เห็น)
3. **กลาง** — `Colors.grey[200-400]` เป็น border/surface
4. **ต่ำ** — Semantic colors (red/green/orange) — ส่วนใหญ่ OK

### Step 3: แก้เป็น batch (ใช้ subagent แยกตาม folder)
```
screens/config/      → agent 1 (28+ files, รวม rebuild*, product_category)
screens/master/      → agent 2 (11 files)
screens/transaction/ → agent 3 (18 files)
screens/report/      → agent 4
screens/gl/          → agent 5 (gl_process_screen)
screens/other/       → agent 6 (chatbot, coupon, form, etc.)
screen_search/       → agent 7 (listObject text color)
utils/               → manual fix (shared components)
```

### Step 4: ตรวจ compile
```bash
dart analyze lib/ 2>&1 | grep -c "error"
# ต้อง 0 errors
```

### ข้อควรระวังเมื่อแก้ batch
- `static const Color _xxx = Colors.white` → `Color get _xxx => global.theme.cardColor;` (ลบ static const)
- Widget ที่เป็น child ของ `const` → ลบ `const` ทุกตัวในสาย
- `Colors.white` ใน `master_*_screen.dart` header gradient → ใช้ `onPrimaryColor` (ขาวบนเข้ม)

---

## 9. bclms Project (ระบบ theme ต่างกัน)

bclms (`D:\bcdev\frontend\bclms`) ใช้ `AppTheme` แทน `global.theme`:
- ไม่มี `ThemeRefreshMixin`
- ไม่มี `global.theme.*` properties
- ใช้ `Theme.of(context).xxx` pattern
- ต้องแก้แยก — ห้ามใช้ skill นี้ตรงๆ กับ bclms

---

## Sources & Best Practices
- [Flutter Official: Use themes to share colors and font styles](https://docs.flutter.dev/cookbook/design/themes)
- [Flutter Dark Mode Guide: Theme, Toggle & Persistence](https://www.f22labs.com/blogs/how-to-implement-dark-mode-in-your-flutter-app/)
- [Design Harmony: A Comprehensive Theming Guide for Flutter](https://www.arhaminfo.com/2025/11/design-harmony-comprehensive-theming-guide-flutter.html)
- [Flutter Theme Management: Custom Color Schemes Made Easy](https://mobisoftinfotech.com/resources/blog/flutter-theme-management-custom-color-schemes)
- [Mastering Dynamic Theming in Flutter](https://medium.com/@and.santucci.97/mastering-dynamic-theme-in-flutter-override-theme-mode-and-user-custom-color-2dcc2cdaedca)
