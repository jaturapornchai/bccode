---
name: flutter-expert
description: Auto-activate when writing, reviewing, or modifying Flutter/Dart code
---
When working with Flutter/Dart:

## Mandatory Steps
1. LSP goToDefinition before modifying an unfamiliar Widget
2. Check widget tree — do not exceed 3 levels deep without extracting
3. Run dart analyze after every change

## Required Patterns
- State: always use BLoC pattern — no setState in production
- Widget: use const constructor wherever possible
- Async: async/await — do not use .then() chains
- Dispose: every controller and subscription must call dispose()
- Image: use cached_network_image for network images
- Theme: use global.theme.* — never hardcode colors

## Forbidden Anti-patterns
- setState inside StatelessWidget
- Nested FutureBuilders
- Using BuildContext across async gap without checking mounted
- Non-responsive layouts
- showDatePicker (use CustomDatePicker to support Buddhist Era)
