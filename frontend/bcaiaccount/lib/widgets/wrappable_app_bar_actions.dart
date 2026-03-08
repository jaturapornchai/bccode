import 'package:flutter/material.dart';

/// Wrappable AppBar Actions - แก้ปัญหาจอล้น
/// ใช้ Wrap แทน Row เพื่อให้ปุ่มขึ้นบรรทัดใหม่เมื่อเต็มแถว
class WrappableAppBarActions extends StatelessWidget {
  final List<Widget> children;
  final double spacing;
  final double runSpacing;

  const WrappableAppBarActions({
    super.key,
    required this.children,
    this.spacing = 0,
    this.runSpacing = 4,
  });

  @override
  Widget build(BuildContext context) {
    return Wrap(
      spacing: spacing,
      runSpacing: runSpacing,
      alignment: WrapAlignment.end,
      crossAxisAlignment: WrapCrossAlignment.center,
      children: children,
    );
  }
}
