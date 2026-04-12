---
name: date-time-picker
description: >
  Date and time picker components for Flutter ERP. CustomDatePicker with Thai Buddhist
  calendar (B.E.), CustomTimePicker with 24-hour quick select. Use instead of Flutter
  built-ins whenever the user mentions date/time input in a form screen.
---

# Date & Time Picker Components

## When to Use

- Adding a date field in a document header (PO, SO, QT, PR, etc.)
- Adding a time field alongside a date field in a transaction screen
- Screens with date range (firstDate/lastDate) for filtering
- Reviewing whether a screen uses Flutter built-in `showDatePicker()` (must fix)
- Checking that date/time merge correctly when selected separately

## Anti-Patterns

- Do NOT use `showDatePicker()` — does not support Thai Buddhist calendar (B.E.)
- Do NOT use `showTimePicker()` — no quick select and not 24hr format
- Do NOT hardcode `useBuddhistCalendar: true/false` — use `null` to auto-detect from branch settings
- Do NOT hardcode `languageCode` — use `null` to auto-detect
- Do NOT forget to preserve existing time when selecting date (and existing date when selecting time)
- Do NOT forget `.toLocal()` before display — otherwise timezone will be wrong

## Related Skills

- `/data-list` — SplitView form pattern where date picker is used in edit panel
- `/tab-focus` — Tab focus cycling including date/time fields
- `/theming` — form input decoration colors

## Components

| Component | File | Purpose |
|-----------|------|---------|
| `CustomDatePicker` | `lib/utils/date_picker.dart` | Thai Buddhist calendar (B.E.) date picker with overlay popup |
| `CustomTimePicker` | `lib/utils/time_picker.dart` | 24-hour time picker with quick select buttons |
| `DatePickerWidget` | `lib/widgets/date_picker_widget.dart` | Wrapper with label + required indicator |

---

## CustomDatePicker

### Parameters
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `onDateSelected` | `Function(DateTime?)` | required | Callback on date selection |
| `labelText` | `String` | required | Field label |
| `decoration` | `InputDecoration` | required | Required param but widget builds its own decoration |
| `useBuddhistCalendar` | `bool?` | `null` (auto) | `null` = auto-detect from `global.getYearType()` |
| `languageCode` | `String?` | `null` (auto) | `null` = auto-detect from `global.getBranchLanguage()` |
| `useIconSelectDate` | `bool` | `false` | Show calendar icon button instead of text field |
| `initialDate` | `DateTime?` | `null` | Initial date |
| `firstDate` / `lastDate` | `DateTime?` | `null` | Date range limits |

### Features
- Auto-detects language and year type from branch settings
- Thai weekday/month names, Buddhist year (+543)
- Gradient header + glassmorphism popup

### Usage
```dart
CustomDatePicker(
  labelText: global.language("doc_date"),
  useIconSelectDate: true,
  initialDate: global.safeParseDatetime(screenData.docdatetime).toLocal(),
  onDateSelected: (date) {
    if (date != null) {
      setState(() {
        final existing = global.safeParseDatetime(screenData.docdatetime).toLocal();
        screenData.docdatetime = DateTime(
          date.year, date.month, date.day,
          existing.hour, existing.minute, existing.second,
        ).toIso8601String();
      });
    }
  },
  decoration: _buildInputDecoration(labelText: global.language("doc_date")),
),
```

---

## CustomTimePicker

### Parameters
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `onTimeSelected` | `Function(TimeOfDay?)` | required | Callback on time selection |
| `labelText` | `String` | required | Field label |
| `controller` | `TextEditingController?` | `null` | External controller for display |
| `useIconSelectTime` | `bool` | `false` | Show clock icon button instead of text field |
| `initialTime` | `TimeOfDay?` | `null` | Initial time |

### Features
- Hour/Minute spinners with arrow buttons
- Quick select: 00:00, 08:00, 12:00, 18:00
- 24-hour format

### Usage
```dart
CustomTimePicker(
  labelText: global.language("doc_time"),
  controller: docTimeController,
  useIconSelectTime: true,
  initialTime: TimeOfDay(hour: dt.hour, minute: dt.minute),
  onTimeSelected: (time) {
    if (time != null) {
      setState(() {
        final existing = global.safeParseDatetime(screenData.docdatetime).toLocal();
        screenData.docdatetime = DateTime(
          existing.year, existing.month, existing.day,
          time.hour, time.minute,
        ).toIso8601String();
      });
    }
  },
),
```

---

## Date+Time Pattern in Document Headers

When selecting date, preserve existing time. When selecting time, preserve existing date. Store as ISO8601 string.

---

## Rules

| Rule | Reason |
|------|--------|
| Use `CustomDatePicker`, not `showDatePicker()` | Built-in lacks Thai Buddhist calendar support |
| Use `CustomTimePicker`, not `showTimePicker()` | Built-in lacks quick select + 24hr format |
| Keep `useBuddhistCalendar: null` | Let auto-detect work — don't hardcode |
| Keep `languageCode: null` | Let auto-detect work |
| `decoration` param is required but unused | Widget builds its own InputDecoration internally |
| Store datetime as ISO8601 | Use `.toIso8601String()` in model |
| Preserve date/time on partial selection | Merge new date with old time, or vice versa |

## Pitfalls

| Problem | Cause | Fix |
|---------|-------|-----|
| Year shows CE instead of BE | Hardcoded `useBuddhistCalendar: false` | Use `null` (auto-detect) |
| Time lost after date selection | Not preserving existing time in callback | Merge new date with existing time |
| Date lost after time selection | Not preserving existing date in callback | Merge new time with existing date |
| Wrong timezone display | Missing `.toLocal()` before display | Use `global.safeParseDatetime(...).toLocal()` |
