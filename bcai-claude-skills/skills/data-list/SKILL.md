---
name: data-list
description: มาตรฐาน Data List Screen สำหรับ Flutter — โครงสร้าง SplitView, listScreen/editScreen, _getContainerColor 3 สถานะ, hover, search bar, column header, ThemeRefreshMixin — ใช้เมื่อสร้างหน้า config/master ใหม่, แก้ไข data list ที่แสดงผลผิด, ตรวจสอบ pattern ของจอ list+edit, หรือเมื่อพูดถึง data list, split view, row color, hover, list screen, edit screen
user-invocable: true
---

# Data List — มาตรฐานหน้าจอ List + Edit

## วิธีใช้
`/data-list` — ตรวจสอบ/สร้างหน้าจอ data list ให้ตรงมาตรฐาน
`/data-list <screen-name>` — ตรวจจอเฉพาะ

## ต้นแบบ (Reference Screen)
`lib/screens/config/product_barcode_screen.dart` — จอที่ครบทุก pattern

---

## 1. โครงสร้าง State Class

```dart
class _MyScreenState extends State<MyScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {

  // === Selection/Edit state ===
  int _hoverIndex = -1;
  bool isEditMode = false;     // Group A: มี isEditMode
  bool isSaveAllow = false;    // Group B: มี isSaveAllow เท่านั้น
  String selectGuid = "";
  bool showCheckBox = false;

  // === List management ===
  List<String> guidListChecked = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];

  // === UI controllers ===
  late SplitViewController splitViewController;
  late TabController tabController;
  final _debouncer = global.Debouncer(1000);
  final _formKey = GlobalKey<FormState>();
}
```

**ThemeRefreshMixin บังคับ** — ถ้าไม่ใส่ จอจะไม่ rebuild เมื่อเปลี่ยนธีม

---

## 2. Responsive Layout (SplitView / TabBarView)

```dart
@override
Widget build(BuildContext context) {
  return LayoutBuilder(
    builder: (context, constraints) {
      return Scaffold(
        backgroundColor: global.theme.backgroundColor,
        body: (constraints.maxWidth > 800)
            ? SplitView(
                controller: splitViewController,
                gripSize: 8,
                gripColor: global.theme.dividerBorderColor,
                gripColorActive: global.theme.primaryColor,
                viewMode: SplitViewMode.Horizontal,
                indicator: const SplitIndicator(
                  viewMode: SplitViewMode.Horizontal,
                ),
                activeIndicator: const SplitIndicator(
                  viewMode: SplitViewMode.Horizontal,
                  isActive: true,
                ),
                children: [
                  listScreen(mobileScreen: false),
                  editScreen(mobileScreen: false),
                ],
              )
            : TabBarView(
                physics: const NeverScrollableScrollPhysics(),
                controller: tabController,
                children: [
                  listScreen(mobileScreen: true),
                  editScreen(mobileScreen: true),
                ],
              ),
      );
    },
  );
}
```

| Property | ค่า | หมายเหตุ |
|----------|-----|----------|
| `gripColor` | `dividerBorderColor` | ไม่ใช่ appBarColor (มืดเกินใน dark) |
| `gripColorActive` | `primaryColor` | สีเข้มขึ้นตอนลาก |
| breakpoint | `800` | desktop vs mobile |

---

## 3. listScreen — หน้าจอฝั่งซ้าย (รายการ)

```dart
Widget listScreen({bool mobileScreen = false}) {
  return Scaffold(
    backgroundColor: global.theme.backgroundColor,  // <-- ใช้ backgroundColor
    appBar: AppBar(
      backgroundColor: global.theme.appBarColor,
      title: Text(global.language("my_title")),
      // ... actions
    ),
    body: Column(
      children: [
        // 1. Search Bar
        _buildSearchBar(),
        // 2. Separator
        Container(color: global.theme.appBarColor, height: 6),
        // 3. Column Header
        _buildColumnHeader(),
        // 4. Data List
        Expanded(child: _buildDataList()),
      ],
    ),
  );
}
```

### Search Bar
```dart
Container(
  padding: const EdgeInsets.all(5),
  decoration: BoxDecoration(
    color: global.theme.searchBarColor,
    borderRadius: BorderRadius.circular(2),
  ),
  child: Row(
    children: [
      Expanded(
        child: TextFormField(
          onChanged: (value) {
            _debouncer.run(() {
              setState(() => listDatas = []);
              loadDataList(value);
            });
          },
          decoration: InputDecoration(
            isDense: true,
            border: InputBorder.none,
            prefixIcon: Icon(Icons.search, size: 20),
            hintText: global.language('search'),
          ),
        ),
      ),
      ListFontSizeControl(onChanged: () => setState(() {})),
      const SizedBox(width: 4),
      if (listDatas.isNotEmpty)
        Text('(${listDatas.length})', style: TextStyle(fontSize: 11, color: global.theme.textSecondaryColor)),
    ],
  ),
)
```

### Column Header
```dart
Container(
  color: global.theme.columnHeaderColor,
  padding: const EdgeInsets.only(left: 10, right: 10, top: 5, bottom: 5),
  child: Row(
    children: [
      Expanded(
        flex: 5,
        child: Text(
          global.language("code"),
          style: TextStyle(
            color: global.theme.columnHeaderTextColor,
            fontWeight: FontWeight.bold,
            fontSize: global.deviceConfig.listDataFontSize + 2,
          ),
        ),
      ),
      // ... more columns
    ],
  ),
)
```

---

## 4. _getContainerColor — 3 สถานะสี (สำคัญมาก)

```dart
Color? _getContainerColor(String itemGuid, int index) {
  // 1. แถวที่เลือก — แยก edit กับ view
  if (selectGuid.isNotEmpty && selectGuid == itemGuid) {
    return isEditMode   // หรือ isSaveAllow (ดูตัวแปรที่ screen มี)
        ? global.theme.rowEditColor      // ส้ม = กำลังแก้ไข
        : global.theme.rowSelectedColor; // ฟ้า = เลือกดู
  }
  // 2. hover
  if (_hoverIndex == index) {
    return global.theme.rowHoverColor;
  }
  // 3. alternating rows
  return (index % 2 == 0)
      ? global.theme.columnAlternateEvenColor
      : global.theme.columnAlternateOddColor;
}
```

### ลำดับความสำคัญ
| ลำดับ | สถานะ | สี | Property |
|-------|--------|-----|----------|
| 1 | เลือก + แก้ไข | ส้ม | `rowEditColor` |
| 2 | เลือก + ดู | ฟ้า | `rowSelectedColor` |
| 3 | Hover | ฟ้าอ่อน | `rowHoverColor` |
| 4 | แถวคู่ | เทาอ่อน/เข้ม | `columnAlternateEvenColor` |
| 5 | แถวคี่ | ขาว/เข้ม | `columnAlternateOddColor` |

### ตัวแปร edit mode — ต้องเลือกให้ถูก

| Group | ตัวแปร | จอตัวอย่าง |
|-------|--------|-----------|
| A | `isEditMode` | bank, color, employee, kitchen, barcode, product, coupon |
| B | `isSaveAllow` | business_type, department, product_group, product_unit, master_* ทุกจอ |

**Pitfall:** ห้ามใช้ `screenEvent == global.ScreenEventEnum.edit` — `switchToEdit()` ไม่ set screenEvent เป็น edit

---

## 5. List Item — MouseRegion + GestureDetector + Container

```dart
Widget listObject(int index, MyModel value) {
  return MouseRegion(
    cursor: SystemMouseCursors.click,
    onEnter: (_) => setState(() => _hoverIndex = index),
    onExit: (_) => setState(() => _hoverIndex = -1),
    child: GestureDetector(
      onTap: () {
        setState(() {
          discardData(callBack: () {
            isSaveAllow = false;
            isEditMode = false;
            selectGuid = value.guidfixed;
            getData(selectGuid);
            WidgetsBinding.instance.addPostFrameCallback((_) {
              tabController.animateTo(1);
            });
          });
        });
      },
      onDoubleTap: () {
        if (!showCheckBox) switchToEdit(value);
      },
      child: Container(
        key: index < listKeys.length ? listKeys[index] : null,
        decoration: BoxDecoration(
          color: _getContainerColor(value.guidfixed, index),
        ),
        padding: const EdgeInsets.only(
          left: 10, right: 10,
          top: 5, bottom: 5,
        ),
        child: Row(
          children: [
            // data columns
          ],
        ),
      ),
    ),
  );
}
```

| Pattern | สิ่งที่ต้องทำ |
|---------|-------------|
| Single tap | เลือกดู (view mode) |
| Double tap | เข้า edit mode |
| Hover enter | `_hoverIndex = index` |
| Hover exit | `_hoverIndex = -1` |

---

## 6. editScreen — ฝั่งขวา (ฟอร์มแก้ไข)

```dart
Widget editScreen({bool mobileScreen = false}) {
  return Scaffold(
    backgroundColor: global.theme.cardColor,  // <-- cardColor ไม่ใช่ backgroundColor
    appBar: AppBar(
      backgroundColor: isEditMode
          ? global.theme.toolBarEditModeColor
          : global.theme.appBarColor,
      title: Text(headerEdit + global.language("my_title")),
      actions: [
        // ปุ่มลบ
        if (selectGuid.isNotEmpty)
          IconButton(
            icon: Icon(Icons.delete, color: global.theme.onPrimaryColor),
            onPressed: () => deleteData(),
          ),
        // ปุ่มแก้ไข (toggle)
        if (!isSaveAllow && selectGuid.trim().isNotEmpty)
          IconButton(
            icon: Icon(Icons.edit, color: global.theme.onPrimaryColor),
            onPressed: () => switchToEdit(currentData),
          ),
      ],
    ),
    body: SingleChildScrollView(
      child: Form(
        key: _formKey,
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(children: [
            // form fields...
            // Save button
            if (isSaveAllow)
              ElevatedButton.icon(
                style: ElevatedButton.styleFrom(
                  backgroundColor: global.theme.buttonColor,
                ),
                onPressed: () {
                  if (_formKey.currentState!.validate()) saveOrUpdateData();
                },
                icon: Icon(Icons.save),
                label: Text(global.language("save")),
              ),
          ]),
        ),
      ),
    ),
  );
}
```

### สีพื้นหลัง editScreen vs listScreen
| Screen | Property | ทำไม |
|--------|----------|------|
| listScreen | `backgroundColor` | ตรง grey tone กับ alternating rows |
| editScreen | `cardColor` | ขาว (light) / card dark (dark) — แยกให้เห็นว่าเป็น form |

### AppBar สี 2 สถานะ
| สถานะ | Property |
|--------|----------|
| View mode | `appBarColor` |
| Edit mode | `toolBarEditModeColor` |

---

## 7. Master Data Selector (ปุ่มเลือก master ใน edit form)

```dart
Widget _buildMasterDataSelector({
  required String label,
  required String code,
  required String displayName,
  required VoidCallback onTap,
  required VoidCallback onClear,
  bool isSelected = false,
}) {
  return Container(
    height: 36,
    decoration: BoxDecoration(
      border: Border.all(
        color: isSelected
            ? global.theme.primaryColor       // ไม่ใช่ appBarColor!
            : global.theme.textSecondaryColor,
      ),
      borderRadius: BorderRadius.circular(4),
      color: isSelected
          ? global.theme.primaryColor.withValues(alpha: 0.08)
          : null,
    ),
    child: Row(
      children: [
        Text('$label: ', style: TextStyle(
          color: global.theme.textSecondaryColor,
        )),
        Expanded(child: Text(
          isSelected ? '$code ~ $displayName' : '-',
          style: TextStyle(
            color: isSelected
                ? global.theme.textColor
                : global.theme.textSecondaryColor,
          ),
        )),
        if (isSelected && isEditMode)
          InkWell(
            onTap: onClear,
            child: Icon(Icons.clear,
              color: global.theme.negativeHighlightTextColor, size: 16),
          ),
        Icon(Icons.search,
          color: isEditMode
              ? global.theme.primaryColor     // ไม่ใช่ appBarColor!
              : global.theme.textSecondaryColor,
          size: 16,
        ),
      ],
    ),
  );
}
```

**Pitfall:** ห้ามใช้ `appBarColor` เป็น accent/border — ใน dark mode = `#0D1B2A` มืดมาก มองไม่เห็น → ใช้ `primaryColor` แทน

### ElevatedButton Selector (ปุ่มใหญ่เลือก master เช่น ประเภทสินค้า, ผู้ผลิต)

```dart
ElevatedButton(
  style: ButtonStyle(
    backgroundColor: WidgetStateProperty.all<Color>(
      (isDataSelected)
          ? global.theme.primaryColor.withValues(alpha: 0.08)  // tint อ่อน
          : global.theme.cardColor,                             // พื้นขาว/เข้ม
    ),
    foregroundColor: WidgetStateProperty.all<Color>(global.theme.textColor),
    elevation: WidgetStateProperty.all(0),          // flat ไม่มีเงา
    shape: WidgetStateProperty.all(
      RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(8),
        side: BorderSide(
          color: (isDataSelected)
              ? global.theme.primaryColor           // ขอบสี primary
              : global.theme.dividerBorderColor,    // ขอบเทา
        ),
      ),
    ),
  ),
  child: Row(children: [
    Icon(Icons.category, color: global.theme.primaryColor),
    const SizedBox(width: 8),
    Expanded(child: Text(...)),
    if (isDataSelected)
      IconButton(
        icon: Icon(Icons.clear, color: global.theme.negativeHighlightTextColor),
        onPressed: onClear,
      ),
    Icon(Icons.search, color: global.theme.textSecondaryColor),
  ]),
)
```

**หลักการ:** ปุ่ม selector ใหญ่ต้อง flat + border เหมือน form field — ห้ามใช้สีเข้มเป็น background (ดูหนักเกินไป)

---

## 8. Edit Font Scale — ปรับขนาด font ฝั่ง edit (A- % A+)

ฝั่ง edit (ขวา) รองรับ user ปรับ font scale 50%-300% ผ่านปุ่ม A-/A+ บน AppBar

### 8.1 AppBar — เพิ่ม EditFontSizeControl

```dart
import 'package:smlaicloud/widgets/edit_font_size_control.dart';

// ใน editScreen AppBar actions (ก่อนปุ่มอื่น)
actions: [
  EditFontSizeControl(onChanged: () => setState(() {})),
  // ... ปุ่ม delete, edit, save
],
```

### 8.2 Body wrapper — scale ทุกอย่าง

ครอบ body ของ editScreen ด้วย 4 layers:

```dart
body: LoaderOverlay(
  child: Builder(
    builder: (context) {
      final scale = global.editFontScaleFactor;
      // เพิ่ม spacing ระหว่าง formWidgets ตาม scale
      List<Widget> scaledFormWidgets = formWidgets;
      if (scale > 1.0) {
        final extraGap = 8.0 * (scale - 1);
        scaledFormWidgets = [];
        for (int i = 0; i < formWidgets.length; i++) {
          scaledFormWidgets.add(formWidgets[i]);
          if (i < formWidgets.length - 1) {
            scaledFormWidgets.add(SizedBox(height: extraGap));
          }
        }
      }
      return MediaQuery(
        data: MediaQuery.of(context).copyWith(
          textScaler: TextScaler.linear(scale),  // scale text ทุก child
        ),
        child: Theme(
          data: Theme.of(context).copyWith(
            inputDecorationTheme: Theme.of(context).inputDecorationTheme.copyWith(
              contentPadding: EdgeInsets.fromLTRB(
                12 * scale, 20 * scale, 12 * scale, 12 * scale,
              ),
            ),
          ),
          child: IconTheme(
            data: IconTheme.of(context).copyWith(
              size: 24.0 * scale,  // scale icon
            ),
            child: SingleChildScrollView(
              child: Padding(
                padding: EdgeInsets.only(
                  top: 10 * scale, bottom: 10 * scale,
                ),
                child: Form(
                  key: _formKey,
                  child: Column(children: scaledFormWidgets),
                ),
              ),
            ),
          ),
        ),
      );
    },
  ),
),
```

### 8.3 หลักการ Scale 4 layers

| Layer | ทำหน้าที่ | ทำไมต้องมี |
|-------|----------|-----------|
| `MediaQuery.textScaler` | scale text ทุก child widget | propagate ลง search fields, child widgets ทั้งหมด |
| `Theme.inputDecorationTheme` | scale contentPadding ของ InputDecoration | ป้องกัน floating label ทับ content ใน fields ที่ไม่ได้กำหนด contentPadding เอง |
| `IconTheme` | scale icon size | icon ต้องใหญ่ตาม text |
| `SizedBox` spacing | เพิ่ม gap ระหว่าง formWidgets | ป้องกัน floating label ของ field ล่างทับ field บน |

### 8.4 Global helpers (มีแล้วใน global.dart)

```dart
// ลด font scale 5% (min 50%)
void editFontScaleDecrease()
// เพิ่ม font scale 5% (max 300%)
void editFontScaleIncrease()
// คืนค่า scale factor (100 → 1.0, 80 → 0.8)
double get editFontScaleFactor => deviceConfig.editFontScale / 100;
```

### 8.5 Pitfalls — วิธีที่ไม่ work

| วิธี | ปัญหา |
|------|--------|
| `Transform.scale` | ไม่ affect layout → scroll extent ผิด |
| `FittedBox + SizedBox` | text wrap ผิด (หน้าจอเล็กลงแล้ว zoom) |
| `DefaultTextStyle.merge` เฉยๆ | ไม่ propagate ลง child widgets (search fields) |
| `MediaQuery.textScaler` อย่างเดียว | text scale แต่ padding/spacing ไม่ scale → label ทับ |

### 8.6 Transaction Edit Screens (PO/QT/SO/Invoice)

Transaction edit screens (ใบสั่งซื้อ, ใบเสนอราคา, ใบขาย ฯลฯ) ใช้ pattern เดียวกัน แต่ต่างจาก data list ตรงที่:
- AppBar ใช้ helper static class (เช่น `POAppBarActions.build()`, `QTAppBarActions.build()`)
- Body เป็น nested Scaffold ที่มี TabBar (เอกสาร / รายการสินค้า)
- Form อยู่ใน `DocumentHeaderWidget` (shared component)

**วิธีเพิ่ม EditFontSizeControl:**

```dart
// 1. AppBar — ใส่ก่อน actions จาก helper (ใช้ spread)
actions: [
  EditFontSizeControl(onChanged: () => setState(() {})),
  ...POAppBarActions.build(...),  // หรือ QTAppBarActions.build(...)
],

// 2. Body — wrap TabBarView ของ inner Scaffold
child: Builder(
  builder: (context) {
    final scaleFactor = global.editFontScaleFactor;
    return MediaQuery(
      data: MediaQuery.of(context).copyWith(
        textScaler: TextScaler.linear(scaleFactor),
      ),
      child: Theme(
        data: Theme.of(context).copyWith(
          inputDecorationTheme: Theme.of(context).inputDecorationTheme.copyWith(
            contentPadding: EdgeInsets.symmetric(
              horizontal: 12 * scaleFactor,
              vertical: 10 * scaleFactor,
            ),
          ),
          iconTheme: IconThemeData(size: 24 * scaleFactor),
        ),
        child: TabBarView(...),  // เอกสาร / รายการสินค้า
      ),
    );
  },
),
```

**จอที่เพิ่มแล้ว (มี EditFontSizeControl):**
- `purchaseorder_edit_screen.dart` — ใบสั่งซื้อ
- `quotation_edit_screen.dart` — ใบเสนอราคา
- `transaction_edit.dart` — ใบขาย/ซื้อทั่วไป

**จอที่ยังไม่มี (ถ้า Jead ต้องการเพิ่ม):**
- `purchase_add.dart`, `product_add.dart`, `option_add.dart`, `option_detail.dart`
- `coupon_screen.dart`, `currency_form_screen.dart`
- `transaction_paid.dart`, `paidpayment_screen.dart`

---

## 9. Checklist ตรวจจอ Data List

- [ ] `State` class มี `global.ThemeRefreshMixin`
- [ ] `_getContainerColor()` ใช้ `isEditMode` หรือ `isSaveAllow` (ไม่ใช่ `screenEvent`)
- [ ] `_getContainerColor()` ครบ 3 สถานะ: edit/selected/hover/alternating
- [ ] listScreen `backgroundColor: global.theme.backgroundColor`
- [ ] editScreen `backgroundColor: global.theme.cardColor`
- [ ] AppBar edit mode ใช้ `toolBarEditModeColor`
- [ ] Search bar ใช้ `searchBarColor`
- [ ] Search bar มี `ListFontSizeControl` (A-/size/A+/≡) — **ห้ามมี `IconButton(Icons.line_weight)` ซ้ำ** และ **ห้าม inline font control** (เขียน A-/A+ เอง)
- [ ] Font size ใน list item ใช้ `global.deviceConfig.listDataFontSize` — ห้ามใช้ตัวแปร local เช่น `_listFontSize`
- [ ] Column header ใช้ `columnHeaderColor` + `columnHeaderTextColor`
- [ ] SplitView gripColor ใช้ `dividerBorderColor`
- [ ] ไม่มี Colors.* hardcode (ยกเว้น shadow/overlay/color picker data)
- [ ] Master selector ใช้ `primaryColor` ไม่ใช่ `appBarColor`
- [ ] ปุ่ม danger/delete ใช้ `buttonDangerColor` (ไม่ใช่ `negativeHighlightTextColor`)
- [ ] ElevatedButton selector ใช้ flat style (elevation: 0, border, cardColor bg)
- [ ] ไม่มี `Color.fromARGB` / `Color(0x...)` hardcode ในไฟล์
- [ ] MouseRegion + GestureDetector ครบ (hover, tap, doubleTap)
- [ ] Phone fields ไม่มี `maxLength` / `inputFormatters` (ให้พิมพ์อิสระ)

---

## 9. Pitfalls

| ปัญหา | สาเหตุ | วิธีแก้ |
|--------|--------|---------|
| ปุ่ม line spacing ซ้ำ 2 ตัว | `ListFontSizeControl` มีปุ่ม ≡ อยู่แล้ว + มี `IconButton(Icons.line_weight)` แยกอีกตัว | ลบ `IconButton(Icons.line_weight)` ออก — ใช้แค่ `ListFontSizeControl` ตัวเดียว |
| เบอร์โทร input ขึ้น error | `maxLength: 10` + `inputFormatters` จำกัดเฉพาะตัวเลข | ลบ `maxLength` + `inputFormatters` ออก — เบอร์โทรอาจมี `-`, `+`, ช่องว่าง |
| `_getContainerColor()` ไม่ทำงาน | มี method แต่ไม่ได้เรียกใช้ — Container ใช้ inline ternary แทน | ใช้ `color: _getContainerColor(value.guidfixed, index)` ใน Container |
| Row edit ไม่เปลี่ยนสีส้ม | ใช้ `screenEvent == edit` ตรวจ | ใช้ `isEditMode` หรือ `isSaveAllow` แทน |
| ธีมเปลี่ยนแล้วปุ่มไม่เปลี่ยนตาม | ใช้ `Color.fromARGB` / `Colors.*` hardcode | เปลี่ยนเป็น `global.theme.*` ทุกจุด |
| Font size control ไม่ sync ข้ามจอ | ใช้ตัวแปร local `_listFontSize` แทน global | ลบตัวแปร local + ใช้ `global.deviceConfig.listDataFontSize` + ใช้ `ListFontSizeControl` widget |
