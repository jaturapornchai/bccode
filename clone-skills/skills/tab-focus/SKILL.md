---
name: tab-focus
description: >
  Tab focus cycling for Flutter config/edit screens. Controls Tab/Shift+Tab to cycle
  only through text fields. Cursor goes to end of text. Cross-platform. Use when the
  user mentions Tab key behavior in edit panels, config forms, or transaction screens.
---

# Tab Focus Cycling

## When to Use

- Building a new edit screen with multiple form fields
- Tab key escapes from the form to other buttons or out of the screen
- Tab is slow (may be using Approach B with 100 pre-allocated FocusNodes)
- Migrating a legacy screen from `RawKeyboardListener` / Approach B to Approach A
- Need Enter to behave like Tab (next focus)

## Anti-Patterns

- Do NOT use `RawKeyboardListener` — deprecated, use `Focus` + `onKeyEvent` instead
- Do NOT pre-allocate 100 FocusNodes (Approach B) for new screens — slow, use Approach A
- Do NOT forget to dispose self-created FocusNodes — memory leak on long-running sessions
- Do NOT call `requestFocus()` inside focusNode listener — infinite loop
- Do NOT use `context.widget is EditableText` — use `findAncestorWidgetOfExactType<EditableText>()` instead
- Do NOT forget `skipTraversal: true` on buttons/checkboxes inside a FocusTraversalGroup

## Related Skills

- `/data-list` — SplitView editScreen that uses tab-focus pattern
- `/theming` — not directly related, but form structure is shared
- `/date-time-picker` — date/time fields in the same focus traversal group

## Architecture — 2 Approaches

### Approach A: FocusTraversalGroup + TextFieldTraversalPolicy (Recommended)
For screens with many or dynamic fields. Auto-detects TextField/TextFormField via EditableText ancestor. Skips buttons, checkboxes, radio, tab headers, and readOnly fields automatically.

### Approach B: fieldFocusNodes + Focus onKeyEvent (Legacy)
For small screens (2-5 fields) needing per-field readonly control. Requires pre-allocated FocusNodes and manual skipTraversal marking.

---

## Approach A: FocusTraversalGroup

Uses utilities from `lib/utils/focus_utils.dart`:
- `TextFieldTraversalPolicy` — filters to EditableText nodes only
- `focusFirstTextField(context)` — auto-focus first non-readOnly field
- `focusAndCursorToEnd(node)` — focus + cursor to end (not select all)

### Implementation

#### 1. Import
```dart
import 'package:smlaicloud/utils/focus_utils.dart';
```

#### 2. Wrap edit body
```dart
body: Focus(
  skipTraversal: true,
  onKeyEvent: (node, event) {
    if (event is KeyUpEvent) return KeyEventResult.ignored;
    if (event.logicalKey == LogicalKeyboardKey.f10) {
      if (event is KeyDownEvent) saveOrUpdateData();
      return KeyEventResult.handled;
    }
    if (event.logicalKey == LogicalKeyboardKey.enter) {
      if (event is KeyDownEvent) {
        FocusManager.instance.primaryFocus?.nextFocus();
      }
      return KeyEventResult.handled;
    }
    return KeyEventResult.ignored;
  },
  child: FocusTraversalGroup(
    policy: TextFieldTraversalPolicy(),
    child: /* ... form content ... */
  ),
),
```

Tab/Shift+Tab is handled by FocusTraversalGroup — no manual interception needed.

#### 3. Auto-focus on edit/add
```dart
void switchToEdit(Model value) {
  setState(() { /* ... */ });
  WidgetsBinding.instance.addPostFrameCallback((_) {
    focusFirstTextField(context);
  });
}
```

#### 4. skipTraversal for widgets inside the group
Widgets outside the FocusTraversalGroup are skipped automatically. For widgets inside:
```dart
ElevatedButton(focusNode: FocusNode(skipTraversal: true), ...)
Checkbox(focusNode: FocusNode(skipTraversal: true), ...)
```

### How TextFieldTraversalPolicy Works
- Filters FocusNodes that have EditableText as ancestor
- Skips `editableText.readOnly == true` fields
- Overrides next/previous to wrap around (last->first, first->last)
- Uses `focusAndCursorToEnd()` for cursor placement

### Cursor to End
`focusAndCursorToEnd()` calls `requestFocus()`, then in `addPostFrameCallback` looks up EditableText ancestor to get controller and sets `TextSelection.collapsed(offset: text.length)`.

The EditableText lookup must happen inside `addPostFrameCallback` — `requestFocus()` can trigger rebuild, making the old controller stale.

---

## Approach B: fieldFocusNodes (Legacy)

### State Variables
```dart
int focusNodeMax = 0;
List<global.FieldFocusModel> fieldFocusNodes = [];
int focusNodeIndex = 0;
```

### initState — Pre-allocate FocusNodes
```dart
for (int i = 0; i < 100; i++) {
  fieldFocusNodes.add(global.FieldFocusModel(focusNode: FocusNode()));
  fieldFocusNodes[i].focusNode.addListener(() {
    if (fieldFocusNodes[i].focusNode.hasFocus) focusNodeIndex = i;
  });
}
```

Do NOT call `requestFocus()` inside the listener — causes infinite loop.

### findFocusNext/findFocusPrev
```dart
void findFocusNext(int index) {
  focusNodeIndex = index;
  do {
    focusNodeIndex++;
    if (focusNodeIndex > focusNodeMax) focusNodeIndex = 0;
  } while (fieldFocusNodes[focusNodeIndex].isReadOnly);
  WidgetsBinding.instance.addPostFrameCallback((_) {
    focusAndCursorToEnd(fieldFocusNodes[focusNodeIndex].focusNode);
  });
}
```

### Focus Widget Wrapper
```dart
body: Focus(
  skipTraversal: true,
  onKeyEvent: (node, event) {
    if (event is KeyUpEvent) return KeyEventResult.ignored;
    if (event.logicalKey == LogicalKeyboardKey.f10) {
      if (event is KeyDownEvent) saveOrUpdateData();
      return KeyEventResult.handled;
    }
    if (event.logicalKey == LogicalKeyboardKey.tab ||
        event.logicalKey == LogicalKeyboardKey.enter) {
      if (event is KeyDownEvent) {
        HardwareKeyboard.instance.isShiftPressed
            ? findFocusPrev(focusNodeIndex)
            : findFocusNext(focusNodeIndex);
      }
      return KeyEventResult.handled;
    }
    return KeyEventResult.ignored;
  },
  child: /* ... */
),
```

---

## Pitfalls

| Problem | Cause | Fix |
|---------|-------|-----|
| Tab very slow | 100 pre-allocated FocusNodes + Timer | Use Approach A |
| Tab goes to buttons/headers | Default policy includes all focusables | Use `TextFieldTraversalPolicy` |
| Tab doesn't work at all | Checking `context.widget is EditableText` | Use `findAncestorWidgetOfExactType<EditableText>()` |
| Focus selects all text | Default `requestFocus()` behavior | Use `focusAndCursorToEnd()` |
| Tab escapes to other screens | Missing wrapper or returning `ignored` on Tab | Approach A: FocusTraversalGroup / Approach B: return `handled` |
| Tab stuck on key hold | Only handling KeyDownEvent | Return `handled` for all non-KeyUp events |
| Infinite focus loop | `requestFocus()` inside focusNode listener | Remove from listener |
| Cursor not moving to end | EditableText looked up before callback | Lookup inside `addPostFrameCallback` |
| Focus lands on readOnly field | Not checking `editableText.readOnly` | Fixed in focus_utils.dart |

---

## Migration: Legacy to Approach A

### What to Change (13 points)
1. Add `import 'package:smlaicloud/utils/focus_utils.dart'`
2. Replace `RawKeyboardListener` with `Focus + FocusTraversalGroup(policy: TextFieldTraversalPolicy())`
3. `switchToEdit`: replace `findFocusNext(0)` with `focusFirstTextField(context)`
4. Add button: replace `fieldFocusNodes[0].focusNode.requestFocus()` with `focusFirstTextField(context)`
5. `clearEditData()`: remove `fieldFocusNodes[focusNodeIndex].focusNode.requestFocus()`
6. `findFocusNext/Prev`: `requestFocus()` -> `focusAndCursorToEnd()`
7. initState listener: remove `requestFocus()` (keep index tracking)
8. Timer refresh: `requestFocus()` -> `focusAndCursorToEnd()`
9. BLoC GetSuccess: nested `findFocusNext(0)` -> `focusFirstTextField(context)` in callback
10. `onEditingComplete`: -> `focusAndCursorToEnd(someFocusNode)`
11. Dropdown `onChanged`: -> `focusAndCursorToEnd(someFocusNode)`
12. Helper functions: `setFocusNode(focus)` -> `focusAndCursorToEnd(focus)`
13. Remove manual `_focusAndMoveCursorToEnd` implementations

### What NOT to Change
- `fieldFocusNodes` infrastructure — TextFormFields still use focusNode from it
- `initState` / `dispose`
- `fieldFocusNodes[i].isReadOnly` logic
- `searchFocusNode.requestFocus()` — correct for returning focus to search on save/cancel

### Migrated Screens (Reference)
All 40+ config screens fully migrated. Also: promotion_screen, doc_format_screen, product_barcode_screen, currency_form_screen, pattern_screen.
