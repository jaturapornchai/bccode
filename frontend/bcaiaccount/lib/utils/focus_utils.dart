import 'package:flutter/services.dart'
    show HardwareKeyboard, KeyDownEvent, LogicalKeyboardKey;
import 'package:flutter/widgets.dart' show FocusNode, KeyEventResult;

/// สร้าง FocusNode หลายตัวที่วน Tab วนลูปกันเอง
/// เช่น ถ้าส่ง count=3 จะได้ [node0, node1, node2]
/// กด Tab ที่ node2 → วนกลับไป node0
/// กด Shift+Tab ที่ node0 → วนกลับไป node2
///
/// ตัวอย่างการใช้:
/// ```dart
/// late final List<FocusNode> _focusNodes;
///
/// @override
/// void initState() {
///   super.initState();
///   _focusNodes = createTabCycleFocusNodes(3);
/// }
///
/// @override
/// void dispose() {
///   disposeFocusNodes(_focusNodes);
///   super.dispose();
/// }
///
/// // ใน build:
/// TextField(focusNode: _focusNodes[0], ...),
/// TextField(focusNode: _focusNodes[1], ...),
/// TextField(focusNode: _focusNodes[2], ...),
/// ```
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

/// ติด Tab cycle ให้ FocusNode ที่มีอยู่แล้ว (เช่น จาก FieldFocusModel)
/// เรียกหลังจากสร้าง FocusNode ครบทุกตัวแล้ว
///
/// ตัวอย่าง:
/// ```dart
/// // หลัง setSystemLanguageList() ที่สร้าง fieldFocusNodes ครบ:
/// applyTabCycle(fieldFocusNodes.map((f) => f.focusNode).toList());
/// ```
void applyTabCycle(List<FocusNode> nodes) {
  if (nodes.length < 2) return;
  final count = nodes.length;

  for (var i = 0; i < count; i++) {
    final prevIndex = (i - 1 + count) % count;
    final nextIndex = (i + 1) % count;
    final currentNode = nodes[i];

    // ลบ onKeyEvent listener เดิม (ถ้ามี) แล้วใส่ใหม่ผ่าน wrapper
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
      // เรียก original handler ถ้ามี
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
