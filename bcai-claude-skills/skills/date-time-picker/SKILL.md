---
name: date-time-picker
description: >
  Date and time picker components for bcdev ERP Flutter app.
  CustomDatePicker supports Thai Buddhist calendar (พ.ศ.) with auto-detection
  from branch settings. CustomTimePicker provides 24-hour time selection with
  quick select buttons. Always use these instead of Flutter's built-in
  showDatePicker/showTimePicker. Use when: (1) adding date/time fields to
  forms, (2) user says "date picker", "time picker", "เลือกวันที่", "เลือกเวลา",
  (3) Buddhist calendar / ปี พ.ศ. related work, (4) creating document headers
  with date/time fields.
---

# Date & Time Picker Components

## Components Overview

| Component | File | Purpose |
|-----------|------|---------|
| `CustomDatePicker` | `lib/utils/date_picker.dart` | Thai Buddhist calendar date picker with overlay popup |
| `CustomTimePicker` | `lib/utils/time_picker.dart` | 24-hour time picker with overlay popup + quick select |
| `DatePickerWidget` | `lib/widgets/date_picker_widget.dart` | Wrapper widget around CustomDatePicker with label + required indicator |

---

## CustomDatePicker

Thai Buddhist calendar (พ.ศ.) date picker with gradient header, glassmorphism overlay popup.

### Parameters

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `onDateSelected` | `Function(DateTime?)` | required | Callback เมื่อเลือกวันที่ |
| `labelText` | `String` | required | Label แสดงบน field |
| `decoration` | `InputDecoration` | required | **Required แต่ไม่ได้ใช้จริง** — widget สร้าง InputDecoration เอง |
| `useBuddhistCalendar` | `bool?` | `null` (auto) | `null` = auto-detect จาก `global.getYearType()` |
| `languageCode` | `String?` | `null` (auto) | `null` = auto-detect จาก `global.getBranchLanguage()` |
| `useIconSelectDate` | `bool` | `false` | แสดง calendar icon button แทน text field |
| `initialDate` | `DateTime?` | `null` | วันที่เริ่มต้น |
| `firstDate` | `DateTime?` | `null` | วันที่แรกสุดที่เลือกได้ |
| `lastDate` | `DateTime?` | `null` | วันที่ท้ายสุดที่เลือกได้ |

### Features

- Auto-detects language from `global.getBranchLanguage()` — แสดงชื่อวัน/เดือนเป็นภาษาไทย
- Auto-detects year type from `global.getYearType()` — แสดงปี พ.ศ. อัตโนมัติ
- Gradient header design + glassmorphism popup
- Shows weekday names in Thai, Thai month names, Buddhist year (+543)

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

24-hour time picker with overlay popup, hour/minute spinners, and quick select buttons.

### Parameters

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `onTimeSelected` | `Function(TimeOfDay?)` | required | Callback เมื่อเลือกเวลา |
| `labelText` | `String` | required | Label แสดงบน field |
| `controller` | `TextEditingController?` | `null` | External controller สำหรับแสดงเวลาใน text field |
| `useIconSelectTime` | `bool` | `false` | แสดง clock icon button แทน text field |
| `initialTime` | `TimeOfDay?` | `null` | เวลาเริ่มต้น |
| `languageCode` | `String?` | `null` (auto) | `null` = auto-detect จาก `global.getBranchLanguage()` |

### Features

- Hour/Minute spinners with up/down arrow buttons
- Quick select buttons:
  - เที่ยงคืน (00:00)
  - เช้า (08:00)
  - เที่ยง (12:00)
  - เย็น (18:00)
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

## DatePickerWidget

Wrapper widget around CustomDatePicker — เพิ่ม label text + required indicator (`*`).

ใช้เมื่อต้องการ label + required mark ครอบ CustomDatePicker โดยไม่ต้องจัด layout เอง.

---

## Date/Time Pattern ใน Document Headers

เมื่อ form มีทั้งวันที่และเวลา ให้ใช้คู่กัน:

```dart
// Date field
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

// Time field
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

**Pattern สำคัญ**: เมื่อเลือกวันที่ → เก็บเวลาเดิมไว้ / เมื่อเลือกเวลา → เก็บวันที่เดิมไว้

---

## Rules

| Rule | Description |
|------|-------------|
| ห้ามใช้ `showDatePicker()` | ใช้ `CustomDatePicker` เสมอ — รองรับ Thai Buddhist calendar |
| ห้ามใช้ `showTimePicker()` | ใช้ `CustomTimePicker` เสมอ — มี quick select + 24hr format |
| `useBuddhistCalendar` = `null` | ปล่อย auto-detect — อย่า hardcode `true`/`false` ยกเว้นมีเหตุผลชัดเจน |
| `languageCode` = `null` | ปล่อย auto-detect — อย่า hardcode ยกเว้นมีเหตุผลชัดเจน |
| `decoration` required แต่ไม่ใช้ | CustomDatePicker ต้องใส่ `decoration` param แต่ widget สร้าง decoration เอง |
| เก็บวันเวลาเป็น ISO8601 | ใช้ `.toIso8601String()` เก็บใน model, แสดงผลด้วย `global.formatDateTime()` |
| Date+Time คู่กัน | เมื่อเลือกวันที่ → preserve เวลาเดิม / เลือกเวลา → preserve วันที่เดิม |

---

## Common Pitfalls

| Problem | Cause | Fix |
|---------|-------|-----|
| ปีแสดงเป็น ค.ศ. ทั้งที่ตั้ง พ.ศ. | Hardcode `useBuddhistCalendar: false` | ใช้ `null` (auto-detect) |
| เลือกวันที่แล้วเวลาหาย | ไม่ได้ preserve เวลาเดิมใน onDateSelected | ต้อง merge date ใหม่กับ time เดิมจาก existing datetime |
| เลือกเวลาแล้ววันที่หาย | ไม่ได้ preserve วันที่เดิมใน onTimeSelected | ต้อง merge time ใหม่กับ date เดิมจาก existing datetime |
| วันที่แสดงผิด timezone | ไม่ได้ `.toLocal()` ก่อนแสดง | ใช้ `global.safeParseDatetime(...).toLocal()` สำหรับ initialDate |
| ใช้ Flutter showDatePicker | ไม่รองรับปี พ.ศ. + ไม่มี Thai locale | ใช้ `CustomDatePicker` เสมอ |
