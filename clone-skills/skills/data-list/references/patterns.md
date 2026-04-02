# Data List Code Patterns

## Search Bar

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

## Column Header

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
    ],
  ),
)
```

## List Item

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
        padding: const EdgeInsets.only(left: 10, right: 10, top: 5, bottom: 5),
        child: Row(children: [/* data columns */]),
      ),
    ),
  );
}
```

## editScreen

```dart
Widget editScreen({bool mobileScreen = false}) {
  return Scaffold(
    backgroundColor: global.theme.cardColor,
    appBar: AppBar(
      backgroundColor: isEditMode
          ? global.theme.toolBarEditModeColor
          : global.theme.appBarColor,
      title: Text(headerEdit + global.language("my_title")),
      actions: [
        if (selectGuid.isNotEmpty)
          IconButton(
            icon: Icon(Icons.delete, color: global.theme.onPrimaryColor),
            onPressed: () => deleteData(),
          ),
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
            if (isSaveAllow)
              ElevatedButton.icon(
                style: ElevatedButton.styleFrom(backgroundColor: global.theme.buttonColor),
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

### Background Colors
| Screen | Property | Reason |
|--------|----------|--------|
| listScreen | `backgroundColor` | Matches alternating row grey tones |
| editScreen | `cardColor` | White/dark card — visually separated as form |

## Master Data Selector

```dart
Widget _buildMasterDataSelector({
  required String label, required String code,
  required String displayName, required VoidCallback onTap,
  required VoidCallback onClear, bool isSelected = false,
}) {
  return Container(
    height: 36,
    decoration: BoxDecoration(
      border: Border.all(
        color: isSelected ? global.theme.primaryColor : global.theme.textSecondaryColor,
      ),
      borderRadius: BorderRadius.circular(4),
      color: isSelected ? global.theme.primaryColor.withValues(alpha: 0.08) : null,
    ),
    child: Row(children: [
      Text('$label: ', style: TextStyle(color: global.theme.textSecondaryColor)),
      Expanded(child: Text(
        isSelected ? '$code ~ $displayName' : '-',
        style: TextStyle(
          color: isSelected ? global.theme.textColor : global.theme.textSecondaryColor,
        ),
      )),
      if (isSelected && isEditMode)
        InkWell(
          onTap: onClear,
          child: Icon(Icons.clear, color: global.theme.negativeHighlightTextColor, size: 16),
        ),
      Icon(Icons.search,
        color: isEditMode ? global.theme.primaryColor : global.theme.textSecondaryColor,
        size: 16,
      ),
    ]),
  );
}
```

Pitfall: Use `primaryColor` for accent/border, not `appBarColor` (#0D1B2A in dark = invisible).

## ElevatedButton Selector (Large Master Selector)

Flat style with border — no heavy colored background:

```dart
ElevatedButton(
  style: ButtonStyle(
    backgroundColor: WidgetStateProperty.all<Color>(
      isDataSelected
          ? global.theme.primaryColor.withValues(alpha: 0.08)
          : global.theme.cardColor,
    ),
    foregroundColor: WidgetStateProperty.all<Color>(global.theme.textColor),
    elevation: WidgetStateProperty.all(0),
    shape: WidgetStateProperty.all(RoundedRectangleBorder(
      borderRadius: BorderRadius.circular(8),
      side: BorderSide(
        color: isDataSelected ? global.theme.primaryColor : global.theme.dividerBorderColor,
      ),
    )),
  ),
  child: Row(children: [
    Icon(Icons.category, color: global.theme.primaryColor),
    const SizedBox(width: 8),
    Expanded(child: Text(...)),
    if (isDataSelected)
      IconButton(icon: Icon(Icons.clear, color: global.theme.negativeHighlightTextColor), onPressed: onClear),
    Icon(Icons.search, color: global.theme.textSecondaryColor),
  ]),
)
```

## Edit Font Scale (4-Layer Wrapper)

```dart
body: LoaderOverlay(
  child: Builder(
    builder: (context) {
      final scale = global.editFontScaleFactor;
      List<Widget> scaledFormWidgets = formWidgets;
      if (scale > 1.0) {
        final extraGap = 8.0 * (scale - 1);
        scaledFormWidgets = [];
        for (int i = 0; i < formWidgets.length; i++) {
          scaledFormWidgets.add(formWidgets[i]);
          if (i < formWidgets.length - 1) scaledFormWidgets.add(SizedBox(height: extraGap));
        }
      }
      return MediaQuery(
        data: MediaQuery.of(context).copyWith(textScaler: TextScaler.linear(scale)),
        child: Theme(
          data: Theme.of(context).copyWith(
            inputDecorationTheme: Theme.of(context).inputDecorationTheme.copyWith(
              contentPadding: EdgeInsets.fromLTRB(12 * scale, 20 * scale, 12 * scale, 12 * scale),
            ),
          ),
          child: IconTheme(
            data: IconTheme.of(context).copyWith(size: 24.0 * scale),
            child: SingleChildScrollView(
              child: Form(key: _formKey, child: Column(children: scaledFormWidgets)),
            ),
          ),
        ),
      );
    },
  ),
),
```

### Transaction Edit Screens (PO/QT/SO)
```dart
// AppBar
actions: [
  EditFontSizeControl(onChanged: () => setState(() {})),
  ...POAppBarActions.build(...),
],

// Body — wrap TabBarView
child: Builder(
  builder: (context) {
    final scaleFactor = global.editFontScaleFactor;
    return MediaQuery(
      data: MediaQuery.of(context).copyWith(textScaler: TextScaler.linear(scaleFactor)),
      child: Theme(
        data: Theme.of(context).copyWith(
          inputDecorationTheme: Theme.of(context).inputDecorationTheme.copyWith(
            contentPadding: EdgeInsets.symmetric(horizontal: 12 * scaleFactor, vertical: 10 * scaleFactor),
          ),
          iconTheme: IconThemeData(size: 24 * scaleFactor),
        ),
        child: TabBarView(...),
      ),
    );
  },
),
```
