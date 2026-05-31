---
name: flutter-screen
description: Create Flutter screens following BC Account app patterns.
---

- **BLoC State**: Use BLoC (no `setState` in production). Use StatefulWidget as wrapper.
- **Widgets**: Support responsive layouts (phone/tablet). Keep tree depth under 3 (extract subwidgets).
- **Date Picker**: Use `CustomDatePicker` to support the Buddhist Era.
- **Language**: Use `global.language('key')`. Include loading, error, and empty states.
