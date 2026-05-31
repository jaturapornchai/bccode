---
name: flutter-expert
description: Auto-activate when writing, reviewing, or modifying Flutter/Dart code.
---

- **BLoC State**: Always use the BLoC pattern. No `setState` in production files.
- **Code Hygiene**: Use `const` widgets, use async/await (no `.then`), dispose all controllers.
- **Theme & I18n**: Use `global.theme.*` for styling (no hardcoded colors) and `global.language('key')`.
- **Buddhist Era**: Use `CustomDatePicker` instead of `showDatePicker`.
- **Checks**: Run `dart analyze` after changes. No nested FutureBuilders.
