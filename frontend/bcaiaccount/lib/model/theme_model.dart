import 'dart:ui';

import 'package:flutter/animation.dart';

class ThemeModel {
  /// Background color
  late Color backgroundColor;

  /// head Title
  late Color headTitleColor;

  ///
  late Color appBarColor;

  /// ช่องป้อนข้อมูล (บังคับ)
  late Color inputTextBoxForceColor;

  /// ช่องป้อนข้อมูล (ไม่บังคับ)
  late Color inputTextBoxColor;

  /// Column Header
  late Color columnHeaderColor;

  /// Column Header Text
  late Color columnHeaderTextColor;

  /// สีีพื้น Column คู่
  late Color columnAlternateEvenColor;

  /// สีีพื้น Column คี่
  late Color columnAlternateOddColor;

  /// Background Icon
  late Color buttonIconBackgroundColor;

  /// ปุ่ม
  late Color buttonColor;

  /// Yes
  late Color buttonYesColor;

  /// No
  late Color buttonNoColor;

  /// Danger button (ลบ, delete, error action)
  late Color buttonDangerColor;

  /// Tool Bar (Edit mode)
  late Color toolBarEditModeColor;

  /// primary color
  late Color primaryColor;

  /// primary light color
  late Color primaryLightColor;

  /// secondary color
  late Color secondaryColor;

  // ================== UI Colors (สีสำหรับ UI ทั่วไป) ==================

  /// สีพื้น Card / Section container
  late Color cardColor;

  /// สีตัวอักษรหลัก
  late Color textColor;

  /// สีตัวอักษรรอง (hint, subtitle)
  late Color textSecondaryColor;

  /// สี Surface (tab bar header, containers)
  late Color surfaceColor;

  /// สีพื้น Search Bar (ทึมกว่า surfaceColor)
  late Color searchBarColor;

  /// สีเส้นขอบ / divider
  late Color dividerBorderColor;

  /// สีพื้น Dialog
  late Color dialogColor;

  /// สีพื้น Input field
  late Color inputFillColor;

  /// Gradient สำหรับ body (3 สี: top, middle, bottom)
  late List<Color> bodyGradientColors;

  // ================== Icon Colors ==================

  /// สี Icon หลัก
  late Color iconColor;

  /// สี Icon รอง (muted)
  late Color iconSecondaryColor;

  // ================== Navigation & Bar Colors ==================

  /// สีพื้น Info Bar (แถบข้อมูลร้าน/สาขา)
  late Color infoBarColor;

  /// สีพื้น Bottom Navigation Bar
  late Color bottomNavColor;

  /// สี selected item ใน Bottom Nav
  late Color bottomNavSelectedColor;

  /// สี unselected item ใน Bottom Nav
  late Color bottomNavUnselectedColor;

  // ================== On-Surface Colors (ตัวอักษร/ไอคอนบนพื้นสีเข้ม) ==================

  /// สีตัวอักษรบนพื้น primary (AppBar, info bar)
  late Color onPrimaryColor;

  /// สีตัวอักษรรองบนพื้น primary
  late Color onPrimarySecondaryColor;

  // ================== Data List Colors ==================

  /// สีพื้น hover ของแถวใน data list
  late Color rowHoverColor;

  /// สีพื้น selected ของแถวใน data list (กดเลือกดู)
  late Color rowSelectedColor;

  /// สีพื้น edit mode ของแถวใน data list (กดแก้ไข)
  late Color rowEditColor;

  // ================== Form Colors ==================

  /// สี label ของ required field (บังคับกรอก)
  late Color formLabelRequiredColor;

  /// สี label ปกติ
  late Color formLabelColor;

  /// สีเส้นขอบ field ปกติ
  late Color formBorderColor;

  /// สีเส้นขอบ field เมื่อ focus
  late Color formFocusBorderColor;

  /// สีพื้น field (ใช้ร่วมกับ filled: true)
  late Color formFillColor;

  /// สีตัวอักษรใน field
  late Color formTextColor;

  /// สี hint text
  late Color formHintColor;

  // ================== Semantic Highlight Colors ==================

  /// สีพื้น highlight เมื่อมีค่า (เช่น ราคา > 0, จำนวน > 0)
  late Color positiveHighlightColor;

  /// สีตัวอักษร highlight เมื่อมีค่า
  late Color positiveHighlightTextColor;

  /// สีพื้น highlight สำหรับข้อมูลอ้างอิง (ref barcode, info box)
  late Color infoHighlightColor;

  /// สีตัวอักษร highlight สำหรับข้อมูลอ้างอิง
  late Color infoHighlightTextColor;

  /// สีพื้น highlight เมื่อมีค่าติดลบ (error, ลบ, cash out)
  late Color negativeHighlightColor;

  /// สีตัวอักษร highlight เมื่อมีค่าติดลบ (error, ลบ, cash out)
  late Color negativeHighlightTextColor;

  /// สีพื้น highlight สำหรับคำเตือน (warning, pending)
  late Color warningHighlightColor;

  /// สีตัวอักษร highlight สำหรับคำเตือน (warning, pending)
  late Color warningHighlightTextColor;
}
