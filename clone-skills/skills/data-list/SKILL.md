---
name: data-list
description: >
  Data list screen standards for Flutter ERP. SplitView layout, listScreen/editScreen,
  row colors (hover/selected/edit), search bar, column header, ThemeRefreshMixin.
  Triggers: "data list", "split view", "row color", "hover", "list screen", "edit screen".
user-invocable: true
---

# Data List — List + Edit Screen Standards

## Usage
- `/data-list` — Check/create data list screens to match standards
- `/data-list <screen-name>` — Check specific screen

## Reference Screen
`lib/screens/config/product_barcode_screen.dart` — covers all patterns

---

## 1. State Class

```dart
class _MyScreenState extends State<MyScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  int _hoverIndex = -1;
  bool isEditMode = false;     // Group A screens
  bool isSaveAllow = false;    // Group B screens
  String selectGuid = "";
  bool showCheckBox = false;
  List<String> guidListChecked = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  late SplitViewController splitViewController;
  late TabController tabController;
  final _debouncer = global.Debouncer(1000);
  final _formKey = GlobalKey<FormState>();
}
```

ThemeRefreshMixin is required — without it, screen won't rebuild on theme change.

---

## 2. Responsive Layout (SplitView / TabBarView)

```dart
return LayoutBuilder(
  builder: (context, constraints) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      body: (constraints.maxWidth > 800)
          ? SplitView(
              controller: splitViewController,
              gripColor: global.theme.dividerBorderColor,
              gripColorActive: global.theme.primaryColor,
              children: [listScreen(mobileScreen: false), editScreen(mobileScreen: false)],
            )
          : TabBarView(
              physics: const NeverScrollableScrollPhysics(),
              controller: tabController,
              children: [listScreen(mobileScreen: true), editScreen(mobileScreen: true)],
            ),
    );
  },
);
```

| Property | Value | Note |
|----------|-------|------|
| `gripColor` | `dividerBorderColor` | Not appBarColor (too dark in dark mode) |
| breakpoint | `800` | Desktop vs mobile |

---

## 3. listScreen (Left Panel)

Structure: AppBar > Column > [SearchBar, Separator, ColumnHeader, DataList]

- Scaffold `backgroundColor: global.theme.backgroundColor`
- Search bar uses `searchBarColor` + `ListFontSizeControl`
- Column header uses `columnHeaderColor` + `columnHeaderTextColor`
- Font size: `global.deviceConfig.listDataFontSize` (not local variables)

See [references/patterns.md](references/patterns.md) for full code snippets.

---

## 4. _getContainerColor — Row Color (3 States)

```dart
Color? _getContainerColor(String itemGuid, int index) {
  if (selectGuid.isNotEmpty && selectGuid == itemGuid) {
    return isEditMode // or isSaveAllow
        ? global.theme.rowEditColor      // orange = editing
        : global.theme.rowSelectedColor; // blue = viewing
  }
  if (_hoverIndex == index) return global.theme.rowHoverColor;
  return (index % 2 == 0)
      ? global.theme.columnAlternateEvenColor
      : global.theme.columnAlternateOddColor;
}
```

### Edit Mode Variable — Choose Correctly
| Group | Variable | Example Screens |
|-------|----------|----------------|
| A | `isEditMode` | bank, color, employee, kitchen, barcode, product, coupon |
| B | `isSaveAllow` | business_type, department, product_group, product_unit, master_* |

Pitfall: `screenEvent == edit` does not work — `switchToEdit()` doesn't set screenEvent.

---

## 5. List Item — MouseRegion + GestureDetector

| Action | Behavior |
|--------|----------|
| Single tap | Select (view mode) |
| Double tap | Enter edit mode |
| Hover enter | `_hoverIndex = index` |
| Hover exit | `_hoverIndex = -1` |

```dart
MouseRegion(
  onEnter: (_) => setState(() => _hoverIndex = index),
  onExit: (_) => setState(() => _hoverIndex = -1),
  child: GestureDetector(
    onTap: () { /* select + getData */ },
    onDoubleTap: () { if (!showCheckBox) switchToEdit(value); },
    child: Container(
      color: _getContainerColor(value.guidfixed, index),
      // ... row content
    ),
  ),
)
```

---

## 6. editScreen (Right Panel)

- Scaffold `backgroundColor: global.theme.cardColor` (not backgroundColor — separates form visually)
- AppBar color: `appBarColor` (view) / `toolBarEditModeColor` (edit)
- Master selector borders use `primaryColor` (not `appBarColor` — too dark in dark mode)

See [references/patterns.md](references/patterns.md) for editScreen, master selector, and ElevatedButton selector code.

---

## 7. Edit Font Scale (A- % A+)

Edit panel supports 50%-300% font scaling via `EditFontSizeControl`.

4-layer scaling approach:
| Layer | Purpose |
|-------|---------|
| `MediaQuery.textScaler` | Scale all text in child widgets |
| `Theme.inputDecorationTheme` | Scale contentPadding (prevents label overlap) |
| `IconTheme` | Scale icons with text |
| `SizedBox` spacing | Extra gap between form widgets at higher scales |

Failed approaches: `Transform.scale` (breaks scroll), `FittedBox` (breaks text wrap), `DefaultTextStyle.merge` alone (doesn't propagate to child widgets), `MediaQuery` alone (padding/spacing don't scale).

For transaction screens (PO/QT/SO), wrap TabBarView with MediaQuery + Theme (same pattern, skip SizedBox spacing).

---

## 8. Checklist

- [ ] State class has `global.ThemeRefreshMixin`
- [ ] `_getContainerColor()` uses `isEditMode` or `isSaveAllow` (not `screenEvent`)
- [ ] `_getContainerColor()` covers all 3 states: edit/selected/hover/alternating
- [ ] listScreen `backgroundColor: global.theme.backgroundColor`
- [ ] editScreen `backgroundColor: global.theme.cardColor`
- [ ] AppBar edit mode uses `toolBarEditModeColor`
- [ ] Search bar uses `searchBarColor` + `ListFontSizeControl` (no duplicate line spacing button)
- [ ] Font size uses `global.deviceConfig.listDataFontSize` (not local variable)
- [ ] Column header uses `columnHeaderColor` + `columnHeaderTextColor`
- [ ] SplitView gripColor uses `dividerBorderColor`
- [ ] No `Colors.*` hardcode (except shadow/overlay/color picker data)
- [ ] Master selector uses `primaryColor` not `appBarColor`
- [ ] ElevatedButton selector is flat (elevation: 0, border, cardColor bg)
- [ ] MouseRegion + GestureDetector complete (hover, tap, doubleTap)
- [ ] Phone fields have no `maxLength`/`inputFormatters`

## Pitfalls

| Problem | Cause | Fix |
|---------|-------|-----|
| Duplicate line spacing button | `ListFontSizeControl` already has it + separate `IconButton` | Remove standalone `IconButton(Icons.line_weight)` |
| Phone input error | `maxLength: 10` + digit-only formatters | Remove both — phones can have `-`, `+`, spaces |
| `_getContainerColor` not working | Method exists but not called — inline ternary used instead | Use `color: _getContainerColor(value.guidfixed, index)` |
| Row edit color not orange | Using `screenEvent == edit` | Use `isEditMode` or `isSaveAllow` |
| Theme change not reflected | Hardcoded `Color.fromARGB`/`Colors.*` | Replace with `global.theme.*` |
| Font size not syncing | Local `_listFontSize` variable | Use `global.deviceConfig.listDataFontSize` + `ListFontSizeControl` |
