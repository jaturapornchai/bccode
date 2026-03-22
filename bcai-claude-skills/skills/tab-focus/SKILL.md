---
name: tab-focus
description: >
  Tab focus cycling skill for Flutter config/edit screens in bcdev ERP.
  Controls Tab/Shift+Tab to cycle ONLY through text input fields, preventing
  focus from escaping to buttons, checkboxes, radio buttons, tab headers, menus,
  or other screens. Cursor goes to end of text (not select all).
  Cross-platform: web, Windows, macOS. Use when: (1) adding Tab focus
  control to a config screen, (2) fixing Tab escaping to other screens/menus,
  (3) user says "tab วน", "tab หลุด", "focus cycling", "tab ค้าง",
  (4) creating new edit/config screens that need keyboard navigation.
---

# Tab Focus Cycling

## Architecture — 2 Approaches

### Approach A: FocusTraversalGroup + TextFieldTraversalPolicy (แนะนำ)

ใช้สำหรับจอที่มี fields เยอะ หรือ fields แบบ dynamic (sub-widgets, language lists)
- **เร็ว** — ไม่ต้อง pre-allocate FocusNodes, ไม่มี timer
- **อัตโนมัติ** — ตรวจ TextField/TextFormField จาก EditableText ancestor
- **ข้าม tab headers, buttons, checkboxes, radio ทั้งหมด** โดยไม่ต้อง mark skipTraversal ทีละตัว
- **ข้าม readOnly fields อัตโนมัติ** — เช่น field รหัสที่ห้ามแก้ตอน edit (ตรวจ `editableText.readOnly`)

### Approach B: fieldFocusNodes + Focus onKeyEvent (Legacy)

ใช้สำหรับจอเล็กที่มี fields น้อย (2-5 fields) และต้อง control readonly/skip เฉพาะ field
- ต้อง pre-allocate FocusNodes + assign index ทีละ field
- ต้อง mark `skipTraversal: true` ทุก non-text widget เอง

---

## Approach A: FocusTraversalGroup (แนะนำ)

ใช้ utilities จาก `lib/utils/focus_utils.dart`:
- `TextFieldTraversalPolicy` — custom policy ที่ filter เฉพาะ EditableText
- `focusFirstTextField(context)` — auto-focus field แรกที่ไม่ใช่ readOnly
- `focusAndCursorToEnd(node)` — focus + cursor ไปท้ายข้อความ (public function, ใช้ได้ทั้ง Approach A และ B)

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
    // F10 = save
    if (event.logicalKey == LogicalKeyboardKey.f10) {
      if (event is KeyDownEvent) saveOrUpdateData();
      return KeyEventResult.handled;
    }
    // Enter = next field
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

**สำคัญ**: ไม่ต้อง intercept Tab — `FocusTraversalGroup` + `TextFieldTraversalPolicy` จัดการ Tab/Shift+Tab เอง

#### 3. Auto-focus on enter edit/add

```dart
void switchToEdit(Model value) {
  setState(() { /* ... */ });
  WidgetsBinding.instance.addPostFrameCallback((_) {
    focusFirstTextField(context);
  });
}

// ปุ่มเพิ่ม:
onPressed: () {
  setState(() { /* ... clear/setup ... */ });
  WidgetsBinding.instance.addPostFrameCallback((_) {
    focusFirstTextField(context);
  });
}
```

#### 4. ยังต้อง skipTraversal บาง widget (เฉพาะที่อยู่ใน FocusTraversalGroup)

ถ้า widget อยู่นอก FocusTraversalGroup ไม่ต้อง mark — policy จะข้ามให้อัตโนมัติ
แต่ถ้าอยู่ใน group เดียวกัน ให้ mark เผื่อ:

```dart
ElevatedButton(focusNode: FocusNode(skipTraversal: true), ...)
Checkbox(focusNode: FocusNode(skipTraversal: true), ...)
Radio(focusNode: FocusNode(skipTraversal: true), ...)
Switch(focusNode: FocusNode(skipTraversal: true), ...)
IconButton(focusNode: FocusNode(skipTraversal: true), ...)
```

### How TextFieldTraversalPolicy works

```dart
class TextFieldTraversalPolicy extends ReadingOrderTraversalPolicy {
  // filter เฉพาะ FocusNode ที่มี EditableText เป็น ancestor
  // → TextField/TextFormField ทุกตัวมี EditableText ข้างใน
  // → FocusNode ถูก attach ที่ Focus widget ใน EditableText
  // → findAncestorWidgetOfExactType<EditableText>() != null = text field
  // → ข้าม editableText.readOnly == true (เช่น field รหัส ที่ห้ามแก้)

  // override next/previous ให้วนลูป (last→first, first→last)
  // ใช้ focusAndCursorToEnd() เพื่อ cursor ไปท้ายข้อความ (ไม่ select all)
}
```

### Cursor to End (ไม่ select all)

`focusAndCursorToEnd()` ใน focus_utils.dart:
1. `requestFocus()` บน FocusNode
2. `addPostFrameCallback` → หา `EditableText` ancestor → ได้ `controller`
3. `TextSelection.collapsed(offset: text.length)`

**สำคัญ**:
- ต้อง lookup `EditableText` **ใน** `addPostFrameCallback` — ไม่ใช่ก่อน
- เพราะ `requestFocus()` อาจ trigger rebuild → widget instance เปลี่ยน → controller ตัวเก่า stale
- จอที่สร้าง `TextEditingController(text: ...)` inline ใน build จะโดนปัญหานี้แน่นอน
- ถ้า lookup ก่อน callback → set selection บน controller เก่าที่ไม่ได้ใช้แล้ว → cursor ไม่ขยับ

### Checklist — Approach A

1. `import 'package:smlaicloud/utils/focus_utils.dart';`
2. Wrap body: `Focus(onKeyEvent: F10+Enter) → FocusTraversalGroup(policy: TextFieldTraversalPolicy())`
3. Auto-focus: `addPostFrameCallback → focusFirstTextField(context)`
4. Test: Tab cycles through text fields only, Shift+Tab backward, wraps around
5. Test: cursor goes to end (not select all)
6. Test: readOnly fields (เช่น รหัส) ถูกข้ามอัตโนมัติ — ทั้ง Tab cycling และ auto-focus
7. Test on: web (Chrome), Windows, macOS

---

## Approach B: fieldFocusNodes (Legacy — จอเล็ก)

ใช้สำหรับจอที่มี fields น้อย + ต้อง control readonly per field

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
    if (fieldFocusNodes[i].focusNode.hasFocus) {
      focusNodeIndex = i;
    }
  });
}
```

**IMPORTANT**: Do NOT call `requestFocus()` inside the listener — causes infinite loop.

### findFocusNext / findFocusPrev — ใช้ focusAndCursorToEnd

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

**สำคัญ**: ใช้ `focusAndCursorToEnd()` แทน `requestFocus()` เพื่อให้ cursor ไปท้ายข้อความ ไม่ select all

### Focus Widget Wrapping body

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
        if (HardwareKeyboard.instance.isShiftPressed) {
          findFocusPrev(focusNodeIndex);
        } else {
          findFocusNext(focusNodeIndex);
        }
      }
      return KeyEventResult.handled;
    }
    return KeyEventResult.ignored;
  },
  child: /* ... */
),
```

### skipTraversal on ALL non-text widgets (ต้อง mark ทุกตัวเอง)

```dart
ElevatedButton(focusNode: FocusNode(skipTraversal: true), ...)
IconButton(focusNode: FocusNode(skipTraversal: true), ...)
Checkbox(focusNode: FocusNode(skipTraversal: true), ...)
Radio(focusNode: FocusNode(skipTraversal: true), ...)
Switch(focusNode: FocusNode(skipTraversal: true), ...)
```

---

## Common Pitfalls

| Problem | Cause | Fix |
|---------|-------|-----|
| Tab ช้ามาก / ไม่ตอบสนอง | Pre-allocate 100 FocusNodes + Timer polling | ใช้ Approach A: `FocusTraversalGroup` + `TextFieldTraversalPolicy` |
| Tab ไป tab headers/buttons | Default policy วน focusable ทั้งหมด | ใช้ `TextFieldTraversalPolicy` — filter เฉพาะ EditableText |
| Tab ไม่ทำงานเลย | ตรวจ `context.widget is EditableText` (ผิด) | ใช้ `ctx.findAncestorWidgetOfExactType<EditableText>()` — EditableText เป็น ancestor ไม่ใช่ context widget |
| Focus selects all text | `requestFocus()` default behavior | ใช้ `focusAndCursorToEnd()` — set `TextSelection.collapsed` ใน `addPostFrameCallback` |
| Tab escapes to other screens | Missing wrapper or `KeyEventResult.ignored` on Tab | Approach A: `FocusTraversalGroup` contain traversal / Approach B: return `handled` |
| Tab stuck/spinning on hold | Only handling `KeyDownEvent` | Return `handled` for ALL non-KeyUp events |
| Infinite focus loop | `requestFocus()` inside focusNode listener | Remove `requestFocus()` from listener |
| Dynamic fields index ผิด | Loop variable `i` captured by closure แต่ offset ไม่ถูก | Approach A ไม่มีปัญหานี้ / Approach B: ใช้ `final fieldIndex = i + offset;` |
| Cursor ไม่ไปท้าย (inline controller) | `focusAndCursorToEnd` จับ EditableText ก่อน callback → rebuild สร้าง controller ใหม่ → set selection บน controller เก่า | ต้อง lookup EditableText **ใน** `addPostFrameCallback` ไม่ใช่ก่อน (แก้แล้วใน focus_utils.dart) |
| Focus ไปที่ read-only field (รหัส) | ไม่ได้ตรวจ `editableText.readOnly` — field รหัสที่ห้ามแก้ตอน edit ยังถูก focus | แก้แล้วใน focus_utils.dart — `TextFieldTraversalPolicy` + `focusFirstTextField` ข้าม `readOnly: true` fields อัตโนมัติ |
| แก้ `searchFocusNode.requestFocus()` ผิด | `searchFocusNode` ใช้ focus กลับช่องค้นหาเมื่อ save/cancel — ไม่ใช่ edit form | **ไม่ต้องแก้** — ถูกต้องแล้ว ไม่ต้องเปลี่ยนเป็น `focusAndCursorToEnd` |

---

## Batch Migration Guide — แก้ทุกจอ config

### Pattern สำหรับ migrate จาก Legacy → Approach A

จอ config ส่วนใหญ่มี pattern เหมือนกัน สามารถ migrate เป็น batch ได้:

#### สิ่งที่ต้องแก้ (13 จุด)

1. **เพิ่ม import** — `import 'package:smlaicloud/utils/focus_utils.dart';`

2. **แก้ edit body** — เปลี่ยน `RawKeyboardListener` → `Focus + FocusTraversalGroup`:
   ```dart
   // เดิม:
   body: RawKeyboardListener(
     focusNode: FocusNode(skipTraversal: true),
     onKey: (event) { ... if (event.logicalKey == LogicalKeyboardKey.f10) ... },
     child: Builder(/* ... */),
   ),

   // ใหม่:
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
       child: Builder(/* ... เดิม ... */),
     ),
   ),
   ```
   **หมายเหตุ**: เพิ่ม 1 nesting level → ต้องเพิ่ม `)` ปิด FocusTraversalGroup

3. **แก้ switchToEdit** — เปลี่ยน `findFocusNext(0)` / `fieldFocusNodes[0].focusNode.requestFocus()` → `focusFirstTextField(context)`:
   ```dart
   // ใน addPostFrameCallback:
   focusFirstTextField(context);
   ```

4. **แก้ปุ่มเพิ่ม (Icons.add)** — เปลี่ยน `fieldFocusNodes[0].focusNode.requestFocus()` → `focusFirstTextField(context)`

5. **แก้ clearEditData()** — ลบ `fieldFocusNodes[focusNodeIndex].focusNode.requestFocus()` ออก

#### สิ่งที่ไม่ต้องแก้

- **fieldFocusNodes infrastructure** — เก็บไว้เพราะ TextFormField ยังใช้ focusNode จาก fieldFocusNodes
- **initState / dispose** — ไม่ต้องเปลี่ยน
- **fieldFocusNodes[i].isReadOnly** logic — ไม่ต้องเปลี่ยน

6. **แก้ findFocusNext/findFocusPrev** — `requestFocus()` + TextSelection → `focusAndCursorToEnd()`

7. **แก้ initState listener** — ลบ `requestFocus()` (เหลือแค่ track focusNodeIndex)

8. **แก้ Timer refresh** — `requestFocus()` → `focusAndCursorToEnd()`

9. **แก้ BLoC GetSuccess** — nested `setState(() { findFocusNext(0); })` → `focusFirstTextField(context)` ใน `addPostFrameCallback`

10. **แก้ onEditingComplete** — `someFocusNode.requestFocus()` → `focusAndCursorToEnd(someFocusNode)`

11. **แก้ dropdown onChanged** — `someFocusNode.requestFocus()` → `focusAndCursorToEnd(someFocusNode)`

12. **แก้ helper functions** — เช่น `setFocusNode(focus) { focus.requestFocus(); }` → `focusAndCursorToEnd(focus)`

13. **แก้จอที่มี `_focusAndMoveCursorToEnd` เอง** — ลบ manual cursor logic, ใช้ `focusAndCursorToEnd()` จาก focus_utils แทน

#### จอที่แก้แล้ว (Reference)

| จอ | Status |
|----|--------|
| ทุกจอ config (40+ จอ) | ✅ full migration — ไม่เหลือ active requestFocus |
| promotion_screen.dart | ✅ 6 จุด: dropdown onChanged + 5 onEditingComplete → `focusAndCursorToEnd()` |
| doc_format_screen.dart | ✅ 1 จุด: checkbox toggle → `focusAndCursorToEnd()` |
| product_barcode_screen.dart | ✅ 3 จุด: `setFocusNode()` helper + refbarcode/BOM edit → `focusAndCursorToEnd()` |
| currency_form_screen.dart | ✅ เพิ่ม import + เปลี่ยน `_focusAndMoveCursorToEnd()` ใช้ `focusAndCursorToEnd()` |
| pattern_screen.dart | ✅ เพิ่ม import + 3 จุด: add/detail buttons → `focusAndCursorToEnd()` |

**หมายเหตุ**: `searchFocusNode.requestFocus()` (focus กลับช่องค้นหาเมื่อ save/cancel) ไม่ต้องแก้ — ถูกต้องแล้ว

#### ระวัง

- ถ้าจอไม่มี `RawKeyboardListener` → wrap edit body ด้วย `Focus + FocusTraversalGroup` โดยตรง
- ถ้าจอไม่มี `fieldFocusNodes` เลย → ไม่ต้องแก้ switchToEdit/clearEditData
- ตรวจ save function name — บางจอใช้ `saveData()` แทน `saveOrUpdateData()`
- เพิ่ม `)` ปิด `FocusTraversalGroup` ให้ครบ — นับ bracket ดีๆ
