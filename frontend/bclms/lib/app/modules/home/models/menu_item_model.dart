import 'package:flutter/material.dart';

/// โมเดลข้อมูลเมนูย่อยในแต่ละ Tab
class MenuItemModel {
  final String id;
  final String label;
  final IconData icon;
  final String route;
  final Color color;
  final bool isEnabled;

  const MenuItemModel({
    required this.id,
    required this.label,
    required this.icon,
    required this.route,
    this.color = const Color(0xFFB5432A),
    this.isEnabled = true,
  });
}
