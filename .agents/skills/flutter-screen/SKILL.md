---
name: flutter-screen
description: >
  Create Flutter screens following BC Account app patterns.
  Use when: creating a new screen, screen, page, or widget.
  Trigger: "สร้างหน้า", "new screen", "create page", "เพิ่มหน้าจอ", "add screen"
---

When creating a new Flutter screen:

1. Use StatefulWidget with BLoC pattern
2. Use Theme from global.theme.* (never hardcode colors)
3. Support responsive layout (phone + tablet)
4. Include loading state, error state, and empty state
5. Use global.language('key') for all text strings
6. Always write a widget test alongside the screen
7. No setState in production (use BLoC)
8. Use CustomDatePicker instead of showDatePicker (supports Buddhist Era)
