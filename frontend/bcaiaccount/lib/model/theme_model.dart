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

  /// Tool Bar (Edit mode)
  late Color toolBarEditModeColor;

  /// primary color
  late Color primaryColor;

  /// primary light color
  late Color primaryLightColor;

  /// secondary color
  late Color secondaryColor;
}
