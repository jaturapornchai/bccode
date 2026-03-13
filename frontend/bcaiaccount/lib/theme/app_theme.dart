/// App Theme — จัดการสี, ธีม, Display Settings ทั้งหมด
///
/// แยกจาก global.dart เพื่อลดขนาดไฟล์และจัดกลุ่ม code ให้ชัดเจน
/// ใช้งานผ่าน global.theme.*, global.isDarkMode(), global.ThemeRefreshMixin ฯลฯ
/// (re-export จาก global.dart ทำให้ไม่ต้องเปลี่ยน import ในไฟล์อื่น)
library;

import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:smlaicloud/model/theme_model.dart';

// ================== Theme Instance ==================

/// Theme instance — ทุก UI ใช้สีจากที่นี่
ThemeModel theme = ThemeModel();

// ================== Theme Mode Settings ==================

/// โหมดธีม: 'light' = กลางวัน, 'dark' = กลางคืน
String displayThemeMode = 'light';

/// ตรวจว่าตอนนี้ dark mode หรือไม่
bool isDarkMode() {
  return displayThemeMode == 'dark';
}

/// Apply dark/light theme colors ตาม mode ปัจจุบัน (เรียกหลัง themeSelect)
void applyThemeMode() {
  // เรียก themeSelect ก่อนเพื่อ set brand colors (light base)
  themeSelect(_currentBrandTheme);

  if (isDarkMode()) {
    // Base colors — dark override
    theme.secondaryColor = colorFromHex("2D2D2D");
    theme.backgroundColor = colorFromHex("1A1A2E");
    theme.appBarColor = colorFromHex("0D1B2A");
    theme.headTitleColor = Colors.white;
    theme.inputTextBoxForceColor = colorFromHex("FF6B6B");
    theme.inputTextBoxColor = colorFromHex("E0E0E0");
    theme.columnHeaderColor = colorFromHex("2D3A4A");
    theme.columnHeaderTextColor = colorFromHex("E0E0E0");
    theme.columnAlternateEvenColor = colorFromHex("1E1E30"); // จางลง ใกล้ background
    theme.columnAlternateOddColor = colorFromHex("1A1A2E"); // เท่ากับ background
    theme.buttonIconBackgroundColor = colorFromHex("2D2D2D");
    theme.buttonNoColor = colorFromHex("374151");
    theme.toolBarEditModeColor = colorFromHex("1B3A5C");

    // UI Colors — dark override
    theme.cardColor = colorFromHex("252538");
    theme.textColor = colorFromHex("E0E0E0");
    theme.textSecondaryColor = Colors.grey.shade400;
    theme.surfaceColor = colorFromHex("252538");
    theme.searchBarColor = colorFromHex("141825");
    theme.dividerBorderColor = Colors.grey.shade700;
    theme.dialogColor = colorFromHex("2D2D2D");
    theme.inputFillColor = colorFromHex("2D2D2D");
    theme.bodyGradientColors = [
      colorFromHex("1A1A2E"),
      colorFromHex("16213E"),
      colorFromHex("1A1A2E"),
    ];
    // Icon Colors — dark
    theme.iconColor = colorFromHex("B0BEC5"); // blueGrey200
    theme.iconSecondaryColor = Colors.grey.shade500;
    // Navigation & Bar — dark
    theme.infoBarColor = colorFromHex("0D1B2A").withValues(alpha: 0.95);
    theme.bottomNavColor = colorFromHex("0D1B2A");
    theme.bottomNavSelectedColor = Colors.white;
    theme.bottomNavUnselectedColor = Colors.white.withValues(alpha: 0.60);
    // On-Surface — dark (เหมือน light เพราะอยู่บนพื้นเข้มเหมือนกัน)
    theme.onPrimaryColor = Colors.white;
    theme.onPrimarySecondaryColor = Colors.white70;
    // Data List — dark (จางลง ใช้ alpha ให้เนียนกับพื้น)
    theme.rowHoverColor = Colors.white.withValues(alpha: 0.06);
    theme.rowSelectedColor = Colors.cyan.withValues(alpha: 0.12);
    theme.rowEditColor = Colors.orange.withValues(alpha: 0.15);
    // Form — dark
    theme.formLabelRequiredColor = colorFromHex("FF6B6B");
    theme.formLabelColor = colorFromHex("E0E0E0");
    theme.formBorderColor = Colors.grey.shade600;
    theme.formFocusBorderColor = colorFromHex("89C2D9");
    theme.formFillColor = colorFromHex("2D2D2D");
    theme.formTextColor = colorFromHex("E0E0E0");
    theme.formHintColor = Colors.grey.shade500;
    // Semantic Highlight — dark
    theme.positiveHighlightColor = colorFromHex("1B5E20").withValues(alpha: 0.3);
    theme.positiveHighlightTextColor = colorFromHex("66BB6A");
    theme.infoHighlightColor = colorFromHex("1A237E").withValues(alpha: 0.3);
    theme.infoHighlightTextColor = colorFromHex("7986CB");
    theme.negativeHighlightColor = colorFromHex("7F1D1D").withValues(alpha: 0.3);
    theme.negativeHighlightTextColor = colorFromHex("F87171");
    theme.warningHighlightColor = colorFromHex("78350F").withValues(alpha: 0.3);
    theme.warningHighlightTextColor = colorFromHex("FBBF24");
    // Button semantic — dark
    theme.buttonDangerColor = colorFromHex("EF4444");
  }
}

/// Brand theme ที่เลือก (0=default, 1=dohome)
int _currentBrandTheme = 0;

// ================== Display Settings (เก็บในเครื่อง) ==================
/// การตั้งค่าการแสดงผล - เก็บในเครื่องแต่ละเครื่อง

/// Font Family ที่ใช้ทั้งระบบ
String displayFontFamily = 'Sarabun';

/// ระดับ Zoom (50% - 300%, default 100%)
double displayZoomLevel = 100.0;

/// รายการ Font Family ที่รองรับ (10 Popular Thai fonts from Google Fonts)
List<String> availableFontFamilies = [
  'Sarabun', // คลาสสิก เป็นทางการ - ยอดนิยมอันดับ 1
  'Kanit', // เรขาคณิต ทันสมัย
  'Prompt', // สะอาด โปรเฟสชันแนล
  'Noto Sans Thai', // มาตรฐาน Unicode
  'IBM Plex Sans Thai', // เทคโนโลยี
  'Mitr', // เป็นมิตร อ่านง่าย
  'K2D', // ทันสมัย มินิมอล
  'Bai Jamjuree', // สดใส มีชีวิตชีวา
  'Pridi', // คลาสสิก Serif-like
  'Sriracha', // ลายมือ สนุกสนาน
];

/// Notifier สำหรับแจ้งเตือนเมื่อ display settings เปลี่ยน
/// ใช้เพื่อ rebuild UI ทุกจอเมื่อตั้งค่าเปลี่ยน
final displaySettingsNotifier = ValueNotifier<int>(0);

/// คำนวณ font size ตาม zoom level
double scaledFontSize(double baseSize) {
  return baseSize * (displayZoomLevel / 100.0);
}

/// Mixin สำหรับ StatefulWidget ที่ต้อง rebuild เมื่อเปลี่ยนธีม
/// ใช้งาน: class _MyScreenState extends State<MyScreen> with ThemeRefreshMixin
/// เมื่อกดเปลี่ยนธีม (saveThemeSettings) จะ rebuild screen อัตโนมัติ
mixin ThemeRefreshMixin<T extends StatefulWidget> on State<T> {
  VoidCallback? _themeRefreshCallback;

  @override
  void initState() {
    super.initState();
    _themeRefreshCallback = _onThemeChanged;
    displaySettingsNotifier.addListener(_themeRefreshCallback!);
  }

  void _onThemeChanged() {
    if (mounted) setState(() {});
  }

  @override
  void dispose() {
    if (_themeRefreshCallback != null) {
      displaySettingsNotifier.removeListener(_themeRefreshCallback!);
    }
    super.dispose();
  }
}

// ================== PDF Settings (เก็บในเครื่อง) ==================
/// การตั้งค่าสำหรับพิมพ์ PDF - เก็บในเครื่องแต่ละเครื่อง

/// รายการ Font Family ที่รองรับสำหรับ PDF (fonts ที่มีจริงใน backend)
/// หมายเหตุ: รายการนี้ต่างจาก availableFontFamilies (สำหรับหน้าจอ) เพราะ PDF ใช้ font files ที่ต่างกัน
List<String> availablePdfFontFamilies = [
  'Sarabun', // Google Thai - คลาสสิก (Default)
  'Kanit', // Google Thai - เรขาคณิต ทันสมัย
  'Prompt', // Google Thai - สะอาด โปรเฟสชันแนล
  'Mitr', // Google Thai - เป็นมิตร อ่านง่าย
  'GoNotoCurrent', // Universal - รองรับทุกภาษา
  'NotoSansThai', // Noto Thai - ปกติ
];

/// คำนวณขนาดอื่นๆ ตาม zoom level (เช่น icon size, padding)
double scaledSize(double baseSize) {
  return baseSize * (displayZoomLevel / 100.0);
}

/// ดึง TextTheme จาก Google Fonts ตาม font family ที่เลือก
/// ใช้สำหรับ ThemeData ใน MaterialApp
TextTheme getDisplayTextTheme([TextTheme? baseTextTheme]) {
  final base = baseTextTheme ?? const TextTheme();
  switch (displayFontFamily) {
    case 'Sarabun':
      return GoogleFonts.sarabunTextTheme(base);
    case 'Kanit':
      return GoogleFonts.kanitTextTheme(base);
    case 'Prompt':
      return GoogleFonts.promptTextTheme(base);
    case 'Noto Sans Thai':
      return GoogleFonts.notoSansThaiTextTheme(base);
    case 'IBM Plex Sans Thai':
      return GoogleFonts.ibmPlexSansThaiTextTheme(base);
    case 'Mitr':
      return GoogleFonts.mitrTextTheme(base);
    case 'K2D':
      return GoogleFonts.k2dTextTheme(base);
    case 'Bai Jamjuree':
      return GoogleFonts.baiJamjureeTextTheme(base);
    case 'Pridi':
      return GoogleFonts.pridiTextTheme(base);
    case 'Sriracha':
      return GoogleFonts.srirachaTextTheme(base);
    default:
      return GoogleFonts.sarabunTextTheme(base);
  }
}

// ================== Brand Theme Colors ==================

/// กำหนดสี brand theme ตาม mode (0=default, 1=dohome)
/// เรียกจาก applyThemeMode() เสมอ ไม่ควรเรียกตรง
void themeSelect(int mode) {
  _currentBrandTheme = mode;
  switch (mode) {
    case 0:
      theme.primaryColor = colorFromHex("2A6F97");
      theme.primaryLightColor = colorFromHex("89C2D9");
      theme.secondaryColor = colorFromHex("A9D6E5");
      theme.backgroundColor = colorFromHex("E4F1F5");
      theme.appBarColor = colorFromHex("012A4A");
      theme.headTitleColor = colorFromHex("013A63");
      theme.inputTextBoxForceColor = colorFromHex("8A1606");
      theme.inputTextBoxColor = Colors.black;
      theme.columnHeaderColor = colorFromHex("89C2D9");
      theme.columnHeaderTextColor = Colors.black;
      theme.columnAlternateEvenColor = colorFromHex("F3F7FA");
      theme.columnAlternateOddColor = Colors.white;
      theme.buttonIconBackgroundColor = Colors.white;
      theme.buttonColor = colorFromHex("2A6F97");
      theme.buttonYesColor = colorFromHex("2A6F97");
      theme.buttonNoColor = colorFromHex("A9D6E5");
      theme.toolBarEditModeColor = colorFromHex("2A6F97");
    case 1: // dohome
      theme.primaryColor = colorFromHex("2A6F97");
      theme.secondaryColor = colorFromHex("A9D6E5");
      theme.backgroundColor = colorFromHex("E4F1F5");
      theme.appBarColor = colorFromHex("012A4A");
      theme.headTitleColor = colorFromHex("013A63");
      theme.inputTextBoxForceColor = colorFromHex("8A1606");
      theme.inputTextBoxColor = Colors.black;
      theme.columnHeaderColor = colorFromHex("89C2D9");
      theme.columnHeaderTextColor = Colors.black;
      theme.columnAlternateEvenColor = colorFromHex("F3F7FA");
      theme.columnAlternateOddColor = Colors.white;
      theme.buttonIconBackgroundColor = Colors.white;
      theme.buttonColor = colorFromHex("2A6F97");
      theme.buttonYesColor = colorFromHex("2A6F97");
      theme.buttonNoColor = colorFromHex("A9D6E5");
      theme.toolBarEditModeColor = colorFromHex("2A6F97");
  }

  // UI Colors — Light mode defaults
  theme.cardColor = Colors.white;
  theme.textColor = const Color(0xFF212121); // black87
  theme.textSecondaryColor = Colors.grey.shade600;
  theme.surfaceColor = Colors.grey.shade100;
  theme.searchBarColor = Colors.grey.shade200;
  theme.dividerBorderColor = Colors.grey.shade300;
  theme.dialogColor = Colors.white;
  theme.inputFillColor = Colors.white;
  theme.bodyGradientColors = [
    Colors.white,
    theme.primaryLightColor.withValues(alpha: 0.3),
    Colors.white.withValues(alpha: 0.9),
  ];
  // Icon Colors — Light
  theme.iconColor = const Color(0xFF424242); // grey800
  theme.iconSecondaryColor = Colors.grey.shade500;
  // Navigation & Bar — Light
  theme.infoBarColor = colorFromHex("D4A373").withValues(alpha: 0.95); // เดิมใช้ primaryDarkColor
  theme.bottomNavColor = colorFromHex("D4A373").withValues(alpha: 0.95);
  theme.bottomNavSelectedColor = Colors.white;
  theme.bottomNavUnselectedColor = Colors.white.withValues(alpha: 0.60);
  // On-Surface — Light (ตัวอักษรบนพื้นเข้ม เช่น AppBar, info bar)
  theme.onPrimaryColor = Colors.white;
  theme.onPrimarySecondaryColor = Colors.white70;
  // Data List — Light
  theme.rowHoverColor = Colors.blue.shade50;
  theme.rowSelectedColor = Colors.cyan.shade100;
  theme.rowEditColor = Colors.orange.shade100;
  // Form — Light
  theme.formLabelRequiredColor = colorFromHex("8A1606");
  theme.formLabelColor = Colors.black;
  theme.formBorderColor = Colors.grey.shade300;
  theme.formFocusBorderColor = colorFromHex("2A6F97");
  theme.formFillColor = Colors.white;
  theme.formTextColor = const Color(0xFF212121);
  theme.formHintColor = Colors.grey.shade500;
  // Button semantic — light
  theme.buttonDangerColor = colorFromHex("DC2626");
  // Semantic Highlight — light
  theme.positiveHighlightColor = colorFromHex("E8F5E9");
  theme.positiveHighlightTextColor = colorFromHex("2E7D32");
  theme.infoHighlightColor = colorFromHex("E8EAF6");
  theme.infoHighlightTextColor = colorFromHex("283593");
  theme.negativeHighlightColor = colorFromHex("FEF2F2");
  theme.negativeHighlightTextColor = colorFromHex("B91C1C");
  theme.warningHighlightColor = colorFromHex("FFFBEB");
  theme.warningHighlightTextColor = colorFromHex("B45309");
}

// ================== Utilities ==================

/// แปลง hex color string เป็น Color object
Color colorFromHex(String hexColor) {
  final hexCode = hexColor.replaceAll('#', '');
  return Color(int.parse('FF$hexCode', radix: 16));
}
