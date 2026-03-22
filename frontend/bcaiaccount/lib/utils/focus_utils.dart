import 'package:flutter/services.dart'
    show HardwareKeyboard, KeyDownEvent, KeyUpEvent, LogicalKeyboardKey;
import 'package:flutter/widgets.dart';

/// Custom FocusTraversalPolicy ที่วน Tab เฉพาะ TextField/TextFormField
/// ข้าม buttons, checkboxes, radio, switches ทั้งหมดอัตโนมัติ
/// วนลูป: field สุดท้าย → field แรก, field แรก ← field สุดท้าย
///
/// ใช้คู่กับ FocusTraversalGroup:
/// ```dart
/// FocusTraversalGroup(
///   policy: TextFieldTraversalPolicy(),
///   child: Column(children: [
///     TextField(...),       // Tab วนที่นี่
///     ElevatedButton(...),  // ข้าม
///     TextField(...),       // Tab วนที่นี่
///   ]),
/// )
/// ```
class TextFieldTraversalPolicy extends ReadingOrderTraversalPolicy {
  TextFieldTraversalPolicy();

  /// กรองเฉพาะ FocusNode ที่เป็น text input field (EditableText) ที่แก้ไขได้
  List<FocusNode> _getTextFieldNodes(Iterable<FocusNode> descendants) {
    return descendants.where((node) {
      if (node.skipTraversal) return false;
      if (!node.canRequestFocus) return false;

      final ctx = node.context;
      if (ctx == null) return false;

      // FocusNode ของ TextField ถูก attach ที่ Focus widget ข้างใน EditableText
      // ดังนั้น EditableText จะเป็น ancestor ของ FocusNode
      final editableText =
          ctx.findAncestorWidgetOfExactType<EditableText>();
      if (editableText == null) return false;
      // ข้าม read-only fields (เช่น รหัส ที่แก้ไม่ได้ตอน edit)
      if (editableText.readOnly) return false;
      return true;
    }).toList();
  }

  @override
  Iterable<FocusNode> sortDescendants(
    Iterable<FocusNode> descendants,
    FocusNode currentNode,
  ) {
    // ใช้ reading order (ตำแหน่งบนจอ) แล้วกรองเฉพาะ text fields
    final allSorted = super.sortDescendants(descendants, currentNode);
    return _getTextFieldNodes(allSorted);
  }

  @override
  bool next(FocusNode currentNode) => _move(currentNode, forward: true);

  @override
  bool previous(FocusNode currentNode) => _move(currentNode, forward: false);

  bool _move(FocusNode currentNode, {required bool forward}) {
    final scope = currentNode.nearestScope;
    if (scope == null) return false;

    final sorted = sortDescendants(scope.descendants, currentNode).toList();
    if (sorted.isEmpty) return false;

    final currentIndex = sorted.indexOf(currentNode);

    int nextIndex;
    if (currentIndex == -1) {
      // current node ไม่อยู่ใน list → focus ตัวแรก/สุดท้าย
      nextIndex = forward ? 0 : sorted.length - 1;
    } else if (forward) {
      // วนลูป: ถ้าอยู่ตัวสุดท้าย → กลับตัวแรก
      nextIndex = (currentIndex + 1) % sorted.length;
    } else {
      // วนลูป: ถ้าอยู่ตัวแรก → กลับตัวสุดท้าย
      nextIndex = (currentIndex - 1 + sorted.length) % sorted.length;
    }

    focusAndCursorToEnd(sorted[nextIndex]);
    return true;
  }
}

/// Focus node แล้ว set cursor ไปท้ายข้อความ (ไม่ select all)
/// ต้อง lookup EditableText ใน postFrameCallback เพราะ widget อาจ rebuild
/// (เช่น จอที่สร้าง TextEditingController inline ใน build)
void focusAndCursorToEnd(FocusNode node) {
  node.requestFocus();
  // ต้อง set cursor หลัง Flutter select-all เสร็จ (frame ถัดไป)
  // และต้อง lookup EditableText ใน callback เพราะ requestFocus อาจ trigger rebuild
  // ทำให้ widget instance เปลี่ยน (controller ตัวเก่า stale)
  WidgetsBinding.instance.addPostFrameCallback((_) {
    final ctx = node.context;
    if (ctx == null) return;
    final editableText = ctx.findAncestorWidgetOfExactType<EditableText>();
    if (editableText != null && editableText.controller.text.isNotEmpty) {
      editableText.controller.selection = TextSelection.collapsed(
        offset: editableText.controller.text.length,
      );
    }
  });
}

/// Focus first text field ใน FocusTraversalGroup
/// เรียกใน addPostFrameCallback หลัง setState:
/// ```dart
/// WidgetsBinding.instance.addPostFrameCallback((_) {
///   focusFirstTextField(context);
/// });
/// ```
void focusFirstTextField(BuildContext context) {
  final scope = FocusScope.of(context);
  final descendants = scope.descendants.where((node) {
    if (node.skipTraversal || !node.canRequestFocus) return false;
    final ctx = node.context;
    if (ctx == null) return false;
    final editableText =
        ctx.findAncestorWidgetOfExactType<EditableText>();
    if (editableText == null) return false;
    // ข้าม read-only fields
    if (editableText.readOnly) return false;
    return true;
  });
  if (descendants.isNotEmpty) {
    focusAndCursorToEnd(descendants.first);
  }
}

/// Focus widget wrapper ที่ intercept Tab/Shift+Tab/Enter/F10
/// ใช้คู่กับ TextFieldTraversalPolicy
///
/// ```dart
/// body: TabFocusScope(
///   onSave: () => saveOrUpdateData(),
///   child: FocusTraversalGroup(
///     policy: TextFieldTraversalPolicy(),
///     child: Form(...),
///   ),
/// )
/// ```
class TabFocusScope extends StatelessWidget {
  final Widget child;
  final VoidCallback? onSave;

  const TabFocusScope({
    super.key,
    required this.child,
    this.onSave,
  });

  @override
  Widget build(BuildContext context) {
    return Focus(
      skipTraversal: true,
      onKeyEvent: (node, event) {
        if (event is KeyUpEvent) return KeyEventResult.ignored;

        // F10 = save
        if (event.logicalKey == LogicalKeyboardKey.f10) {
          if (event is KeyDownEvent && onSave != null) {
            onSave!();
          }
          return KeyEventResult.handled;
        }

        // Tab / Enter = ใช้ FocusTraversalPolicy จัดการ
        if (event.logicalKey == LogicalKeyboardKey.tab) {
          if (event is KeyDownEvent) {
            final currentFocus = FocusManager.instance.primaryFocus;
            if (currentFocus != null) {
              if (HardwareKeyboard.instance.isShiftPressed) {
                currentFocus.previousFocus();
              } else {
                currentFocus.nextFocus();
              }
            }
          }
          return KeyEventResult.handled;
        }

        // Enter = next field (เหมือน Tab)
        if (event.logicalKey == LogicalKeyboardKey.enter) {
          if (event is KeyDownEvent) {
            final currentFocus = FocusManager.instance.primaryFocus;
            if (currentFocus != null) {
              if (HardwareKeyboard.instance.isShiftPressed) {
                currentFocus.previousFocus();
              } else {
                currentFocus.nextFocus();
              }
            }
          }
          return KeyEventResult.handled;
        }

        return KeyEventResult.ignored;
      },
      child: child,
    );
  }
}

/// สร้าง FocusNode หลายตัวที่วน Tab วนลูปกันเอง (legacy)
List<FocusNode> createTabCycleFocusNodes(int count) {
  assert(count >= 2, 'ต้องมีอย่างน้อย 2 fields ถึงจะวน Tab ได้');

  final nodes = List<FocusNode>.generate(count, (_) => FocusNode());

  for (var i = 0; i < count; i++) {
    final prevIndex = (i - 1 + count) % count;
    final nextIndex = (i + 1) % count;

    nodes[i] = FocusNode(
      onKeyEvent: (node, event) {
        if (event is KeyDownEvent &&
            event.logicalKey == LogicalKeyboardKey.tab) {
          if (HardwareKeyboard.instance.isShiftPressed) {
            nodes[prevIndex].requestFocus();
          } else {
            nodes[nextIndex].requestFocus();
          }
          return KeyEventResult.handled;
        }
        return KeyEventResult.ignored;
      },
    );
  }

  return nodes;
}

/// ติด Tab cycle ให้ FocusNode ที่มีอยู่แล้ว (legacy)
void applyTabCycle(List<FocusNode> nodes) {
  if (nodes.length < 2) return;
  final count = nodes.length;

  for (var i = 0; i < count; i++) {
    final prevIndex = (i - 1 + count) % count;
    final nextIndex = (i + 1) % count;
    final currentNode = nodes[i];

    final originalOnKeyEvent = currentNode.onKeyEvent;
    currentNode.onKeyEvent = (node, event) {
      if (event is KeyDownEvent &&
          event.logicalKey == LogicalKeyboardKey.tab) {
        if (HardwareKeyboard.instance.isShiftPressed) {
          nodes[prevIndex].requestFocus();
        } else {
          nodes[nextIndex].requestFocus();
        }
        return KeyEventResult.handled;
      }
      return originalOnKeyEvent?.call(node, event) ?? KeyEventResult.ignored;
    };
  }
}

/// Dispose FocusNode ทั้ง list — เรียกใน dispose() ของ State
void disposeFocusNodes(List<FocusNode> nodes) {
  for (final node in nodes) {
    node.dispose();
  }
}
