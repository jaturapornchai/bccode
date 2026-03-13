import 'dart:convert';
import 'dart:typed_data';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'package:printing/printing.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:pdf/pdf.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:shared_preferences/shared_preferences.dart';

/// ==================== PDF Theme Models ====================

/// สีหลักของ theme
class PdfThemeColors {
  final String primary;
  final String secondary;
  final String accent;
  final String danger;
  final String warning;

  const PdfThemeColors({required this.primary, required this.secondary, required this.accent, required this.danger, required this.warning});

  Map<String, dynamic> toJson() => {'primary': primary, 'secondary': secondary, 'accent': accent, 'danger': danger, 'warning': warning};

  factory PdfThemeColors.fromJson(Map<String, dynamic> json) => PdfThemeColors(
    primary: json['primary'] ?? '#2980B9',
    secondary: json['secondary'] ?? '#3498DB',
    accent: json['accent'] ?? '#27AE60',
    danger: json['danger'] ?? '#E74C3C',
    warning: json['warning'] ?? '#F39C12',
  );
}

/// การตั้งค่าส่วน header ของ PDF
class PdfHeaderTheme {
  final String titleColor;
  final String subtitleColor;
  final String lineColor;
  final double lineWidth;

  const PdfHeaderTheme({required this.titleColor, required this.subtitleColor, required this.lineColor, required this.lineWidth});

  Map<String, dynamic> toJson() => {'titleColor': titleColor, 'subtitleColor': subtitleColor, 'lineColor': lineColor, 'lineWidth': lineWidth};

  factory PdfHeaderTheme.fromJson(Map<String, dynamic> json) => PdfHeaderTheme(
    titleColor: json['titleColor'] ?? '#2980B9',
    subtitleColor: json['subtitleColor'] ?? '#646464',
    lineColor: json['lineColor'] ?? '#2980B9',
    lineWidth: (json['lineWidth'] ?? 0.5).toDouble(),
  );
}

/// การตั้งค่าส่วน section ของ PDF
class PdfSectionTheme {
  final String labelColor;
  final String textColor;
  final String backgroundColor;

  const PdfSectionTheme({required this.labelColor, required this.textColor, required this.backgroundColor});

  Map<String, dynamic> toJson() => {'labelColor': labelColor, 'textColor': textColor, 'backgroundColor': backgroundColor};

  factory PdfSectionTheme.fromJson(Map<String, dynamic> json) =>
      PdfSectionTheme(labelColor: json['labelColor'] ?? '#2980B9', textColor: json['textColor'] ?? '#3C3C3C', backgroundColor: json['backgroundColor'] ?? '#FFFFFF');
}

/// การตั้งค่าส่วนตารางของ PDF
class PdfTableTheme {
  final String headerBgColor;
  final String headerTextColor;
  final String rowTextColor;
  final String rowBgColor;
  final String rowAltBgColor;
  final String borderColor;
  final double borderWidth;

  const PdfTableTheme({
    required this.headerBgColor,
    required this.headerTextColor,
    required this.rowTextColor,
    required this.rowBgColor,
    required this.rowAltBgColor,
    required this.borderColor,
    required this.borderWidth,
  });

  Map<String, dynamic> toJson() => {
    'headerBgColor': headerBgColor,
    'headerTextColor': headerTextColor,
    'rowTextColor': rowTextColor,
    'rowBgColor': rowBgColor,
    'rowAltBgColor': rowAltBgColor,
    'borderColor': borderColor,
    'borderWidth': borderWidth,
  };

  factory PdfTableTheme.fromJson(Map<String, dynamic> json) => PdfTableTheme(
    headerBgColor: json['headerBgColor'] ?? '#2980B9',
    headerTextColor: json['headerTextColor'] ?? '#FFFFFF',
    rowTextColor: json['rowTextColor'] ?? '#000000',
    rowBgColor: json['rowBgColor'] ?? '#FFFFFF',
    rowAltBgColor: json['rowAltBgColor'] ?? '#F8FAFC',
    borderColor: json['borderColor'] ?? '#DCDCDC',
    borderWidth: (json['borderWidth'] ?? 0.1).toDouble(),
  );
}

/// การตั้งค่าส่วนสรุปของ PDF
class PdfSummaryTheme {
  final String bgColor;
  final String textColor;
  final String highlightBgColor;
  final String highlightTextColor;
  final String borderColor;

  const PdfSummaryTheme({required this.bgColor, required this.textColor, required this.highlightBgColor, required this.highlightTextColor, required this.borderColor});

  Map<String, dynamic> toJson() => {'bgColor': bgColor, 'textColor': textColor, 'highlightBgColor': highlightBgColor, 'highlightTextColor': highlightTextColor, 'borderColor': borderColor};

  factory PdfSummaryTheme.fromJson(Map<String, dynamic> json) => PdfSummaryTheme(
    bgColor: json['bgColor'] ?? '#F8FAFC',
    textColor: json['textColor'] ?? '#000000',
    highlightBgColor: json['highlightBgColor'] ?? '#2980B9',
    highlightTextColor: json['highlightTextColor'] ?? '#FFFFFF',
    borderColor: json['borderColor'] ?? '#DCDCDC',
  );
}

/// การตั้งค่าส่วน footer ของ PDF
class PdfFooterTheme {
  final String textColor;
  final String lineColor;

  const PdfFooterTheme({required this.textColor, required this.lineColor});

  Map<String, dynamic> toJson() => {'textColor': textColor, 'lineColor': lineColor};

  factory PdfFooterTheme.fromJson(Map<String, dynamic> json) => PdfFooterTheme(textColor: json['textColor'] ?? '#969696', lineColor: json['lineColor'] ?? '#DCDCDC');
}

/// โครงสร้างหลักของ PDF Theme
class PdfTheme {
  final String name;
  final PdfThemeColors colors;
  final PdfHeaderTheme header;
  final PdfSectionTheme section;
  final PdfTableTheme table;
  final PdfSummaryTheme summary;
  final PdfFooterTheme footer;

  const PdfTheme({required this.name, required this.colors, required this.header, required this.section, required this.table, required this.summary, required this.footer});

  Map<String, dynamic> toJson() => {
    'name': name,
    'colors': colors.toJson(),
    'header': header.toJson(),
    'section': section.toJson(),
    'table': table.toJson(),
    'summary': summary.toJson(),
    'footer': footer.toJson(),
  };

  factory PdfTheme.fromJson(Map<String, dynamic> json) => PdfTheme(
    name: json['name'] ?? 'default',
    colors: PdfThemeColors.fromJson(json['colors'] ?? {}),
    header: PdfHeaderTheme.fromJson(json['header'] ?? {}),
    section: PdfSectionTheme.fromJson(json['section'] ?? {}),
    table: PdfTableTheme.fromJson(json['table'] ?? {}),
    summary: PdfSummaryTheme.fromJson(json['summary'] ?? {}),
    footer: PdfFooterTheme.fromJson(json['footer'] ?? {}),
  );
}

/// ==================== PDF Font Sizes ====================

/// ขนาดตัวอักษรแต่ละส่วนของ PDF
class PdfFontSizes {
  final int header; // ขนาดตัวอักษรส่วนหัว
  final int detail; // ขนาดตัวอักษรส่วนรายละเอียด/ตาราง
  final int summary; // ขนาดตัวอักษรส่วนสรุป
  final int footer; // ขนาดตัวอักษรส่วนท้าย
  final double lineSpacing; // ระยะห่างระหว่างบรรทัด (1.0 = ปกติ, 1.5 = 1.5 เท่า, 2.0 = 2 เท่า)

  const PdfFontSizes({this.header = 10, this.detail = 7, this.summary = 9, this.footer = 8, this.lineSpacing = 1.0});

  Map<String, dynamic> toJson() => {'header': header, 'detail': detail, 'summary': summary, 'footer': footer, 'lineSpacing': lineSpacing};

  factory PdfFontSizes.fromJson(Map<String, dynamic> json) => PdfFontSizes(
        header: json['header'] ?? 10,
        detail: json['detail'] ?? 7,
        summary: json['summary'] ?? 9,
        footer: json['footer'] ?? 8,
        lineSpacing: (json['lineSpacing'] ?? 1.0).toDouble(),
      );

  /// ค่าเริ่มต้น
  static const PdfFontSizes defaultSizes = PdfFontSizes();

  /// คัดลอกพร้อมแก้ไขบางค่า
  PdfFontSizes copyWith({int? header, int? detail, int? summary, int? footer, double? lineSpacing}) {
    return PdfFontSizes(
      header: header ?? this.header,
      detail: detail ?? this.detail,
      summary: summary ?? this.summary,
      footer: footer ?? this.footer,
      lineSpacing: lineSpacing ?? this.lineSpacing,
    );
  }
}

/// ==================== PDF Language ====================

/// รายการภาษาที่รองรับ (ตาม assets/language และ assets/flags)
class PdfSupportedLanguages {
  static const List<Map<String, String>> all = [
    {'code': 'th', 'name': 'ไทย', 'flag': 'assets/flags/th.png'},
    {'code': 'en', 'name': 'English', 'flag': 'assets/flags/en.png'},
    {'code': 'cn', 'name': '中文', 'flag': 'assets/flags/cn.png'},
    {'code': 'ja', 'name': '日本語', 'flag': 'assets/flags/ja.png'},
    {'code': 'ko', 'name': '한국어', 'flag': 'assets/flags/ko.png'},
    {'code': 'lo', 'name': 'ລາວ', 'flag': 'assets/flags/lo.png'},
    {'code': 'km', 'name': 'ខ្មែរ', 'flag': 'assets/flags/km.png'},
    {'code': 'my', 'name': 'မြန်မာ', 'flag': 'assets/flags/my.png'},
    {'code': 'vi', 'name': 'Tiếng Việt', 'flag': 'assets/flags/vi.png'},
  ];

  /// ภาษาหลัก (ไทย, อังกฤษ) - ไม่ต้องใช้ AI แปล
  static const List<String> primaryLanguages = ['th', 'en'];

  /// ภาษาที่ต้องใช้ AI แปล
  static List<Map<String, String>> get aiTranslateLanguages {
    return all.where((lang) => !primaryLanguages.contains(lang['code'])).toList();
  }

  /// ดึงชื่อภาษาตาม code
  static String getLanguageName(String code) {
    final lang = all.firstWhere((l) => l['code'] == code, orElse: () => {'name': code});
    return lang['name'] ?? code;
  }

  /// ดึง flag path ตาม code
  static String getFlagPath(String code) {
    final lang = all.firstWhere((l) => l['code'] == code, orElse: () => {'flag': 'assets/flags/th.png'});
    return lang['flag'] ?? 'assets/flags/th.png';
  }
}

/// การตั้งค่าภาษาสำหรับ PDF
class PdfLanguageSettings {
  final String languageCode; // รหัสภาษา: th, en, cn, ja, ko, lo, km, my, vi
  final bool useAiTranslate; // ใช้ AI แปลภาษา (สำหรับภาษาที่ไม่ใช่ th, en)

  const PdfLanguageSettings({this.languageCode = 'th', this.useAiTranslate = false});

  /// แปลงเป็น string สำหรับ payload
  /// ถ้าเป็นภาษาหลัก (th, en) ส่งตรงๆ
  /// ถ้าเป็นภาษาอื่น และ useAiTranslate = true ส่งเป็น "ai:xx"
  String toLanguageString() {
    if (PdfSupportedLanguages.primaryLanguages.contains(languageCode)) {
      return languageCode;
    }
    // ภาษาอื่นๆ ต้องใช้ AI แปล
    return 'ai:$languageCode';
  }

  /// คัดลอกพร้อมแก้ไขบางค่า
  PdfLanguageSettings copyWith({String? languageCode, bool? useAiTranslate}) {
    return PdfLanguageSettings(languageCode: languageCode ?? this.languageCode, useAiTranslate: useAiTranslate ?? this.useAiTranslate);
  }

  /// ค่าเริ่มต้น
  static const PdfLanguageSettings defaultSettings = PdfLanguageSettings();

  /// ตรวจสอบว่าเป็นภาษาหลักหรือไม่
  bool get isPrimaryLanguage => PdfSupportedLanguages.primaryLanguages.contains(languageCode);

  /// ชื่อแสดงผลของภาษา
  String get displayName => PdfSupportedLanguages.getLanguageName(languageCode);

  /// path ของธง
  String get flagPath => PdfSupportedLanguages.getFlagPath(languageCode);
}

/// ==================== PDF Date Format ====================

/// รูปแบบวันที่
enum PdfDateFormat {
  // รูปแบบ slash (/)
  ddmmyy, // 28/11/67 หรือ 28/11/24
  ddmmyyyy, // 28/11/2567 หรือ 28/11/2024
  mmddyyyy, // 11/28/2024 (US format - ค.ศ. เท่านั้น)
  // รูปแบบ dash (-)
  ddmmyyDash, // 28-11-67 หรือ 28-11-24
  ddmmyyyyDash, // 28-11-2567 หรือ 28-11-2024
  yyyymmdd, // 2024-11-28 (ISO format - ค.ศ. เท่านั้น)
  // รูปแบบชื่อเดือนย่อ
  ddMmmYY, // 28 พ.ย. 67 หรือ 28 Nov 24
  ddMmmYYYY, // 28 พ.ย. 2567 หรือ 28 Nov 2024
  // รูปแบบชื่อเดือนเต็ม
  ddMmmmYY, // 28 พฤศจิกายน 67 หรือ 28 November 24
  ddMmmmYYYY, // 28 พฤศจิกายน 2567 หรือ 28 November 2024
}

/// รูปแบบปี
enum PdfYearFormat {
  buddhist, // พ.ศ. (2568)
  christian, // ค.ศ. (2025)
}

/// การตั้งค่ารูปแบบวันที่สำหรับ PDF
class PdfDateSettings {
  final PdfDateFormat dateFormat;
  final PdfYearFormat yearFormat;

  const PdfDateSettings({this.dateFormat = PdfDateFormat.ddmmyyyy, this.yearFormat = PdfYearFormat.buddhist});

  Map<String, dynamic> toJson() => {'dateFormat': dateFormat.name, 'yearFormat': yearFormat.name};

  factory PdfDateSettings.fromJson(Map<String, dynamic> json) => PdfDateSettings(
    dateFormat: PdfDateFormat.values.firstWhere((e) => e.name == json['dateFormat'], orElse: () => PdfDateFormat.ddmmyyyy),
    yearFormat: PdfYearFormat.values.firstWhere((e) => e.name == json['yearFormat'], orElse: () => PdfYearFormat.buddhist),
  );

  /// ค่าเริ่มต้น
  static const PdfDateSettings defaultSettings = PdfDateSettings();

  /// คัดลอกพร้อมแก้ไขบางค่า
  PdfDateSettings copyWith({PdfDateFormat? dateFormat, PdfYearFormat? yearFormat}) {
    return PdfDateSettings(dateFormat: dateFormat ?? this.dateFormat, yearFormat: yearFormat ?? this.yearFormat);
  }

  /// ชื่อแสดงผลของรูปแบบวันที่ (ใช้ format pattern ให้เข้าใจง่าย)
  static String getDateFormatDisplayName(PdfDateFormat format, PdfYearFormat yearFormat) {
    final isBuddhist = yearFormat == PdfYearFormat.buddhist;
    final year2 = isBuddhist ? 'BB' : 'YY';
    final year4 = isBuddhist ? 'BBBB' : 'YYYY';

    switch (format) {
      case PdfDateFormat.ddmmyy:
        return 'DD/MM/$year2';
      case PdfDateFormat.ddmmyyyy:
        return 'DD/MM/$year4';
      case PdfDateFormat.mmddyyyy:
        return 'MM/DD/YYYY'; // US format ค.ศ. เท่านั้น
      case PdfDateFormat.ddmmyyDash:
        return 'DD-MM-$year2';
      case PdfDateFormat.ddmmyyyyDash:
        return 'DD-MM-$year4';
      case PdfDateFormat.yyyymmdd:
        return 'YYYY-MM-DD'; // ISO format ค.ศ. เท่านั้น
      case PdfDateFormat.ddMmmYY:
        return 'DD MMM $year2';
      case PdfDateFormat.ddMmmYYYY:
        return 'DD MMM $year4';
      case PdfDateFormat.ddMmmmYY:
        return 'DD MMMM $year2';
      case PdfDateFormat.ddMmmmYYYY:
        return 'DD MMMM $year4';
    }
  }

  /// แปลงเป็น format string สำหรับ backend
  /// ใช้ BB/BBBB สำหรับปี พ.ศ. และ YY/YYYY สำหรับปี ค.ศ.
  String toFormatString() {
    final isBuddhist = yearFormat == PdfYearFormat.buddhist;
    final year2 = isBuddhist ? 'BB' : 'YY';
    final year4 = isBuddhist ? 'BBBB' : 'YYYY';

    switch (dateFormat) {
      case PdfDateFormat.ddmmyy:
        return 'DD/MM/$year2';
      case PdfDateFormat.ddmmyyyy:
        return 'DD/MM/$year4';
      case PdfDateFormat.mmddyyyy:
        return 'MM/DD/YYYY'; // US format ค.ศ. เท่านั้น
      case PdfDateFormat.ddmmyyDash:
        return 'DD-MM-$year2';
      case PdfDateFormat.ddmmyyyyDash:
        return 'DD-MM-$year4';
      case PdfDateFormat.yyyymmdd:
        return 'YYYY-MM-DD'; // ISO format ค.ศ. เท่านั้น
      case PdfDateFormat.ddMmmYY:
        return 'DD MMM $year2';
      case PdfDateFormat.ddMmmYYYY:
        return 'DD MMM $year4';
      case PdfDateFormat.ddMmmmYY:
        return 'DD MMMM $year2';
      case PdfDateFormat.ddMmmmYYYY:
        return 'DD MMMM $year4';
    }
  }
}

/// ==================== PDF Template ====================

/// รูปแบบการจัดวาง Header
enum PdfHeaderLayout {
  standard, // โลโก้ซ้าย ข้อมูลขวา
  centered, // โลโก้กลาง ข้อมูลด้านล่าง
  minimal, // ข้อมูลบรรทัดเดียว
  split, // โลโก้ซ้าย เลขที่เอกสารขวา ข้อมูลด้านล่าง
  banner, // แถบสีเต็มความกว้าง
  leftAligned, // ข้อมูลชิดซ้ายทั้งหมด
  twoColumn, // 2 คอลัมน์ ซ้าย-ขวา
}

/// รูปแบบการจัดวางตาราง
enum PdfTableStyle {
  bordered, // มีเส้นขอบทุกด้าน
  striped, // แถวสลับสี
  minimal, // เส้นขอบน้อย (บน-ล่างเท่านั้น)
  clean, // ไม่มีเส้นขอบแนวตั้ง
  modern, // เส้นขอบบาง header สี
  elegant, // header โปร่ง เส้นบาง
  compact, // ระยะห่างน้อย
}

/// รูปแบบการจัดวางสรุป
enum PdfSummaryLayout {
  right, // ชิดขวา
  full, // เต็มความกว้าง
  boxed, // กล่องมีขอบ
  minimal, // แบบเรียบง่าย
  twoColumn, // 2 คอลัมน์
  highlighted, // เน้นยอดรวม
}

/// รูปแบบการจัดวาง Footer
enum PdfFooterStyle {
  standard, // เลขหน้ากลาง
  detailed, // วันที่ซ้าย เลขหน้าขวา
  minimal, // เลขหน้าขวา
  branded, // มีชื่อบริษัท
  withSignature, // มีช่องลงนาม
  withTerms, // มีเงื่อนไข
}

/// รูปแบบเส้นขอบ
enum PdfBorderStyle {
  none, // ไม่มีเส้นขอบ
  solid, // เส้นทึบ
  dashed, // เส้นประ
  double_, // เส้นคู่
  rounded, // มุมโค้ง
}

/// รูปแบบ Logo
enum PdfLogoStyle {
  normal, // ขนาดปกติ
  large, // ขนาดใหญ่
  small, // ขนาดเล็ก
  circular, // วงกลม
  withBackground, // มีพื้นหลัง
}

/// Template สำหรับใบสั่งซื้อ
class PdfTemplate {
  final String id;
  final String name;
  final String description;
  // Layout settings
  final PdfHeaderLayout headerLayout;
  final PdfTableStyle tableStyle;
  final PdfSummaryLayout summaryLayout;
  final PdfFooterStyle footerStyle;
  // Display options
  final bool showLogo;
  final PdfLogoStyle logoStyle;
  final bool showWatermark;
  final String watermarkText;
  final bool showBorder;
  final PdfBorderStyle borderStyle;
  final bool showDocTitle;
  final bool showCompanyInfo;
  final bool showCustomerInfo;
  final bool showPaymentInfo;
  final bool showNotes;
  final bool showSignature;
  // Spacing
  final double headerSpacing;
  final double tableSpacing;
  final double summarySpacing;
  final double marginTop;
  final double marginBottom;
  final double marginLeft;
  final double marginRight;
  // Table options
  final bool showRowNumber;
  final bool showUnitPrice;
  final bool showDiscount;
  final bool showTax;
  final bool alternateRowColor;
  // Header options
  final bool showHeaderLine;
  final double headerLineWidth;
  final bool showHeaderBackground;
  // Summary options
  final bool showSubtotal;
  final bool showTotalDiscount;
  final bool showTotalTax;
  final bool highlightTotal;

  const PdfTemplate({
    required this.id,
    required this.name,
    required this.description,
    required this.headerLayout,
    required this.tableStyle,
    required this.summaryLayout,
    required this.footerStyle,
    this.showLogo = true,
    this.logoStyle = PdfLogoStyle.normal,
    this.showWatermark = false,
    this.watermarkText = '',
    this.showBorder = false,
    this.borderStyle = PdfBorderStyle.none,
    this.showDocTitle = true,
    this.showCompanyInfo = true,
    this.showCustomerInfo = true,
    this.showPaymentInfo = true,
    this.showNotes = true,
    this.showSignature = false,
    this.headerSpacing = 20,
    this.tableSpacing = 15,
    this.summarySpacing = 15,
    this.marginTop = 20,
    this.marginBottom = 20,
    this.marginLeft = 20,
    this.marginRight = 20,
    this.showRowNumber = true,
    this.showUnitPrice = true,
    this.showDiscount = true,
    this.showTax = true,
    this.alternateRowColor = true,
    this.showHeaderLine = true,
    this.headerLineWidth = 1.0,
    this.showHeaderBackground = false,
    this.showSubtotal = true,
    this.showTotalDiscount = true,
    this.showTotalTax = true,
    this.highlightTotal = true,
  });

  Map<String, dynamic> toJson() => {
    'id': id,
    'name': name,
    'headerLayout': headerLayout.name,
    'tableStyle': tableStyle.name,
    'summaryLayout': summaryLayout.name,
    'footerStyle': footerStyle.name,
    'showLogo': showLogo,
    'logoStyle': logoStyle.name,
    'showWatermark': showWatermark,
    'watermarkText': watermarkText,
    'showBorder': showBorder,
    'borderStyle': borderStyle.name,
    'showDocTitle': showDocTitle,
    'showCompanyInfo': showCompanyInfo,
    'showCustomerInfo': showCustomerInfo,
    'showPaymentInfo': showPaymentInfo,
    'showNotes': showNotes,
    'showSignature': showSignature,
    'headerSpacing': headerSpacing,
    'tableSpacing': tableSpacing,
    'summarySpacing': summarySpacing,
    'marginTop': marginTop,
    'marginBottom': marginBottom,
    'marginLeft': marginLeft,
    'marginRight': marginRight,
    'showRowNumber': showRowNumber,
    'showUnitPrice': showUnitPrice,
    'showDiscount': showDiscount,
    'showTax': showTax,
    'alternateRowColor': alternateRowColor,
    'showHeaderLine': showHeaderLine,
    'headerLineWidth': headerLineWidth,
    'showHeaderBackground': showHeaderBackground,
    'showSubtotal': showSubtotal,
    'showTotalDiscount': showTotalDiscount,
    'showTotalTax': showTotalTax,
    'highlightTotal': highlightTotal,
  };
}

/// ==================== Predefined Templates ====================

class PdfPredefinedTemplates {
  /// Template 1: มาตรฐาน - ใช้งานทั่วไป
  /// Layout: โลโก้ซ้าย ข้อมูลขวา, ตารางมีเส้นขอบครบ, สรุปชิดขวา
  static const PdfTemplate standard = PdfTemplate(
    id: 'standard',
    name: 'มาตรฐาน',
    description: 'รูปแบบทั่วไป เหมาะกับทุกประเภทธุรกิจ',
    headerLayout: PdfHeaderLayout.standard,
    tableStyle: PdfTableStyle.bordered,
    summaryLayout: PdfSummaryLayout.right,
    footerStyle: PdfFooterStyle.standard,
    showLogo: true,
    logoStyle: PdfLogoStyle.normal,
    showHeaderLine: true,
    headerLineWidth: 1.0,
    alternateRowColor: true,
    showRowNumber: true,
  );

  /// Template 2: ทันสมัย - ดีไซน์สะอาดตา
  /// Layout: Header แบ่ง 2 ส่วน, ตารางเส้นบาง header สี, สรุปแบบกล่อง
  static const PdfTemplate modern = PdfTemplate(
    id: 'modern',
    name: 'ทันสมัย',
    description: 'ดีไซน์สะอาดตา เส้นบางเบา',
    headerLayout: PdfHeaderLayout.split,
    tableStyle: PdfTableStyle.modern,
    summaryLayout: PdfSummaryLayout.boxed,
    footerStyle: PdfFooterStyle.minimal,
    showLogo: true,
    logoStyle: PdfLogoStyle.normal,
    showHeaderLine: false,
    showHeaderBackground: true,
    alternateRowColor: false,
    showRowNumber: false,
    headerSpacing: 25,
    borderStyle: PdfBorderStyle.rounded,
  );

  /// Template 3: มืออาชีพ - เหมาะกับองค์กร
  /// Layout: Header แบบ Banner สีเต็ม, ตารางแถวสลับสี, สรุปเต็มความกว้าง
  static PdfTemplate professional = PdfTemplate(
    id: 'professional',
    name: global.language('theme_professional'),
    description: 'เหมาะกับเอกสารทางการ',
    headerLayout: PdfHeaderLayout.banner,
    tableStyle: PdfTableStyle.striped,
    summaryLayout: PdfSummaryLayout.full,
    footerStyle: PdfFooterStyle.detailed,
    showLogo: true,
    logoStyle: PdfLogoStyle.withBackground,
    showBorder: true,
    borderStyle: PdfBorderStyle.solid,
    showHeaderLine: true,
    headerLineWidth: 2.0,
    showHeaderBackground: true,
    alternateRowColor: true,
    showSignature: true,
    highlightTotal: true,
  );

  /// Template 4: กะทัดรัด - ประหยัดพื้นที่
  /// Layout: Header บรรทัดเดียว, ตารางไม่มีเส้นแนวตั้ง, margin น้อย
  static const PdfTemplate compact = PdfTemplate(
    id: 'compact',
    name: 'กะทัดรัด',
    description: 'ประหยัดพื้นที่ เหมาะกับรายการเยอะ',
    headerLayout: PdfHeaderLayout.minimal,
    tableStyle: PdfTableStyle.compact,
    summaryLayout: PdfSummaryLayout.minimal,
    footerStyle: PdfFooterStyle.minimal,
    showLogo: true,
    logoStyle: PdfLogoStyle.small,
    showHeaderLine: false,
    alternateRowColor: false,
    showRowNumber: false,
    showNotes: false,
    showPaymentInfo: false,
    headerSpacing: 8,
    tableSpacing: 5,
    summarySpacing: 5,
    marginTop: 10,
    marginBottom: 10,
    marginLeft: 15,
    marginRight: 15,
  );

  /// Template 5: เรียบง่าย - ใช้งานง่าย
  /// Layout: ไม่มีโลโก้, ตารางเส้นขอบน้อย, ข้อมูลจำเป็นเท่านั้น
  static const PdfTemplate simple = PdfTemplate(
    id: 'simple',
    name: 'เรียบง่าย',
    description: 'ข้อมูลจำเป็นเท่านั้น ไม่รกตา',
    headerLayout: PdfHeaderLayout.leftAligned,
    tableStyle: PdfTableStyle.minimal,
    summaryLayout: PdfSummaryLayout.right,
    footerStyle: PdfFooterStyle.minimal,
    showLogo: false,
    showHeaderLine: true,
    headerLineWidth: 0.5,
    alternateRowColor: false,
    showRowNumber: false,
    showDiscount: false,
    showNotes: false,
    showPaymentInfo: false,
    showCompanyInfo: true,
    showCustomerInfo: true,
  );

  /// Template 6: คลาสสิก - รูปแบบดั้งเดิม
  /// Layout: โลโก้กลาง, ตารางเส้นขอบเต็ม, เส้นคู่, มีช่องลงนาม
  static const PdfTemplate classic = PdfTemplate(
    id: 'classic',
    name: 'คลาสสิก',
    description: 'รูปแบบดั้งเดิม คุ้นเคย',
    headerLayout: PdfHeaderLayout.centered,
    tableStyle: PdfTableStyle.bordered,
    summaryLayout: PdfSummaryLayout.full,
    footerStyle: PdfFooterStyle.withSignature,
    showLogo: true,
    logoStyle: PdfLogoStyle.large,
    showBorder: true,
    borderStyle: PdfBorderStyle.double_,
    showHeaderLine: true,
    headerLineWidth: 1.5,
    alternateRowColor: true,
    showSignature: true,
    headerSpacing: 30,
    marginTop: 25,
    marginBottom: 25,
  );

  /// Template 7: พรีเมียม - หรูหรา
  /// Layout: Banner header, ตารางสไตล์โมเดิร์น, มี watermark, เน้นยอดรวม
  static const PdfTemplate premium = PdfTemplate(
    id: 'premium',
    name: 'พรีเมียม',
    description: 'ดีไซน์หรูหรา เหมาะกับลูกค้า VIP',
    headerLayout: PdfHeaderLayout.banner,
    tableStyle: PdfTableStyle.elegant,
    summaryLayout: PdfSummaryLayout.highlighted,
    footerStyle: PdfFooterStyle.branded,
    showLogo: true,
    logoStyle: PdfLogoStyle.withBackground,
    showWatermark: true,
    watermarkText: 'PREMIUM',
    showBorder: true,
    borderStyle: PdfBorderStyle.rounded,
    showHeaderLine: true,
    headerLineWidth: 2.0,
    showHeaderBackground: true,
    alternateRowColor: false,
    highlightTotal: true,
    showSignature: true,
    headerSpacing: 25,
    marginTop: 25,
    marginBottom: 25,
    marginLeft: 25,
    marginRight: 25,
  );

  /// Template 8: มินิมอล - น้อยแต่มาก
  /// Layout: ไม่มีโลโก้, เส้นน้อยที่สุด, whitespace เยอะ
  static const PdfTemplate minimalist = PdfTemplate(
    id: 'minimalist',
    name: 'มินิมอล',
    description: 'เรียบง่ายที่สุด ดูสะอาดตา',
    headerLayout: PdfHeaderLayout.minimal,
    tableStyle: PdfTableStyle.minimal,
    summaryLayout: PdfSummaryLayout.minimal,
    footerStyle: PdfFooterStyle.minimal,
    showLogo: false,
    showBorder: false,
    showHeaderLine: false,
    alternateRowColor: false,
    showRowNumber: false,
    showDiscount: false,
    showTax: false,
    showSubtotal: false,
    showNotes: false,
    showPaymentInfo: false,
    highlightTotal: false,
    headerSpacing: 15,
    tableSpacing: 10,
    summarySpacing: 10,
    marginTop: 30,
    marginBottom: 30,
    marginLeft: 30,
    marginRight: 30,
  );

  /// Template 9: องค์กร - สำหรับบริษัทใหญ่
  /// Layout: Header 2 คอลัมน์, ตารางแถวสลับสี, รายละเอียดครบถ้วน, มีเงื่อนไข
  static const PdfTemplate corporate = PdfTemplate(
    id: 'corporate',
    name: 'องค์กร',
    description: 'เหมาะกับบริษัทขนาดใหญ่',
    headerLayout: PdfHeaderLayout.twoColumn,
    tableStyle: PdfTableStyle.striped,
    summaryLayout: PdfSummaryLayout.twoColumn,
    footerStyle: PdfFooterStyle.withTerms,
    showLogo: true,
    logoStyle: PdfLogoStyle.normal,
    showWatermark: true,
    watermarkText: 'CONFIDENTIAL',
    showBorder: true,
    borderStyle: PdfBorderStyle.solid,
    showHeaderLine: true,
    headerLineWidth: 1.0,
    showHeaderBackground: true,
    alternateRowColor: true,
    showRowNumber: true,
    showSignature: true,
    showNotes: true,
    showPaymentInfo: true,
    highlightTotal: true,
  );

  /// Template 10: สร้างสรรค์ - โดดเด่นไม่เหมือนใคร
  /// Layout: Banner สีเด่น, ตารางสะอาด, สรุปแบบกล่องเน้น, footer มีแบรนด์
  static const PdfTemplate creative = PdfTemplate(
    id: 'creative',
    name: 'สร้างสรรค์',
    description: 'ดีไซน์ไม่เหมือนใคร โดดเด่น',
    headerLayout: PdfHeaderLayout.banner,
    tableStyle: PdfTableStyle.clean,
    summaryLayout: PdfSummaryLayout.highlighted,
    footerStyle: PdfFooterStyle.branded,
    showLogo: true,
    logoStyle: PdfLogoStyle.circular,
    showBorder: false,
    showHeaderLine: false,
    showHeaderBackground: true,
    alternateRowColor: true,
    showRowNumber: false,
    highlightTotal: true,
    headerSpacing: 20,
    summarySpacing: 20,
    marginLeft: 20,
    marginRight: 20,
  );

  /// รายการ template ทั้งหมด
  static List<PdfTemplate> get all => [standard, modern, professional, compact, simple, classic, premium, minimalist, corporate, creative];

  /// ดึง template ตาม id
  static PdfTemplate getById(String id) {
    return all.firstWhere((t) => t.id == id, orElse: () => standard);
  }

  /// Default template
  static const PdfTemplate defaultTemplate = standard;
}

/// ==================== Predefined Themes ====================

class PdfPredefinedThemes {
  /// Theme modern-blue - ฟ้าทันสมัย (default)
  static const PdfTheme modernBlueTheme = PdfTheme(
    name: 'modern-blue',
    colors: PdfThemeColors(primary: '#2980B9', secondary: '#3498DB', accent: '#27AE60', danger: '#E74C3C', warning: '#F39C12'),
    header: PdfHeaderTheme(titleColor: '#2980B9', subtitleColor: '#646464', lineColor: '#2980B9', lineWidth: 0.5),
    section: PdfSectionTheme(labelColor: '#2980B9', textColor: '#3C3C3C', backgroundColor: '#FFFFFF'),
    table: PdfTableTheme(headerBgColor: '#2980B9', headerTextColor: '#FFFFFF', rowTextColor: '#000000', rowBgColor: '#FFFFFF', rowAltBgColor: '#F8FAFC', borderColor: '#DCDCDC', borderWidth: 0.1),
    summary: PdfSummaryTheme(bgColor: '#F8FAFC', textColor: '#000000', highlightBgColor: '#2980B9', highlightTextColor: '#FFFFFF', borderColor: '#DCDCDC'),
    footer: PdfFooterTheme(textColor: '#969696', lineColor: '#DCDCDC'),
  );

  /// Theme professional - เทาเข้ม ดูเป็นทางการ
  static const PdfTheme professionalTheme = PdfTheme(
    name: 'professional',
    colors: PdfThemeColors(primary: '#34495E', secondary: '#5D6D7E', accent: '#1ABC9C', danger: '#E74C3C', warning: '#F39C12'),
    header: PdfHeaderTheme(titleColor: '#34495E', subtitleColor: '#7F8C8D', lineColor: '#34495E', lineWidth: 0.5),
    section: PdfSectionTheme(labelColor: '#34495E', textColor: '#2C3E50', backgroundColor: '#FFFFFF'),
    table: PdfTableTheme(headerBgColor: '#34495E', headerTextColor: '#FFFFFF', rowTextColor: '#2C3E50', rowBgColor: '#FFFFFF', rowAltBgColor: '#F8F9FA', borderColor: '#E5E8E8', borderWidth: 0.1),
    summary: PdfSummaryTheme(bgColor: '#F8F9FA', textColor: '#2C3E50', highlightBgColor: '#34495E', highlightTextColor: '#FFFFFF', borderColor: '#E5E8E8'),
    footer: PdfFooterTheme(textColor: '#AEB6BF', lineColor: '#E5E8E8'),
  );

  /// Theme nature - เขียวธรรมชาติ
  static const PdfTheme natureTheme = PdfTheme(
    name: 'nature',
    colors: PdfThemeColors(primary: '#27AE60', secondary: '#2ECC71', accent: '#16A085', danger: '#C0392B', warning: '#D35400'),
    header: PdfHeaderTheme(titleColor: '#27AE60', subtitleColor: '#7F8C8D', lineColor: '#27AE60', lineWidth: 0.5),
    section: PdfSectionTheme(labelColor: '#27AE60', textColor: '#2C3E50', backgroundColor: '#FFFFFF'),
    table: PdfTableTheme(headerBgColor: '#27AE60', headerTextColor: '#FFFFFF', rowTextColor: '#2C3E50', rowBgColor: '#FFFFFF', rowAltBgColor: '#F0FFF4', borderColor: '#BDC3C7', borderWidth: 0.1),
    summary: PdfSummaryTheme(bgColor: '#F0FFF4', textColor: '#2C3E50', highlightBgColor: '#27AE60', highlightTextColor: '#FFFFFF', borderColor: '#BDC3C7'),
    footer: PdfFooterTheme(textColor: '#95A5A6', lineColor: '#BDC3C7'),
  );

  /// Theme warm - ส้มอบอุ่น
  static const PdfTheme warmTheme = PdfTheme(
    name: 'warm',
    colors: PdfThemeColors(primary: '#E67E22', secondary: '#F39C12', accent: '#D35400', danger: '#C0392B', warning: '#F1C40F'),
    header: PdfHeaderTheme(titleColor: '#E67E22', subtitleColor: '#7F8C8D', lineColor: '#E67E22', lineWidth: 0.5),
    section: PdfSectionTheme(labelColor: '#E67E22', textColor: '#2C3E50', backgroundColor: '#FFFFFF'),
    table: PdfTableTheme(headerBgColor: '#E67E22', headerTextColor: '#FFFFFF', rowTextColor: '#2C3E50', rowBgColor: '#FFFFFF', rowAltBgColor: '#FEF5E7', borderColor: '#D5D8DC', borderWidth: 0.1),
    summary: PdfSummaryTheme(bgColor: '#FEF5E7', textColor: '#2C3E50', highlightBgColor: '#E67E22', highlightTextColor: '#FFFFFF', borderColor: '#D5D8DC'),
    footer: PdfFooterTheme(textColor: '#95A5A6', lineColor: '#D5D8DC'),
  );

  /// Theme classic-red - แดงคลาสสิก
  static const PdfTheme classicRedTheme = PdfTheme(
    name: 'classic-red',
    colors: PdfThemeColors(primary: '#C0392B', secondary: '#E74C3C', accent: '#2980B9', danger: '#922B21', warning: '#D68910'),
    header: PdfHeaderTheme(titleColor: '#C0392B', subtitleColor: '#5D6D7E', lineColor: '#C0392B', lineWidth: 0.5),
    section: PdfSectionTheme(labelColor: '#C0392B', textColor: '#1C2833', backgroundColor: '#FFFFFF'),
    table: PdfTableTheme(headerBgColor: '#C0392B', headerTextColor: '#FFFFFF', rowTextColor: '#1C2833', rowBgColor: '#FFFFFF', rowAltBgColor: '#FDEDEC', borderColor: '#D5D8DC', borderWidth: 0.1),
    summary: PdfSummaryTheme(bgColor: '#FDEDEC', textColor: '#1C2833', highlightBgColor: '#C0392B', highlightTextColor: '#FFFFFF', borderColor: '#D5D8DC'),
    footer: PdfFooterTheme(textColor: '#99A3A4', lineColor: '#D5D8DC'),
  );

  /// Theme purple - ม่วงหรูหรา
  static const PdfTheme purpleTheme = PdfTheme(
    name: 'purple',
    colors: PdfThemeColors(primary: '#8E44AD', secondary: '#9B59B6', accent: '#3498DB', danger: '#E74C3C', warning: '#F1C40F'),
    header: PdfHeaderTheme(titleColor: '#8E44AD', subtitleColor: '#7F8C8D', lineColor: '#8E44AD', lineWidth: 0.5),
    section: PdfSectionTheme(labelColor: '#8E44AD', textColor: '#2C3E50', backgroundColor: '#FFFFFF'),
    table: PdfTableTheme(headerBgColor: '#8E44AD', headerTextColor: '#FFFFFF', rowTextColor: '#2C3E50', rowBgColor: '#FFFFFF', rowAltBgColor: '#F5EEF8', borderColor: '#D5D8DC', borderWidth: 0.1),
    summary: PdfSummaryTheme(bgColor: '#F5EEF8', textColor: '#2C3E50', highlightBgColor: '#8E44AD', highlightTextColor: '#FFFFFF', borderColor: '#D5D8DC'),
    footer: PdfFooterTheme(textColor: '#95A5A6', lineColor: '#D5D8DC'),
  );

  /// Theme teal - เขียวน้ำทะเล
  static const PdfTheme tealTheme = PdfTheme(
    name: 'teal',
    colors: PdfThemeColors(primary: '#16A085', secondary: '#1ABC9C', accent: '#2980B9', danger: '#E74C3C', warning: '#F39C12'),
    header: PdfHeaderTheme(titleColor: '#16A085', subtitleColor: '#7F8C8D', lineColor: '#16A085', lineWidth: 0.5),
    section: PdfSectionTheme(labelColor: '#16A085', textColor: '#2C3E50', backgroundColor: '#FFFFFF'),
    table: PdfTableTheme(headerBgColor: '#16A085', headerTextColor: '#FFFFFF', rowTextColor: '#2C3E50', rowBgColor: '#FFFFFF', rowAltBgColor: '#E8F8F5', borderColor: '#D5D8DC', borderWidth: 0.1),
    summary: PdfSummaryTheme(bgColor: '#E8F8F5', textColor: '#2C3E50', highlightBgColor: '#16A085', highlightTextColor: '#FFFFFF', borderColor: '#D5D8DC'),
    footer: PdfFooterTheme(textColor: '#95A5A6', lineColor: '#D5D8DC'),
  );

  /// Theme dark - โทนมืด
  static const PdfTheme darkTheme = PdfTheme(
    name: 'dark',
    colors: PdfThemeColors(primary: '#2C3E50', secondary: '#34495E', accent: '#1ABC9C', danger: '#E74C3C', warning: '#F39C12'),
    header: PdfHeaderTheme(titleColor: '#2C3E50', subtitleColor: '#7F8C8D', lineColor: '#2C3E50', lineWidth: 0.5),
    section: PdfSectionTheme(labelColor: '#2C3E50', textColor: '#1C2833', backgroundColor: '#FFFFFF'),
    table: PdfTableTheme(headerBgColor: '#2C3E50', headerTextColor: '#FFFFFF', rowTextColor: '#1C2833', rowBgColor: '#FFFFFF', rowAltBgColor: '#EBEDEF', borderColor: '#ABB2B9', borderWidth: 0.1),
    summary: PdfSummaryTheme(bgColor: '#EBEDEF', textColor: '#1C2833', highlightBgColor: '#2C3E50', highlightTextColor: '#FFFFFF', borderColor: '#ABB2B9'),
    footer: PdfFooterTheme(textColor: '#85929E', lineColor: '#ABB2B9'),
  );

  /// Default theme alias (ใช้ modern-blue เป็นค่าเริ่มต้น)
  static const PdfTheme defaultTheme = modernBlueTheme;

  /// รายการ theme ทั้งหมด
  static List<PdfTheme> get all => [modernBlueTheme, professionalTheme, natureTheme, warmTheme, classicRedTheme, purpleTheme, tealTheme, darkTheme];

  /// ดึง theme ตามชื่อ
  static PdfTheme getByName(String name) {
    return all.firstWhere((theme) => theme.name == name, orElse: () => defaultTheme);
  }

  /// ชื่อแสดงผลของ theme (ภาษาไทย)
  static String getDisplayName(String name) {
    switch (name) {
      case 'modern-blue':
        return global.language("theme_modern_blue");
      case 'professional':
        return global.language("theme_professional");
      case 'nature':
        return global.language("theme_nature");
      case 'warm':
        return global.language("theme_warm");
      case 'classic-red':
        return global.language("theme_classic_red");
      case 'purple':
        return global.language("theme_purple");
      case 'teal':
        return global.language("theme_teal");
      case 'dark':
        return global.language("theme_dark");
      default:
        return name;
    }
  }

  /// สีหลักของ theme สำหรับแสดงผล preview
  static Color getPreviewColor(String name) {
    switch (name) {
      case 'modern-blue':
        return const Color(0xFF2980B9);
      case 'professional':
        return const Color(0xFF34495E);
      case 'nature':
        return const Color(0xFF27AE60);
      case 'warm':
        return const Color(0xFFE67E22);
      case 'classic-red':
        return const Color(0xFFC0392B);
      case 'purple':
        return const Color(0xFF8E44AD);
      case 'teal':
        return const Color(0xFF16A085);
      case 'dark':
        return const Color(0xFF2C3E50);
      default:
        return const Color(0xFF2980B9);
    }
  }
}

/// ==================== PDF Service ====================

class PdfService {
  /// สร้าง headers สำหรับ API (รวม Authorization token)
  Map<String, String> get _headers {
    final headers = <String, String>{'Content-Type': 'application/json'};
    final token = global.appConfig.getString("token") ?? '';
    if (token.isNotEmpty) {
      headers['Authorization'] = 'Bearer $token';
    }
    return headers;
  }

  /// สร้าง base URL สำหรับ PDF API (ใช้ goApiUrl ถ้ามี)
  String get _baseUrl => global.goApiBaseUrl;

  /// Generate PDF from document
  /// Returns the PDF bytes
  /// [theme] - ธีมสำหรับ PDF หากไม่ระบุจะใช้ธีมมาตรฐาน
  /// [fontSizes] - ขนาดตัวอักษรแต่ละส่วน หากไม่ระบุจะใช้ค่าเริ่มต้น
  /// [template] - รูปแบบเอกสาร หากไม่ระบุจะใช้แบบมาตรฐาน
  /// [dateSettings] - รูปแบบวันที่และปี หากไม่ระบุจะใช้ค่าเริ่มต้น
  /// [languageSettings] - ภาษาสำหรับเอกสาร หากไม่ระบุจะใช้ภาษาไทย
  /// [fontFamily] - ชื่อ font family สำหรับ PDF เช่น Sarabun, Kanit, Prompt
  Future<Uint8List?> generatePdf({
    required String collection,
    required String docNo,
    String? title,
    String pageSize = 'A4',
    String orientation = 'P',
    int fontSize = 9,
    bool colorMode = true,
    String fontFamily = 'GoNotoCurrent',
    bool showDualCurrency = true,
    PdfTheme? theme,
    PdfFontSizes? fontSizes,
    PdfTemplate? template,
    PdfDateSettings? dateSettings,
    PdfLanguageSettings? languageSettings,
  }) async {
    try {
      final baseUrl = _baseUrl;
      final url = Uri.parse('$baseUrl/genpdf');

      // ใช้ค่าเริ่มต้นถ้าไม่ได้ระบุ
      final selectedTheme = theme ?? PdfPredefinedThemes.defaultTheme;
      final selectedTemplate = template ?? PdfPredefinedTemplates.defaultTemplate;
      final selectedDateSettings = dateSettings ?? PdfDateSettings.defaultSettings;
      final selectedLanguageSettings = languageSettings ?? PdfLanguageSettings.defaultSettings;

      AppLogger.info('Generating PDF: $url');
      AppLogger.debug(
        'Params: shopId=${global.getShopId()}, collection=$collection, docNo=$docNo, theme=${selectedTheme.name}, template=${selectedTemplate.id}, lang=${selectedLanguageSettings.toLanguageString()}',
      );
      AppLogger.debug(
        'Font settings: fontFamily=$fontFamily, fontSize=$fontSize, fontSizes=${fontSizes?.toJson()}',
      );

      // สร้าง payload ตาม GenPDF API Specification
      final selectedFontSizes = fontSizes ?? PdfFontSizes.defaultSizes;
      final payload = {
        // Required fields
        'shopid': global.getShopId().trim(),
        'collection': collection.trim(),
        'docno': docNo.trim(),
        // Optional fields
        'title': (title ?? global.language("document")).trim(),
        'pagesize': pageSize.trim().toUpperCase(), // A4, A3, A5, LETTER, LEGAL
        'orientation': orientation.trim().toUpperCase(), // P = Portrait, L = Landscape
        'fontsize': fontSize,
        'fontFamily': fontFamily.trim(), // Font family สำหรับ PDF
        'fontsizes': selectedFontSizes.toJson(), // ขนาด font แยกตามส่วน (header, detail, summary, footer)
        'lineSpacing': selectedFontSizes.lineSpacing, // ระยะห่างระหว่างบรรทัด (1.0, 1.5, 2.0)
        'colormode': colorMode,
        'dateFormat': selectedDateSettings.toFormatString(), // เช่น "DD/MM/BBBB", "DD MMMM YYYY"
        'language': selectedLanguageSettings.toLanguageString(), // "th", "en", "ai:cn"
        'themeName': selectedTheme.name.trim(),
        'templateId': selectedTemplate.id.trim(),
        'showDualCurrency': showDualCurrency,
      };

      final response = await http.post(url, headers: _headers, body: jsonEncode(payload));

      if (response.statusCode == 200) {
        return response.bodyBytes;
      } else {
        try {
          final error = jsonDecode(response.body);
          throw Exception(error['message'] ?? 'Failed to generate PDF');
        } catch (e) {
          throw Exception('Failed to generate PDF: ${response.statusCode} ${response.body}');
        }
      }
    } catch (e) {
      AppLogger.error('Error generating PDF: $e');
      rethrow;
    }
  }

  /// Generate and show PDF in a dialog with preview and print button
  /// [themeName] - ชื่อธีมเริ่มต้น (default, classic, modern, minimal, professional)
  Future<void> generateAndOpenPdf({
    required BuildContext context,
    required String collection,
    required String docNo,
    required String title,
    String pageSize = 'A4',
    String orientation = 'P',
    int fontSize = 9,
    bool colorMode = true,
    String themeName = 'default',
    bool isMultiCurrency = false,
  }) async {
    if (!context.mounted) return;

    // Navigate to PdfViewerDialog with settings
    await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (context) => _PdfViewerDialog(
          collection: collection,
          docNo: docNo,
          title: title,
          initialPageSize: pageSize,
          initialOrientation: orientation,
          initialFontSize: fontSize,
          initialColorMode: colorMode,
          initialThemeName: themeName,
          isMultiCurrency: isMultiCurrency,
        ),
        fullscreenDialog: true,
      ),
    );
  }

  /// Get PdfPageFormat based on pageSize and orientation
  static PdfPageFormat getPageFormat(String pageSize, String orientation) {
    PdfPageFormat format;
    switch (pageSize.toUpperCase()) {
      case 'A3':
        format = PdfPageFormat.a3;
        break;
      case 'A5':
        format = PdfPageFormat.a5;
        break;
      case 'LETTER':
        format = PdfPageFormat.letter;
        break;
      case 'LEGAL':
        format = PdfPageFormat.legal;
        break;
      case 'A4':
      default:
        format = PdfPageFormat.a4;
        break;
    }

    // Apply orientation
    if (orientation.toUpperCase() == 'L') {
      format = format.landscape;
    }

    return format;
  }

  /// Generate and print PDF directly (using Printing package)
  /// [theme] - ธีมสำหรับ PDF
  /// [fontSizes] - ขนาดตัวอักษรแต่ละส่วน
  /// [template] - รูปแบบเอกสาร
  Future<void> generateAndPrintPdf({
    required String collection,
    required String docNo,
    required String title,
    String pageSize = 'A4',
    String orientation = 'P',
    int fontSize = 9,
    bool colorMode = true,
    PdfTheme? theme,
    PdfFontSizes? fontSizes,
    PdfTemplate? template,
  }) async {
    final bytes = await generatePdf(
      collection: collection,
      docNo: docNo,
      title: title,
      pageSize: pageSize,
      orientation: orientation,
      fontSize: fontSize,
      colorMode: colorMode,
      theme: theme,
      fontSizes: fontSizes,
      template: template,
    );

    if (bytes != null) {
      PdfPageFormat format = getPageFormat(pageSize, orientation);
      await Printing.layoutPdf(onLayout: (PdfPageFormat f) async => bytes, format: format, name: '$docNo.pdf');
    }
  }

  /// ดึงประวัติการพิมพ์ของเอกสาร
  /// GET /genpdf/history?shopid=xxx&collection=xxx&docno=xxx
  /// [collection] - ชื่อ collection เช่น transactionPurchaseOrder (optional)
  /// [docNo] - เลขที่เอกสาร (optional)
  Future<List<PdfPrintHistoryItem>> getPrintHistory({String? collection, String? docNo}) async {
    try {
      final baseUrl = _baseUrl;
      final shopId = global.getShopId().trim();

      // Build query parameters
      final queryParams = <String, String>{'shopid': shopId};
      if (collection != null && collection.isNotEmpty) {
        queryParams['collection'] = collection;
      }
      if (docNo != null && docNo.isNotEmpty) {
        queryParams['docno'] = docNo;
      }

      final uri = Uri.parse('$baseUrl/genpdf/history').replace(queryParameters: queryParams);

      AppLogger.info('📜 Fetching print history (GET): $uri');

      final response = await http.get(uri, headers: _headers);

      if (response.statusCode == 200) {
        final List<dynamic> data = jsonDecode(response.body);
        AppLogger.debug('📜 Print history response: ${data.length} items');

        final historyList = data.map((item) => PdfPrintHistoryItem.fromJson(item)).toList();
        return historyList;
      } else if (response.statusCode == 405) {
        AppLogger.warning('⚠️ Print history API not implemented (405)');
        throw PdfHistoryApiNotImplementedException();
      } else {
        AppLogger.warning('⚠️ Failed to fetch print history: ${response.statusCode}');
        return [];
      }
    } catch (e) {
      if (e is PdfHistoryApiNotImplementedException) rethrow;
      AppLogger.error('❌ Error fetching print history: $e');
      return [];
    }
  }

  /// พิมพ์ซ้ำจากประวัติ
  /// [historyId] - ID ของประวัติการพิมพ์
  Future<Uint8List?> reprintFromHistory({required String historyId}) async {
    try {
      final baseUrl = _baseUrl;
      final url = Uri.parse('$baseUrl/genpdf/reprint/$historyId');

      AppLogger.info('🔄 Reprinting from history: $url');

      final response = await http.get(url, headers: _headers);

      if (response.statusCode == 200) {
        AppLogger.success('✅ Reprint successful');
        return response.bodyBytes;
      } else {
        AppLogger.error('❌ Failed to reprint: ${response.statusCode}');
        return null;
      }
    } catch (e) {
      AppLogger.error('❌ Error reprinting: $e');
      return null;
    }
  }

  /// นับจำนวนประวัติการพิมพ์ของเอกสาร (สำหรับแสดง badge บน icon)
  /// [collection] - ชื่อ collection เช่น transactionPurchaseOrder
  /// [docNo] - เลขที่เอกสาร
  /// Returns: จำนวนครั้งที่พิมพ์ หรือ 0 ถ้าไม่มีหรือ error
  Future<int> getPrintHistoryCount({required String collection, required String docNo}) async {
    try {
      final history = await getPrintHistory(collection: collection, docNo: docNo);
      return history.length;
    } catch (e) {
      // ถ้า API ยังไม่ implement หรือ error ให้ return 0
      return 0;
    }
  }

  /// แสดง dialog ประวัติการพิมพ์
  Future<void> showPrintHistoryDialog({required BuildContext context, required String collection, required String docNo, required String title}) async {
    if (!context.mounted) return;

    await showDialog(
      context: context,
      builder: (context) => _PrintHistoryDialog(pdfService: this, collection: collection, docNo: docNo, title: title),
    );
  }
}

/// Exception เมื่อ API ประวัติการพิมพ์ยังไม่ถูก implement
class PdfHistoryApiNotImplementedException implements Exception {
  final String message;
  PdfHistoryApiNotImplementedException([this.message = 'Print history API is not implemented yet']);
  @override
  String toString() => message;
}

/// Model สำหรับประวัติการพิมพ์ (ตาม API Response ใหม่)
class PdfPrintHistoryItem {
  // ข้อมูลหลัก
  final String id;
  final String collection;
  final String docNo;
  final DateTime? docDate;
  final String title;

  // ข้อมูลคู่ค้า
  final String? customerName; // ชื่อลูกหนี้/ลูกค้า (AR)
  final String? vendorName; // ชื่อเจ้าหนี้/ผู้ขาย (AP)

  // ข้อมูลยอดเงิน
  final double totalAmount;
  final String totalAmountText; // format แล้ว มี comma

  // ข้อมูล PDF
  final String theme;
  final String template;
  final String pageSize;
  final String orientation;
  final String language;
  final String filename;
  final int filesize;
  final String filesizeText;
  final String? pdfUrl; // Presigned URL (หมดอายุ 60 นาที)

  // ข้อมูลการพิมพ์
  final DateTime printedAt;
  final String printedBy;
  final int reprintCount;

  PdfPrintHistoryItem({
    required this.id,
    required this.collection,
    required this.docNo,
    this.docDate,
    required this.title,
    this.customerName,
    this.vendorName,
    required this.totalAmount,
    required this.totalAmountText,
    required this.theme,
    required this.template,
    required this.pageSize,
    required this.orientation,
    required this.language,
    required this.filename,
    required this.filesize,
    required this.filesizeText,
    this.pdfUrl,
    required this.printedAt,
    required this.printedBy,
    required this.reprintCount,
  });

  /// ชื่อลูกหนี้/เจ้าหนี้ (ใช้ตัวไหนก็ได้ที่มีค่า)
  String get contactName => customerName ?? vendorName ?? '';

  /// timestamp สำหรับ backward compatibility
  DateTime get timestamp => printedAt;

  /// ตรวจสอบว่า totalAmountText เป็น format string ที่ถูกต้องหรือไม่
  /// (ไม่ใช่ Go format specifier เช่น %!,(float64=xxx).2f)
  bool get isValidAmountText => totalAmountText.isNotEmpty && !totalAmountText.contains('%!') && !totalAmountText.contains('float64');

  /// ยอดรวมที่แสดงผล (ใช้ totalAmountText ถ้าถูกต้อง หรือ format เอง)
  String get displayTotalAmount => isValidAmountText ? totalAmountText : global.formatNumber(totalAmount);

  /// ชื่อเอกสารที่แสดงผล (แปลงจาก collection ถ้า title ว่าง)
  String get displayTitle {
    if (title.isNotEmpty) return title;

    // Mapping ชื่อ collection ภาษาอังกฤษเป็นชื่อภาษาไทย
    final collectionNames = {
      'purchase_order': global.language("purchase_order"),
      'purchaseorder': global.language("purchase_order"),
      'purchase_request': global.language("purchase_request"),
      'purchaserequest': global.language("purchase_request"),
      'sale_order': global.language("sale_order"),
      'saleorder': global.language("sale_order"),
      'sale_invoice': global.language("sale_invoice"),
      'saleinvoice': global.language("sale_invoice"),
      'quotation': global.language("quotation"),
      'receipt': global.language("receipt"),
      'credit_note': global.language("credit_note"),
      'creditnote': global.language("credit_note"),
      'debit_note': global.language("debit_note"),
      'debitnote': global.language("debit_note"),
      'goods_receive': global.language("goods_receive"),
      'goodsreceive': global.language("goods_receive"),
      'goods_issue': global.language("goods_issue"),
      'goodsissue': global.language("goods_issue"),
      'stock_transfer': global.language("stock_transfer"),
      'stocktransfer': global.language("stock_transfer"),
      'stock_count': global.language("stock_count"),
      'stockcount': global.language("stock_count"),
      'payment': global.language("payment_doc"),
      'receive': global.language("receive_doc"),
    };

    // ลองหาจาก mapping ก่อน
    final lowerCollection = collection.toLowerCase();
    if (collectionNames.containsKey(lowerCollection)) {
      return collectionNames[lowerCollection]!;
    }

    // ถ้าไม่มีใน mapping ให้ลองใช้ global.language
    final translated = global.language(collection);
    // ถ้า translation เหมือน key เดิม แสดงว่าไม่มี translation
    if (translated != collection) {
      return translated;
    }

    // fallback: แปลง snake_case เป็น Title Case
    return collection.split('_').map((word) => word.isNotEmpty ? '${word[0].toUpperCase()}${word.substring(1)}' : '').join(' ');
  }

  factory PdfPrintHistoryItem.fromJson(Map<String, dynamic> json) {
    return PdfPrintHistoryItem(
      // ข้อมูลหลัก
      id: json['id'] ?? json['_id'] ?? '',
      collection: json['collection'] ?? '',
      docNo: json['docno'] ?? json['docNo'] ?? '',
      docDate: DateTime.tryParse(json['docdate'] ?? json['docDate'] ?? ''),
      title: json['title'] ?? '',

      // ข้อมูลคู่ค้า
      customerName: json['customername'] ?? json['customerName'],
      vendorName: json['vendorname'] ?? json['vendorName'],

      // ข้อมูลยอดเงิน
      totalAmount: (json['totalamount'] ?? json['totalAmount'] ?? 0).toDouble(),
      totalAmountText: json['totalamountText'] ?? json['totalAmountText'] ?? '',

      // ข้อมูล PDF
      theme: json['theme'] ?? 'default',
      template: json['template'] ?? 'standard',
      pageSize: json['pageSize'] ?? json['pagesize'] ?? 'A4',
      orientation: json['orientation'] ?? 'P',
      language: json['language'] ?? 'th',
      filename: json['filename'] ?? '',
      filesize: json['filesize'] ?? 0,
      filesizeText: json['filesizeText'] ?? '',

      // URL
      pdfUrl: json['pdfurl'] ?? json['pdfUrl'],

      // ข้อมูลการพิมพ์
      printedAt: DateTime.tryParse(json['printedAt'] ?? json['timestamp'] ?? json['createdAt'] ?? '') ?? DateTime.now(),
      printedBy: json['printedBy'] ?? '',
      reprintCount: json['reprintCount'] ?? 0,
    );
  }
}

/// Dialog แสดงประวัติการพิมพ์
class _PrintHistoryDialog extends StatefulWidget {
  final PdfService pdfService;
  final String collection;
  final String docNo;
  final String title;

  const _PrintHistoryDialog({required this.pdfService, required this.collection, required this.docNo, required this.title});

  @override
  State<_PrintHistoryDialog> createState() => _PrintHistoryDialogState();
}

class _PrintHistoryDialogState extends State<_PrintHistoryDialog> with global.ThemeRefreshMixin {
  List<PdfPrintHistoryItem> _historyItems = [];
  bool _isLoading = true;
  String? _errorMessage;
  bool _apiNotImplemented = false;

  @override
  void initState() {
    super.initState();
    _loadHistory();
  }

  Future<void> _loadHistory() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
      _apiNotImplemented = false;
    });

    try {
      final items = await widget.pdfService.getPrintHistory(collection: widget.collection, docNo: widget.docNo);
      if (mounted) {
        setState(() {
          _historyItems = items;
          _isLoading = false;
        });
      }
    } on PdfHistoryApiNotImplementedException {
      if (mounted) {
        setState(() {
          _apiNotImplemented = true;
          _isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _errorMessage = e.toString();
          _isLoading = false;
        });
      }
    }
  }

  /// ดู PDF preview จากประวัติ
  Future<void> _viewItem(PdfPrintHistoryItem item) async {
    // แสดง loading
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => const Center(child: CircularProgressIndicator()),
    );

    Uint8List? bytes;

    // ถ้ามี pdfUrl ให้ดาวน์โหลดจาก URL โดยตรง (เร็วกว่า)
    if (item.pdfUrl != null && item.pdfUrl!.isNotEmpty) {
      try {
        final response = await http.get(Uri.parse(item.pdfUrl!));
        if (response.statusCode == 200) {
          bytes = response.bodyBytes;
        }
      } catch (e) {
        AppLogger.warning('Failed to fetch PDF from URL, falling back to reprint: $e');
      }
    }

    // ถ้าไม่มี pdfUrl หรือดาวน์โหลดไม่สำเร็จ ให้ใช้ reprint API
    bytes ??= await widget.pdfService.reprintFromHistory(historyId: item.id);

    if (mounted) {
      Navigator.of(context).pop(); // ปิด loading
    }

    if (bytes != null && mounted) {
      // สร้าง title ที่แสดงข้อมูลครบ
      final titleText = '${item.displayTitle} - ${item.docNo}';
      final filename = item.filename.isNotEmpty ? item.filename : '${item.docNo}.pdf';

      // แสดง PDF preview dialog
      showDialog(
        context: context,
        builder: (context) => AlertDialog(
          title: Row(
            children: [
              const Icon(Icons.picture_as_pdf, color: Colors.red),
              const SizedBox(width: 8),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(titleText, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.bold)),
                    Text(
                      global.dateTimeBuddhist(item.printedAt, format: global.DateTimeFormatEnum.dateTime),
                      style: TextStyle(fontSize: 11, color: global.theme.iconSecondaryColor, fontWeight: FontWeight.normal),
                    ),
                  ],
                ),
              ),
            ],
          ),
          content: SizedBox(
            width: MediaQuery.of(context).size.width * 0.8,
            height: MediaQuery.of(context).size.height * 0.7,
            child: PdfPreview(
              build: (format) async => bytes!,
              initialPageFormat: PdfService.getPageFormat(item.pageSize, item.orientation),
              canChangeOrientation: false,
              canChangePageFormat: false,
              canDebug: false,
              allowPrinting: true,
              allowSharing: true,
              pdfFileName: filename,
            ),
          ),
          actions: [TextButton(onPressed: () => Navigator.of(context).pop(), child: Text(global.language('close')))],
        ),
      );
    } else if (mounted) {
      global.showErrorSnackBar(context, global.language('view_failed'));
    }
  }

  /// พิมพ์ PDF จากประวัติ
  Future<void> _reprintItem(PdfPrintHistoryItem item) async {
    // แสดง loading
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => const Center(child: CircularProgressIndicator()),
    );

    Uint8List? bytes;

    // ถ้ามี pdfUrl ให้ดาวน์โหลดจาก URL โดยตรง (เร็วกว่า)
    if (item.pdfUrl != null && item.pdfUrl!.isNotEmpty) {
      try {
        final response = await http.get(Uri.parse(item.pdfUrl!));
        if (response.statusCode == 200) {
          bytes = response.bodyBytes;
        }
      } catch (e) {
        AppLogger.warning('Failed to fetch PDF from URL, falling back to reprint: $e');
      }
    }

    // ถ้าไม่มี pdfUrl หรือดาวน์โหลดไม่สำเร็จ ให้ใช้ reprint API
    bytes ??= await widget.pdfService.reprintFromHistory(historyId: item.id);

    if (mounted) {
      Navigator.of(context).pop(); // ปิด loading
    }

    if (bytes != null && mounted) {
      final filename = item.filename.isNotEmpty ? item.filename : '${item.docNo}_reprint.pdf';
      // แสดง print dialog
      await Printing.layoutPdf(onLayout: (PdfPageFormat f) async => bytes!, format: PdfService.getPageFormat(item.pageSize, item.orientation), name: filename);
    } else if (mounted) {
      global.showErrorSnackBar(context, global.language('reprint_failed'));
    }
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Row(
        children: [
          const Icon(Icons.history, color: Colors.blue),
          SizedBox(width: 8),
          Expanded(child: Text('${global.language("print_history")} - ${widget.docNo}', style: TextStyle(fontSize: 16))),
        ],
      ),
      content: SizedBox(width: MediaQuery.of(context).size.width * 0.6, height: MediaQuery.of(context).size.height * 0.5, child: _buildContent()),
      actions: [
        TextButton(onPressed: () => Navigator.of(context).pop(), child: Text(global.language('close'))),
        ElevatedButton.icon(onPressed: _loadHistory, icon: Icon(Icons.refresh, size: 18), label: Text(global.language('refresh'))),
      ],
    );
  }

  Widget _buildContent() {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    // แสดงข้อความเมื่อ API ยังไม่ถูก implement
    if (_apiNotImplemented) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.construction, size: 48, color: Colors.orange[400]),
            SizedBox(height: 16),
            Text(
              global.language('feature_not_available'),
              style: TextStyle(color: Colors.orange[700], fontSize: 16, fontWeight: FontWeight.bold),
            ),
            SizedBox(height: 8),
            Padding(
              padding: EdgeInsets.symmetric(horizontal: 24),
              child: Text(
                global.language('print_history_coming_soon'),
                style: TextStyle(color: global.theme.iconSecondaryColor, fontSize: 14),
                textAlign: TextAlign.center,
              ),
            ),
          ],
        ),
      );
    }

    if (_errorMessage != null) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.error_outline, size: 48, color: Colors.red[300]),
            const SizedBox(height: 16),
            Text(_errorMessage!, style: TextStyle(color: Colors.red[700])),
            SizedBox(height: 16),
            ElevatedButton(onPressed: _loadHistory, child: Text(global.language('retry'))),
          ],
        ),
      );
    }

    if (_historyItems.isEmpty) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.print_disabled, size: 48, color: global.theme.iconSecondaryColor),
            SizedBox(height: 16),
            Text(global.language('no_print_history'), style: TextStyle(color: global.theme.iconSecondaryColor, fontSize: 16)),
          ],
        ),
      );
    }

    return ListView.separated(
      itemCount: _historyItems.length,
      separatorBuilder: (context, index) => const Divider(height: 1),
      itemBuilder: (context, index) {
        final item = _historyItems[index];
        return InkWell(
          onTap: () => _viewItem(item), // กดที่ row เพื่อดู preview
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // ลำดับ
                CircleAvatar(
                  radius: 18,
                  backgroundColor: Colors.blue[100],
                  child: Text(
                    '${index + 1}',
                    style: TextStyle(color: Colors.blue[800], fontWeight: FontWeight.bold, fontSize: 14),
                  ),
                ),
                const SizedBox(width: 12),
                // ข้อมูลหลัก
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // บรรทัดที่ 1: ชื่อเอกสาร (แปลแล้ว)
                      Text(
                        item.displayTitle,
                        style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13),
                        overflow: TextOverflow.ellipsis,
                      ),
                      const SizedBox(height: 2),
                      // บรรทัดที่ 2: เลขที่เอกสาร + วันเวลาพิมพ์
                      Row(
                        children: [
                          Icon(Icons.description, size: 12, color: Colors.blue[600]),
                          const SizedBox(width: 4),
                          Text(
                            item.docNo.isNotEmpty ? item.docNo : widget.docNo,
                            style: TextStyle(fontSize: 12, color: Colors.blue[700], fontWeight: FontWeight.w500),
                          ),
                          const SizedBox(width: 12),
                          Icon(Icons.access_time, size: 12, color: global.theme.iconSecondaryColor),
                          const SizedBox(width: 4),
                          Text(
                            global.dateTimeBuddhist(item.printedAt, format: global.DateTimeFormatEnum.dateTime),
                            style: TextStyle(fontSize: 11, color: global.theme.iconSecondaryColor),
                          ),
                        ],
                      ),
                      const SizedBox(height: 4),
                      // บรรทัดที่ 3: ข้อมูลคู่ค้า + ยอดเงิน (แสดงยอดเงินเสมอถ้ามีค่า)
                      Row(
                        children: [
                          if (item.contactName.isNotEmpty) ...[
                            Icon(Icons.person_outline, size: 14, color: global.theme.iconSecondaryColor),
                            const SizedBox(width: 4),
                            Flexible(
                              child: Text(
                                item.contactName,
                                style: TextStyle(fontSize: 12, color: global.theme.textSecondaryColor),
                                overflow: TextOverflow.ellipsis,
                              ),
                            ),
                            const SizedBox(width: 12),
                          ],
                          if (item.totalAmount > 0) ...[
                            Icon(Icons.payments_outlined, size: 14, color: Colors.green[600]),
                            const SizedBox(width: 4),
                            Text(
                              item.displayTotalAmount,
                              style: TextStyle(fontSize: 12, color: Colors.green[700], fontWeight: FontWeight.w500),
                            ),
                          ],
                        ],
                      ),
                      const SizedBox(height: 4),
                      // วันที่เอกสาร (ถ้ามี)
                      if (item.docDate != null) ...[
                        Row(
                          children: [
                            Icon(Icons.calendar_today, size: 12, color: global.theme.iconSecondaryColor),
                            SizedBox(width: 4),
                            Text(
                              '${global.language("doc_date")}: ${global.dateTimeBuddhist(item.docDate!, format: global.DateTimeFormatEnum.date)}',
                              style: TextStyle(fontSize: 11, color: global.theme.iconSecondaryColor),
                            ),
                          ],
                        ),
                        const SizedBox(height: 4),
                      ],
                      // Chips (theme, pageSize, filesize, reprint)
                      Wrap(
                        spacing: 4,
                        runSpacing: 4,
                        children: [
                          _buildChip(item.theme, Icons.palette),
                          _buildChip('${item.pageSize} ${item.orientation == 'L' ? 'L' : 'P'}', Icons.description),
                          if (item.filesizeText.isNotEmpty) _buildChip(item.filesizeText, Icons.insert_drive_file),
                          if (item.printedBy.isNotEmpty) _buildChip(item.printedBy, Icons.person),
                          if (item.reprintCount > 0) _buildChip('Reprint: ${item.reprintCount}', Icons.replay),
                        ],
                      ),
                    ],
                  ),
                ),
                // ปุ่มพิมพ์
                IconButton(
                  icon: Icon(Icons.print, color: Colors.blue),
                  tooltip: global.language('reprint'),
                  onPressed: () => _reprintItem(item),
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildChip(String label, IconData icon) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      decoration: BoxDecoration(color: global.theme.dividerBorderColor, borderRadius: BorderRadius.circular(4)),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 12, color: global.theme.iconSecondaryColor),
          const SizedBox(width: 2),
          Text(label, style: TextStyle(fontSize: 10, color: global.theme.textSecondaryColor)),
        ],
      ),
    );
  }
}

/// Widget แสดง icon ประวัติการพิมพ์พร้อม badge จำนวนครั้งที่พิมพ์
/// ใช้สำหรับแสดงในหน้า transaction โดยจะดึงข้อมูลจำนวนครั้งที่พิมพ์จาก API
class PrintHistoryIconButton extends StatefulWidget {
  final String collection;
  final String docNo;
  final String title;
  final double iconSize;

  const PrintHistoryIconButton({super.key, required this.collection, required this.docNo, required this.title, this.iconSize = 26.0});

  @override
  State<PrintHistoryIconButton> createState() => _PrintHistoryIconButtonState();
}

class _PrintHistoryIconButtonState extends State<PrintHistoryIconButton> with global.ThemeRefreshMixin {
  int _printCount = 0;
  bool _isLoading = true;
  final PdfService _pdfService = PdfService();

  @override
  void initState() {
    super.initState();
    _loadPrintCount();
  }

  @override
  void didUpdateWidget(PrintHistoryIconButton oldWidget) {
    super.didUpdateWidget(oldWidget);
    // โหลดใหม่ถ้า docNo เปลี่ยน
    if (oldWidget.docNo != widget.docNo || oldWidget.collection != widget.collection) {
      _loadPrintCount();
    }
  }

  /// โหลดจำนวนครั้งที่พิมพ์จาก API
  Future<void> _loadPrintCount() async {
    if (widget.docNo.isEmpty) {
      setState(() {
        _printCount = 0;
        _isLoading = false;
      });
      return;
    }

    setState(() => _isLoading = true);

    final count = await _pdfService.getPrintHistoryCount(collection: widget.collection, docNo: widget.docNo);

    if (mounted) {
      setState(() {
        _printCount = count;
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return IconButton(
      focusNode: FocusNode(skipTraversal: true),
      onPressed: () async {
        await _pdfService.showPrintHistoryDialog(context: context, collection: widget.collection, docNo: widget.docNo, title: widget.title);
        // โหลดใหม่หลังปิด dialog (เผื่อมีการพิมพ์เพิ่ม)
        _loadPrintCount();
      },
      icon: Stack(
        clipBehavior: Clip.none,
        children: [
          Icon(Icons.history, size: widget.iconSize),
          // แสดง badge ถ้ามีจำนวน > 0 และไม่ได้กำลังโหลด
          if (!_isLoading && _printCount > 0)
            Positioned(
              right: -6,
              top: -6,
              child: Container(
                padding: const EdgeInsets.all(4),
                decoration: BoxDecoration(
                  color: Colors.red,
                  shape: BoxShape.circle,
                  border: Border.all(color: global.theme.cardColor, width: 1.5),
                ),
                constraints: const BoxConstraints(minWidth: 18, minHeight: 18),
                child: Text(
                  _printCount > 99 ? '99+' : _printCount.toString(),
                  style: TextStyle(color: global.theme.onPrimaryColor, fontSize: 10, fontWeight: FontWeight.bold),
                  textAlign: TextAlign.center,
                ),
              ),
            ),
        ],
      ),
      tooltip: global.language('print_history'),
    );
  }
}

/// PrintHistoryActionButton — แสดง icon พร้อม badge + label ด้านล่าง
class PrintHistoryActionButton extends StatefulWidget {
  final String collection;
  final String docNo;
  final String title;
  final double iconSize;
  final String label;

  const PrintHistoryActionButton({
    super.key,
    required this.collection,
    required this.docNo,
    required this.title,
    this.iconSize = 22.0,
    required this.label,
  });

  @override
  State<PrintHistoryActionButton> createState() => _PrintHistoryActionButtonState();
}

class _PrintHistoryActionButtonState extends State<PrintHistoryActionButton> with global.ThemeRefreshMixin {
  int _printCount = 0;
  bool _isLoading = true;
  final PdfService _pdfService = PdfService();

  @override
  void initState() {
    super.initState();
    _loadPrintCount();
  }

  @override
  void didUpdateWidget(PrintHistoryActionButton oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.docNo != widget.docNo || oldWidget.collection != widget.collection) {
      _loadPrintCount();
    }
  }

  Future<void> _loadPrintCount() async {
    if (widget.docNo.isEmpty) {
      setState(() {
        _printCount = 0;
        _isLoading = false;
      });
      return;
    }

    setState(() => _isLoading = true);

    final count = await _pdfService.getPrintHistoryCount(collection: widget.collection, docNo: widget.docNo);

    if (mounted) {
      setState(() {
        _printCount = count;
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: () async {
        await _pdfService.showPrintHistoryDialog(context: context, collection: widget.collection, docNo: widget.docNo, title: widget.title);
        _loadPrintCount();
      },
      borderRadius: BorderRadius.circular(8),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 6.0, vertical: 4.0),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Stack(
              clipBehavior: Clip.none,
              children: [
                Icon(Icons.print, size: widget.iconSize, color: global.theme.onPrimaryColor),
                if (!_isLoading && _printCount > 0)
                  Positioned(
                    right: -6,
                    top: -6,
                    child: Container(
                      padding: const EdgeInsets.all(3),
                      decoration: BoxDecoration(
                        color: Colors.red,
                        shape: BoxShape.circle,
                        border: Border.all(color: global.theme.cardColor, width: 1.5),
                      ),
                      constraints: const BoxConstraints(minWidth: 16, minHeight: 16),
                      child: Text(
                        _printCount > 99 ? '99+' : _printCount.toString(),
                        style: TextStyle(color: global.theme.onPrimaryColor, fontSize: 8, fontWeight: FontWeight.bold),
                        textAlign: TextAlign.center,
                      ),
                    ),
                  ),
              ],
            ),
            const SizedBox(height: 2),
            Text(
              widget.label,
              style: TextStyle(fontSize: 9, color: global.theme.onPrimaryColor.withValues(alpha: 0.7)),
              overflow: TextOverflow.ellipsis,
            ),
          ],
        ),
      ),
    );
  }
}

/// PDF Viewer Dialog with settings controls
class _PdfViewerDialog extends StatefulWidget {
  final String collection;
  final String docNo;
  final String title;
  final String initialPageSize;
  final String initialOrientation;
  final int initialFontSize;
  final bool initialColorMode;
  final String initialThemeName;
  final PdfFontSizes initialFontSizes;
  final String initialTemplateId;
  final PdfDateSettings initialDateSettings;
  final bool isMultiCurrency;

  const _PdfViewerDialog({
    required this.collection,
    required this.docNo,
    required this.title,
    required this.initialPageSize,
    required this.initialOrientation,
    required this.initialFontSize,
    required this.initialColorMode,
    this.initialThemeName = 'default',
    this.initialFontSizes = const PdfFontSizes(),
    this.initialTemplateId = 'default',
    this.initialDateSettings = const PdfDateSettings(),
    this.isMultiCurrency = false,
  });

  @override
  State<_PdfViewerDialog> createState() => _PdfViewerDialogState();
}

class _PdfViewerDialogState extends State<_PdfViewerDialog> with global.ThemeRefreshMixin {
  final GlobalKey<ScaffoldState> _scaffoldKey = GlobalKey<ScaffoldState>();
  late String _pageSize;
  late String _orientation;
  late int _fontSize;
  late bool _colorMode;
  late String _themeName;
  late PdfFontSizes _fontSizes;
  late String _templateId;
  late String _fontFamily; // Font family สำหรับ PDF (Sarabun, Kanit, Prompt, etc.)
  late PdfDateSettings _dateSettings;
  late PdfLanguageSettings _languageSettings;
  late bool _showDualCurrency;
  Uint8List? _pdfBytes;
  String _pdfKey = '';
  bool _isGenerating = false;

  @override
  void initState() {
    super.initState();
    // Initialize with default values first
    _pageSize = widget.initialPageSize;
    _orientation = widget.initialOrientation;
    _fontSize = widget.initialFontSize;
    _colorMode = widget.initialColorMode;
    _themeName = widget.initialThemeName;
    _fontSizes = widget.initialFontSizes;
    _templateId = widget.initialTemplateId;
    _fontFamily = 'Sarabun'; // Default font family (Google Thai - คลาสสิก อ่านง่าย)
    _dateSettings = widget.initialDateSettings;
    _languageSettings = const PdfLanguageSettings();
    _showDualCurrency = true; // ค่าเริ่มต้น true, _loadUserPreferences จะ override ตาม isMultiCurrency
    // Then load user preferences
    _loadUserPreferences();
  }

  /// Load user preferences from SharedPreferences
  Future<void> _loadUserPreferences() async {
    final prefs = await SharedPreferences.getInstance();
    final userEmail = global.loginEmail.isNotEmpty ? global.loginEmail : 'default';
    final key = 'pdf_settings_$userEmail';

    final savedPageSize = prefs.getString('${key}_pageSize');
    final savedOrientation = prefs.getString('${key}_orientation');
    final savedFontSize = prefs.getInt('${key}_fontSize');
    final savedColorMode = prefs.getBool('${key}_colorMode');
    final savedThemeName = prefs.getString('${key}_themeName');
    final savedTemplateId = prefs.getString('${key}_templateId');
    final savedFontFamily = prefs.getString('${key}_fontFamily');
    // โหลด font sizes แต่ละส่วน
    final savedFontSizeHeader = prefs.getInt('${key}_fontSizeHeader');
    final savedFontSizeDetail = prefs.getInt('${key}_fontSizeDetail');
    final savedFontSizeSummary = prefs.getInt('${key}_fontSizeSummary');
    final savedFontSizeFooter = prefs.getInt('${key}_fontSizeFooter');
    // โหลด date settings
    final savedDateFormat = prefs.getString('${key}_dateFormat');
    final savedYearFormat = prefs.getString('${key}_yearFormat');
    // โหลด language settings
    final savedLanguageCode = prefs.getString('${key}_languageCode');
    // โหลด showDualCurrency
    final savedShowDualCurrency = prefs.getBool('${key}_showDualCurrency');

    // Update values if saved preferences exist
    if (savedPageSize != null) _pageSize = savedPageSize;
    if (savedOrientation != null) _orientation = savedOrientation;
    if (savedFontSize != null) _fontSize = savedFontSize;
    if (savedColorMode != null) _colorMode = savedColorMode;
    if (savedThemeName != null) _themeName = savedThemeName;
    if (savedTemplateId != null) _templateId = savedTemplateId;
    if (savedFontFamily != null) _fontFamily = savedFontFamily;
    // อัพเดท font sizes
    _fontSizes = _fontSizes.copyWith(header: savedFontSizeHeader, detail: savedFontSizeDetail, summary: savedFontSizeSummary, footer: savedFontSizeFooter);
    // อัพเดท date settings
    if (savedDateFormat != null || savedYearFormat != null) {
      _dateSettings = _dateSettings.copyWith(
        dateFormat: savedDateFormat != null ? PdfDateFormat.values.firstWhere((e) => e.name == savedDateFormat, orElse: () => PdfDateFormat.ddmmyyyy) : null,
        yearFormat: savedYearFormat != null ? PdfYearFormat.values.firstWhere((e) => e.name == savedYearFormat, orElse: () => PdfYearFormat.buddhist) : null,
      );
    }
    // อัพเดท language settings
    if (savedLanguageCode != null) {
      _languageSettings = _languageSettings.copyWith(languageCode: savedLanguageCode);
    }
    // อัพเดท showDualCurrency (ถ้าเอกสารไม่ใช่ multi-currency ให้ force เป็น false)
    if (!widget.isMultiCurrency) {
      _showDualCurrency = false;
    } else if (savedShowDualCurrency != null) {
      _showDualCurrency = savedShowDualCurrency;
    }

    // Load PDF after preferences are loaded
    if (mounted) await _loadPdf();
  }

  /// Save user preferences to SharedPreferences
  Future<void> _saveUserPreferences() async {
    final prefs = await SharedPreferences.getInstance();
    final userEmail = global.loginEmail.isNotEmpty ? global.loginEmail : 'default';
    final key = 'pdf_settings_$userEmail';

    await prefs.setString('${key}_pageSize', _pageSize);
    await prefs.setString('${key}_orientation', _orientation);
    await prefs.setInt('${key}_fontSize', _fontSize);
    await prefs.setBool('${key}_colorMode', _colorMode);
    await prefs.setString('${key}_themeName', _themeName);
    await prefs.setString('${key}_templateId', _templateId);
    await prefs.setString('${key}_fontFamily', _fontFamily);
    // บันทึก font sizes แต่ละส่วน
    await prefs.setInt('${key}_fontSizeHeader', _fontSizes.header);
    await prefs.setInt('${key}_fontSizeDetail', _fontSizes.detail);
    await prefs.setInt('${key}_fontSizeSummary', _fontSizes.summary);
    await prefs.setInt('${key}_fontSizeFooter', _fontSizes.footer);
    // บันทึก date settings
    await prefs.setString('${key}_dateFormat', _dateSettings.dateFormat.name);
    await prefs.setString('${key}_yearFormat', _dateSettings.yearFormat.name);
    // บันทึก language settings
    await prefs.setString('${key}_languageCode', _languageSettings.languageCode);
    // บันทึก showDualCurrency
    await prefs.setBool('${key}_showDualCurrency', _showDualCurrency);
  }

  Future<void> _loadPdf() async {
    if (_isGenerating) return;

    setState(() => _isGenerating = true);

    try {
      final pdfService = PdfService();
      // ดึง theme ตามชื่อที่เลือก
      final theme = PdfPredefinedThemes.getByName(_themeName);
      // ดึง template ตาม id ที่เลือก
      final template = PdfPredefinedTemplates.getById(_templateId);
      final bytes = await pdfService.generatePdf(
        collection: widget.collection,
        docNo: widget.docNo,
        title: widget.title,
        pageSize: _pageSize,
        orientation: _orientation,
        fontSize: _fontSize,
        colorMode: _colorMode,
        fontFamily: _fontFamily,
        showDualCurrency: _showDualCurrency,
        theme: theme,
        fontSizes: _fontSizes,
        template: template,
        dateSettings: _dateSettings,
        languageSettings: _languageSettings,
      );
      if (mounted) {
        setState(() {
          _pdfBytes = bytes;
          _pdfKey =
              '${_pageSize}_${_orientation}_${_fontSize}_${_colorMode}_${_themeName}_${_templateId}_${_fontFamily}_${_fontSizes.header}_${_fontSizes.detail}_${_fontSizes.summary}_${_fontSizes.footer}_${_dateSettings.dateFormat.name}_${_dateSettings.yearFormat.name}_${_languageSettings.languageCode}_$_showDualCurrency';
          _isGenerating = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() => _isGenerating = false);
        global.showErrorSnackBar(context, '${global.language("error_occurred")}: $e');
      }
    }
  }

  Future<void> _handlePrint() async {
    if (_pdfBytes == null) return;

    // Create printing overlay
    final printingOverlay = OverlayEntry(
      builder: (context) => Material(
        color: global.theme.dividerBorderColor,
        child: Center(
          child: Container(
            padding: const EdgeInsets.all(20),
            decoration: BoxDecoration(color: global.theme.cardColor, borderRadius: BorderRadius.circular(10)),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                const SizedBox(width: 50, height: 50, child: CircularProgressIndicator()),
                SizedBox(height: 20),
                Text(global.language("printing"), style: TextStyle(fontSize: 16)),
              ],
            ),
          ),
        ),
      ),
    );

    try {
      Overlay.of(context).insert(printingOverlay);

      final format = PdfService.getPageFormat(_pageSize, _orientation);

      // Use current PDF bytes with all settings applied (pageSize, orientation, fontSize, colorMode)
      // The PDF is already generated with these settings, so we just need to print it
      final result = await Printing.layoutPdf(onLayout: (_) => Future.value(_pdfBytes!), format: format, name: '${widget.docNo}.pdf');

      printingOverlay.remove();

      if (result && mounted) {
        Navigator.pop(context);
        global.showSuccessSnackBar(context, global.language("print_complete"));
      }
    } catch (e) {
      printingOverlay.remove();
      if (mounted) {
        global.showErrorSnackBar(context, '${global.language("error_printing")}: $e');
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      key: _scaffoldKey,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        title: Text(widget.title),
        leading: IconButton(icon: const Icon(Icons.close), onPressed: () => Navigator.pop(context)),
      ),
      body: Row(
        children: [
          // Left: PDF Preview
          Expanded(
            flex: 7,
            child: Container(
              color: global.theme.dividerBorderColor,
              child: Stack(
                children: [
                  if (_pdfBytes != null)
                    _PdfViewerBody(pdfBytes: _pdfBytes!, pdfKey: _pdfKey, pageSize: _pageSize, orientation: _orientation)
                  else
                    Center(
                      child: Column(mainAxisSize: MainAxisSize.min, children: [ CircularProgressIndicator(), SizedBox(height: 16), Text(global.language("loading_pdf"))]),
                    ),
                  if (_isGenerating && _pdfBytes != null)
                    Container(
                      color: global.theme.textSecondaryColor,
                      child: Center(
                        child: Container(
                          padding: const EdgeInsets.all(20),
                          decoration: BoxDecoration(color: global.theme.cardColor, borderRadius: BorderRadius.circular(10)),
                          child: Column(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              const SizedBox(width: 50, height: 50, child: CircularProgressIndicator(strokeWidth: 4)),
                              SizedBox(height: 16),
                              Text(global.language("generating_pdf"), style: TextStyle(fontSize: 16)),
                            ],
                          ),
                        ),
                      ),
                    ),
                ],
              ),
            ),
          ),
          // Right: Settings Panel
          Container(
            width: 320,
            decoration: BoxDecoration(
              color: global.theme.cardColor,
              boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.1), blurRadius: 8, offset: const Offset(-2, 0))],
            ),
            child: _SettingsPanel(
              pageSize: _pageSize,
              orientation: _orientation,
              fontSize: _fontSize,
              colorMode: _colorMode,
              themeName: _themeName,
              fontSizes: _fontSizes,
              pdfBytes: _pdfBytes,
              onPrint: _handlePrint,
              onPageSizeChanged: (value) {
                _pageSize = value;
                _saveUserPreferences();
                _loadPdf();
              },
              onOrientationChanged: (value) {
                _orientation = value;
                _saveUserPreferences();
                _loadPdf();
              },
              onFontSizeChanged: (value) {
                setState(() => _fontSize = value);
              },
              onFontSizeChangeEnd: () {
                _saveUserPreferences();
                _loadPdf();
              },
              onColorModeChanged: (value) {
                _colorMode = value;
                _saveUserPreferences();
                _loadPdf();
              },
              onThemeChanged: (value) {
                _themeName = value;
                _saveUserPreferences();
                _loadPdf();
              },
              onFontSizesChanged: (value) {
                setState(() => _fontSizes = value);
              },
              onFontSizesChangeEnd: () {
                _saveUserPreferences();
                _loadPdf();
              },
              templateId: _templateId,
              onTemplateChanged: (value) {
                _templateId = value;
                _saveUserPreferences();
                _loadPdf();
              },
              fontFamily: _fontFamily,
              onFontFamilyChanged: (value) {
                _fontFamily = value;
                _saveUserPreferences();
                _loadPdf();
              },
              dateSettings: _dateSettings,
              onDateSettingsChanged: (value) {
                _dateSettings = value;
                _saveUserPreferences();
                _loadPdf();
              },
              languageSettings: _languageSettings,
              onLanguageSettingsChanged: (value) {
                _languageSettings = value;
                _saveUserPreferences();
                _loadPdf();
              },
              showDualCurrency: _showDualCurrency,
              isMultiCurrency: widget.isMultiCurrency,
              onShowDualCurrencyChanged: (value) {
                _showDualCurrency = value;
                _saveUserPreferences();
                _loadPdf();
              },
            ),
          ),
        ],
      ),
    );
  }
}

/// Settings Panel Widget
class _SettingsPanel extends StatefulWidget {
  final String pageSize;
  final String orientation;
  final int fontSize;
  final bool colorMode;
  final String themeName;
  final PdfFontSizes fontSizes;
  final String templateId;
  final String fontFamily; // Font family สำหรับ PDF
  final PdfDateSettings dateSettings;
  final PdfLanguageSettings languageSettings;
  final bool showDualCurrency;
  final bool isMultiCurrency;
  final Uint8List? pdfBytes;
  final VoidCallback onPrint;
  final Function(String) onPageSizeChanged;
  final Function(String) onOrientationChanged;
  final Function(int) onFontSizeChanged;
  final VoidCallback onFontSizeChangeEnd;
  final Function(bool) onColorModeChanged;
  final Function(String) onThemeChanged;
  final Function(PdfFontSizes) onFontSizesChanged;
  final VoidCallback onFontSizesChangeEnd;
  final Function(String) onTemplateChanged;
  final Function(String) onFontFamilyChanged; // Callback เมื่อเปลี่ยน font family
  final Function(PdfDateSettings) onDateSettingsChanged;
  final Function(PdfLanguageSettings) onLanguageSettingsChanged;
  final Function(bool) onShowDualCurrencyChanged;

  const _SettingsPanel({
    required this.pageSize,
    required this.orientation,
    required this.fontSize,
    required this.colorMode,
    required this.themeName,
    required this.fontSizes,
    required this.templateId,
    required this.fontFamily,
    required this.dateSettings,
    required this.languageSettings,
    required this.showDualCurrency,
    this.isMultiCurrency = false,
    required this.pdfBytes,
    required this.onPrint,
    required this.onPageSizeChanged,
    required this.onOrientationChanged,
    required this.onFontSizeChanged,
    required this.onFontSizeChangeEnd,
    required this.onColorModeChanged,
    required this.onThemeChanged,
    required this.onFontSizesChanged,
    required this.onFontSizesChangeEnd,
    required this.onTemplateChanged,
    required this.onFontFamilyChanged,
    required this.onDateSettingsChanged,
    required this.onLanguageSettingsChanged,
    required this.onShowDualCurrencyChanged,
  });

  @override
  State<_SettingsPanel> createState() => _SettingsPanelState();
}

class _SettingsPanelState extends State<_SettingsPanel> with global.ThemeRefreshMixin {
  late PdfFontSizes _tempFontSizes;

  @override
  void initState() {
    super.initState();
    _tempFontSizes = widget.fontSizes;
  }

  /// ดึงไอคอนสำหรับแต่ละ template
  IconData _getTemplateIcon(String templateId) {
    switch (templateId) {
      case 'standard':
        return Icons.description;
      case 'modern':
        return Icons.auto_awesome;
      case 'professional':
        return Icons.business_center;
      case 'compact':
        return Icons.compress;
      case 'simple':
        return Icons.article;
      case 'classic':
        return Icons.history_edu;
      case 'premium':
        return Icons.diamond;
      case 'minimalist':
        return Icons.remove_circle_outline;
      case 'corporate':
        return Icons.corporate_fare;
      case 'creative':
        return Icons.palette;
      default:
        return Icons.description;
    }
  }

  /// Helper widget สำหรับสร้าง row เลือกขนาดตัวอักษร
  Widget _buildFontSizeRow({required String label, required int value, required int min, required int max, required Function(int) onChanged, required VoidCallback onChangeEnd}) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        children: [
          SizedBox(width: 80, child: Text(label, style: const TextStyle(fontSize: 13))),
          Expanded(
            child: SliderTheme(
              data: SliderTheme.of(context).copyWith(trackHeight: 3, thumbShape: const RoundSliderThumbShape(enabledThumbRadius: 8)),
              child: Slider(value: value.toDouble(), min: min.toDouble(), max: max.toDouble(), divisions: max - min, onChanged: (val) => onChanged(val.toInt()), onChangeEnd: (_) => onChangeEnd()),
            ),
          ),
          SizedBox(
            width: 40,
            child: Text(
              '$value',
              style: const TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              textAlign: TextAlign.center,
            ),
          ),
        ],
      ),
    );
  }

  /// Helper widget สำหรับสร้าง row เลือกระยะห่างระหว่างบรรทัด
  Widget _buildLineSpacingRow({required double value, required Function(double) onChanged, required VoidCallback onChangeEnd}) {
    // ตัวเลือก line spacing: 1.0, 1.25, 1.5, 1.75, 2.0
    final options = [1.0, 1.25, 1.5, 1.75, 2.0];
    final labels = ['1.0', '1.25', '1.5', '1.75', '2.0'];
    return Wrap(
      spacing: 8,
      runSpacing: 8,
      children: List.generate(options.length, (index) {
        final isSelected = (value - options[index]).abs() < 0.01;
        return ChoiceChip(
          label: Text(labels[index], style: const TextStyle(fontSize: 12)),
          selected: isSelected,
          onSelected: (_) {
            onChanged(options[index]);
            onChangeEnd();
          },
          selectedColor: global.theme.appBarColor.withValues(alpha: 0.3),
          backgroundColor: global.theme.dividerBorderColor,
        );
      }),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Container(
          padding: const EdgeInsets.all(16),
          color: global.theme.appBarColor,
          child: Row(
            children: [
              Icon(Icons.tune, color: global.theme.onPrimaryColor, size: 28),
              SizedBox(width: 12),
              Text(
                global.language("pdf_settings"),
                style: TextStyle(color: global.theme.onPrimaryColor, fontSize: 20, fontWeight: FontWeight.bold),
              ),
            ],
          ),
        ),
        Expanded(
          child: ListView(
            padding: const EdgeInsets.all(16),
            children: [
              // ปุ่มพิมพ์เอกสาร (ด้านบน)
              SizedBox(
                width: double.infinity,
                child: ElevatedButton.icon(
                  icon: Icon(Icons.print, size: 20),
                  label: Text(global.language("print_document"), style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold)),
                  style: ElevatedButton.styleFrom(backgroundColor: global.theme.appBarColor, foregroundColor: global.theme.onPrimaryColor, padding: EdgeInsets.symmetric(vertical: 12), elevation: 2),
                  onPressed: widget.pdfBytes != null ? widget.onPrint : null,
                ),
              ),
              const SizedBox(height: 16),

              // Page Size Section
              Text(global.language("paper_size"), style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold)),
              const SizedBox(height: 8),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: ['A3', 'A4', 'A5', 'LETTER', 'LEGAL'].map((size) {
                  final isSelected = widget.pageSize == size;
                  return ChoiceChip(
                    label: Text(size),
                    selected: isSelected,
                    onSelected: (_) => widget.onPageSizeChanged(size),
                    selectedColor: global.theme.appBarColor.withValues(alpha: 0.3),
                    backgroundColor: global.theme.dividerBorderColor,
                  );
                }).toList(),
              ),
              const SizedBox(height: 12),

              // Orientation Section
              Text(global.language("paper_orientation"), style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold)),
              SizedBox(height: 8),
              Wrap(
                spacing: 8,
                children: [
                  ChoiceChip(
                    label: Row(mainAxisSize: MainAxisSize.min, children: [ Icon(Icons.portrait, size: 18), SizedBox(width: 4), Text(global.language("portrait"))]),
                    selected: widget.orientation == 'P',
                    onSelected: (_) => widget.onOrientationChanged('P'),
                    selectedColor: global.theme.appBarColor.withValues(alpha: 0.3),
                    backgroundColor: global.theme.dividerBorderColor,
                  ),
                  ChoiceChip(
                    label: Row(mainAxisSize: MainAxisSize.min, children: [ Icon(Icons.landscape, size: 18), SizedBox(width: 4), Text(global.language("landscape"))]),
                    selected: widget.orientation == 'L',
                    onSelected: (_) => widget.onOrientationChanged('L'),
                    selectedColor: global.theme.appBarColor.withValues(alpha: 0.3),
                    backgroundColor: global.theme.dividerBorderColor,
                  ),
                ],
              ),
              const SizedBox(height: 12),

              // Font Family Section - เลือก font สำหรับ PDF
              Text(global.language("pdf_font_setting"), style: const TextStyle(fontSize: 14, fontWeight: FontWeight.bold)),
              const SizedBox(height: 8),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: global.availablePdfFontFamilies.map((font) {
                  final isSelected = widget.fontFamily == font;
                  // แสดงชื่อ font ให้อ่านง่ายขึ้น
                  final displayNames = {
                    // Google Thai Fonts - ยอดนิยม
                    'Sarabun': 'Sarabun (คลาสสิก)',
                    'Kanit': 'Kanit (ทันสมัย)',
                    'Prompt': 'Prompt (โปร)',
                    'Mitr': 'Mitr (เป็นมิตร)',
                    // Universal & Noto
                    'GoNotoCurrent': 'Universal (ทุกภาษา)',
                    'NotoSansThai': 'Noto Thai',
                  };
                  final displayName = displayNames[font] ?? font;
                  return ChoiceChip(
                    label: Text(displayName, style: const TextStyle(fontSize: 11)),
                    selected: isSelected,
                    onSelected: (_) => widget.onFontFamilyChanged(font),
                    selectedColor: global.theme.appBarColor.withValues(alpha: 0.3),
                    backgroundColor: global.theme.dividerBorderColor,
                  );
                }).toList(),
              ),
              const SizedBox(height: 12),

              // Font Sizes Section - ขนาดตัวอักษรแต่ละส่วน
              Text(global.language("font_size"), style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold)),
              const SizedBox(height: 8),
              // Header font size
              _buildFontSizeRow(
                label: global.language("document_header"),
                value: _tempFontSizes.header,
                min: 8,
                max: 14,
                onChanged: (val) {
                  setState(() => _tempFontSizes = _tempFontSizes.copyWith(header: val));
                  widget.onFontSizesChanged(_tempFontSizes);
                },
                onChangeEnd: widget.onFontSizesChangeEnd,
              ),
              // Detail font size
              _buildFontSizeRow(
                label: global.language("detail"),
                value: _tempFontSizes.detail,
                min: 6,
                max: 12,
                onChanged: (val) {
                  setState(() => _tempFontSizes = _tempFontSizes.copyWith(detail: val));
                  widget.onFontSizesChanged(_tempFontSizes);
                },
                onChangeEnd: widget.onFontSizesChangeEnd,
              ),
              // Summary font size
              _buildFontSizeRow(
                label: global.language("summary"),
                value: _tempFontSizes.summary,
                min: 7,
                max: 12,
                onChanged: (val) {
                  setState(() => _tempFontSizes = _tempFontSizes.copyWith(summary: val));
                  widget.onFontSizesChanged(_tempFontSizes);
                },
                onChangeEnd: widget.onFontSizesChangeEnd,
              ),
              // Footer font size
              _buildFontSizeRow(
                label: global.language("document_footer"),
                value: _tempFontSizes.footer,
                min: 6,
                max: 10,
                onChanged: (val) {
                  setState(() => _tempFontSizes = _tempFontSizes.copyWith(footer: val));
                  widget.onFontSizesChanged(_tempFontSizes);
                },
                onChangeEnd: widget.onFontSizesChangeEnd,
              ),
              const SizedBox(height: 12),

              // Line Spacing Section - ระยะห่างระหว่างบรรทัด
              Text(global.language("line_spacing_detail"), style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold)),
              const SizedBox(height: 8),
              _buildLineSpacingRow(
                value: _tempFontSizes.lineSpacing,
                onChanged: (val) {
                  setState(() => _tempFontSizes = _tempFontSizes.copyWith(lineSpacing: val));
                  widget.onFontSizesChanged(_tempFontSizes);
                },
                onChangeEnd: widget.onFontSizesChangeEnd,
              ),
              const SizedBox(height: 12),

              // Color Mode Section
              Text(global.language("color_mode"), style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold)),
              SizedBox(height: 8),
              Wrap(
                spacing: 8,
                children: [
                  ChoiceChip(
                    label: Row(mainAxisSize: MainAxisSize.min, children: [ Icon(Icons.color_lens, size: 18), SizedBox(width: 4), Text(global.language("color"))]),
                    selected: widget.colorMode,
                    onSelected: (_) => widget.onColorModeChanged(true),
                    selectedColor: global.theme.appBarColor.withValues(alpha: 0.3),
                    backgroundColor: global.theme.dividerBorderColor,
                  ),
                  ChoiceChip(
                    label: Row(mainAxisSize: MainAxisSize.min, children: [ Icon(Icons.invert_colors, size: 18), SizedBox(width: 4), Text(global.language("black_and_white"))]),
                    selected: !widget.colorMode,
                    onSelected: (_) => widget.onColorModeChanged(false),
                    selectedColor: global.theme.appBarColor.withValues(alpha: 0.3),
                    backgroundColor: global.theme.dividerBorderColor,
                  ),
                ],
              ),
              const SizedBox(height: 12),

              // Dual Currency Section - แสดงตลอด, disable เมื่อเอกสารไม่ได้ใช้ multi-currency
              Text(global.language("currency"), style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold)),
              const SizedBox(height: 8),
              Wrap(
                spacing: 8,
                children: [
                  ChoiceChip(
                    label: Row(mainAxisSize: MainAxisSize.min, children: [
                      Icon(Icons.currency_exchange, size: 18, color: widget.isMultiCurrency ? null : global.theme.textSecondaryColor),
                      const SizedBox(width: 4),
                      Text(global.language("dual_currency"), style: TextStyle(color: widget.isMultiCurrency ? null : global.theme.textSecondaryColor)),
                    ]),
                    selected: widget.showDualCurrency && widget.isMultiCurrency,
                    onSelected: widget.isMultiCurrency ? (_) => widget.onShowDualCurrencyChanged(true) : null,
                    selectedColor: global.theme.appBarColor.withValues(alpha: 0.3),
                    backgroundColor: global.theme.dividerBorderColor,
                    disabledColor: global.theme.cardColor,
                  ),
                  ChoiceChip(
                    label: Row(mainAxisSize: MainAxisSize.min, children: [const Icon(Icons.attach_money, size: 18), const SizedBox(width: 4), Text(global.language("single_currency"))]),
                    selected: !widget.showDualCurrency || !widget.isMultiCurrency,
                    onSelected: widget.isMultiCurrency ? (_) => widget.onShowDualCurrencyChanged(false) : null,
                    selectedColor: global.theme.appBarColor.withValues(alpha: 0.3),
                    backgroundColor: global.theme.dividerBorderColor,
                  ),
                ],
              ),
              const SizedBox(height: 12),

              // Theme Section - ส่วนเลือกธีม (Wrap layout)
              Text(global.language("document_theme"), style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold)),
              const SizedBox(height: 8),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: PdfPredefinedThemes.all.map((theme) {
                  final isSelected = widget.themeName == theme.name;
                  final previewColor = PdfPredefinedThemes.getPreviewColor(theme.name);
                  final displayName = PdfPredefinedThemes.getDisplayName(theme.name);
                  return ChoiceChip(
                    avatar: Container(
                      width: 18,
                      height: 18,
                      decoration: BoxDecoration(
                        color: previewColor,
                        shape: BoxShape.circle,
                        border: Border.all(color: global.theme.cardColor, width: 1),
                      ),
                    ),
                    label: Text(displayName),
                    selected: isSelected,
                    onSelected: (_) => widget.onThemeChanged(theme.name),
                    selectedColor: previewColor.withValues(alpha: 0.3),
                    backgroundColor: global.theme.dividerBorderColor,
                  );
                }).toList(),
              ),
              const SizedBox(height: 12),

              // Template Section - ส่วนเลือกรูปแบบเอกสาร
              Text(global.language("document_template"), style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold)),
              const SizedBox(height: 8),
              ...PdfPredefinedTemplates.all.map((template) {
                final isSelected = widget.templateId == template.id;
                return Padding(
                  padding: const EdgeInsets.only(bottom: 6),
                  child: InkWell(
                    onTap: () => widget.onTemplateChanged(template.id),
                    borderRadius: BorderRadius.circular(6),
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
                      decoration: BoxDecoration(
                        color: isSelected ? global.theme.appBarColor.withValues(alpha: 0.15) : global.theme.cardColor,
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(color: isSelected ? global.theme.appBarColor : global.theme.dividerBorderColor, width: isSelected ? 2 : 1),
                      ),
                      child: Row(
                        children: [
                          // ไอคอนแสดงรูปแบบ
                          Container(
                            width: 32,
                            height: 32,
                            decoration: BoxDecoration(color: isSelected ? global.theme.appBarColor : global.theme.iconSecondaryColor, borderRadius: BorderRadius.circular(4)),
                            child: Icon(_getTemplateIcon(template.id), color: Colors.white, size: 18),
                          ),
                          const SizedBox(width: 10),
                          // ชื่อและคำอธิบาย
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  template.name,
                                  style: TextStyle(fontSize: 13, fontWeight: isSelected ? FontWeight.bold : FontWeight.w500, color: isSelected ? global.theme.appBarColor : global.theme.textColor),
                                ),
                                Text(template.description, style: TextStyle(fontSize: 10, color: global.theme.iconSecondaryColor)),
                              ],
                            ),
                          ),
                          // เครื่องหมายถูกเมื่อเลือก
                          if (isSelected) Icon(Icons.check_circle, color: global.theme.appBarColor, size: 18),
                        ],
                      ),
                    ),
                  ),
                );
              }),
              const SizedBox(height: 12),

              // Date Format Section - รูปแบบวันที่
              Text(global.language("date_format"), style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold)),
              const SizedBox(height: 8),
              // เลือกปี พ.ศ./ค.ศ.
              Row(
                children: [
                  Text('${global.language("year")}: ', style: TextStyle(fontSize: 13)),
                  SizedBox(width: 8),
                  ChoiceChip(
                    label: Text(global.language("buddhist_era")),
                    selected: widget.dateSettings.yearFormat == PdfYearFormat.buddhist,
                    onSelected: (_) => widget.onDateSettingsChanged(widget.dateSettings.copyWith(yearFormat: PdfYearFormat.buddhist)),
                    selectedColor: global.theme.appBarColor.withValues(alpha: 0.3),
                    backgroundColor: global.theme.dividerBorderColor,
                  ),
                  SizedBox(width: 8),
                  ChoiceChip(
                    label: Text(global.language("christian_era")),
                    selected: widget.dateSettings.yearFormat == PdfYearFormat.christian,
                    onSelected: (_) => widget.onDateSettingsChanged(widget.dateSettings.copyWith(yearFormat: PdfYearFormat.christian)),
                    selectedColor: global.theme.appBarColor.withValues(alpha: 0.3),
                    backgroundColor: global.theme.dividerBorderColor,
                  ),
                ],
              ),
              const SizedBox(height: 8),
              // เลือกรูปแบบวันที่
              Wrap(
                spacing: 6,
                runSpacing: 6,
                children: PdfDateFormat.values.map((format) {
                  final isSelected = widget.dateSettings.dateFormat == format;
                  final displayName = PdfDateSettings.getDateFormatDisplayName(format, widget.dateSettings.yearFormat);
                  return ChoiceChip(
                    label: Text(displayName, style: const TextStyle(fontSize: 11)),
                    selected: isSelected,
                    onSelected: (_) => widget.onDateSettingsChanged(widget.dateSettings.copyWith(dateFormat: format)),
                    selectedColor: global.theme.appBarColor.withValues(alpha: 0.3),
                    backgroundColor: global.theme.dividerBorderColor,
                  );
                }).toList(),
              ),
              const SizedBox(height: 12),

              // Language Section - ภาษาเอกสาร
              Text(global.language("document_language"), style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold)),
              const SizedBox(height: 8),
              // แสดงรายการภาษาพร้อมธง
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: PdfSupportedLanguages.all.map((lang) {
                  final code = lang['code']!;
                  final name = lang['name']!;
                  final flag = lang['flag']!;
                  final isSelected = widget.languageSettings.languageCode == code;
                  final isAiTranslate = !PdfSupportedLanguages.primaryLanguages.contains(code);

                  return InkWell(
                    onTap: () => widget.onLanguageSettingsChanged(widget.languageSettings.copyWith(languageCode: code)),
                    borderRadius: BorderRadius.circular(8),
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                      decoration: BoxDecoration(
                        color: isSelected ? global.theme.appBarColor.withValues(alpha: 0.2) : global.theme.cardColor,
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(color: isSelected ? global.theme.appBarColor : global.theme.dividerBorderColor, width: isSelected ? 2 : 1),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          // ธงชาติ
                          ClipRRect(
                            borderRadius: BorderRadius.circular(2),
                            child: Image.asset(flag, width: 20, height: 14, fit: BoxFit.cover),
                          ),
                          const SizedBox(width: 6),
                          // ชื่อภาษา
                          Text(
                            name,
                            style: TextStyle(fontSize: 12, fontWeight: isSelected ? FontWeight.bold : FontWeight.normal, color: isSelected ? global.theme.appBarColor : global.theme.textColor),
                          ),
                          // แสดง AI badge สำหรับภาษาที่ต้องแปล
                          if (isAiTranslate) ...[
                            const SizedBox(width: 4),
                            Container(
                              padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 1),
                              decoration: BoxDecoration(color: global.theme.rowSelectedColor, borderRadius: BorderRadius.circular(4)),
                              child: const Text(
                                'AI',
                                style: TextStyle(fontSize: 8, fontWeight: FontWeight.bold, color: Colors.orange),
                              ),
                            ),
                          ],
                        ],
                      ),
                    ),
                  );
                }).toList(),
              ),
              // คำอธิบาย AI แปลภาษา
              if (!widget.languageSettings.isPrimaryLanguage)
                Padding(
                  padding: EdgeInsets.only(top: 8),
                  child: Row(
                    children: [
                      Icon(Icons.auto_awesome, size: 14, color: Colors.orange[700]),
                      SizedBox(width: 4),
                      Expanded(
                        child: Text(
                          global.language("ai_auto_translate"),
                          style: TextStyle(fontSize: 11, color: Colors.orange[700], fontStyle: FontStyle.italic),
                        ),
                      ),
                    ],
                  ),
                ),
              const SizedBox(height: 12),

              // Print Button Section (ด้านล่าง)
              SizedBox(
                width: double.infinity,
                child: ElevatedButton.icon(
                  icon: Icon(Icons.print, size: 20),
                  label: Text(global.language("print_document"), style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold)),
                  style: ElevatedButton.styleFrom(backgroundColor: global.theme.appBarColor, foregroundColor: global.theme.onPrimaryColor, padding: EdgeInsets.symmetric(vertical: 12), elevation: 2),
                  onPressed: widget.pdfBytes != null ? widget.onPrint : null,
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

/// Separate widget to prevent unnecessary rebuilds of PDF viewer
/// เพิ่มปุ่ม zoom และแสดงหน้าปัจจุบัน
class _PdfViewerBody extends StatefulWidget {
  final Uint8List pdfBytes;
  final String pdfKey;
  final String pageSize;
  final String orientation;

  const _PdfViewerBody({required this.pdfBytes, required this.pdfKey, required this.pageSize, required this.orientation});

  @override
  State<_PdfViewerBody> createState() => _PdfViewerBodyState();

  /// ตรวจสอบว่า PDF bytes ถูกต้องหรือไม่ (ต้องมีขนาดอย่างน้อย 100 bytes และขึ้นต้นด้วย %PDF)
  bool get isValidPdf {
    if (pdfBytes.isEmpty || pdfBytes.length < 100) return false;
    // ตรวจสอบ PDF header: %PDF
    if (pdfBytes.length >= 4) {
      final header = String.fromCharCodes(pdfBytes.sublist(0, 4));
      return header == '%PDF';
    }
    return false;
  }
}

class _PdfViewerBodyState extends State<_PdfViewerBody> with global.ThemeRefreshMixin {
  double _zoomLevel = -1; // default เต็มความสูง (fit height)
  final TransformationController _transformController = TransformationController();

  @override
  void dispose() {
    _transformController.dispose();
    super.dispose();
  }

  /// Zoom in ขยาย 20%
  void _zoomIn() {
    setState(() {
      _zoomLevel = (_zoomLevel + 0.2).clamp(0.5, 3.0);
    });
  }

  /// Zoom out ย่อ 20%
  void _zoomOut() {
    setState(() {
      _zoomLevel = (_zoomLevel - 0.2).clamp(0.5, 3.0);
    });
  }

  /// รีเซ็ต zoom กลับเป็น 100%
  void _resetZoom() {
    setState(() {
      _zoomLevel = 1.0;
    });
  }

  /// พอดีความกว้าง
  void _fitWidth() {
    setState(() {
      _zoomLevel = -2; // ใช้ค่าพิเศษสำหรับ fit width
    });
  }

  /// พอดีความสูง (เต็มหน้า)
  void _fitHeight() {
    setState(() {
      _zoomLevel = -1; // ใช้ค่าพิเศษสำหรับ fit height
    });
  }

  @override
  Widget build(BuildContext context) {
    // ตรวจสอบว่า PDF bytes ถูกต้องหรือไม่
    if (!widget.isValidPdf) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.error_outline, size: 64, color: Colors.red[300]),
            SizedBox(height: 16),
            Text(
              global.language("cannot_display_pdf"),
              style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.red),
            ),
            SizedBox(height: 8),
            Text('${global.language("file_size")}: ${widget.pdfBytes.length} bytes', style: TextStyle(fontSize: 14, color: global.theme.iconSecondaryColor)),
            SizedBox(height: 4),
            Text(global.language("try_change_settings"), style: TextStyle(fontSize: 13, color: global.theme.iconSecondaryColor)),
          ],
        ),
      );
    }

    // คำนวณขนาดหน้าจอที่มี
    final screenWidth = MediaQuery.of(context).size.width * 0.65; // พื้นที่แสดง PDF (หัก settings panel)
    final screenHeight = MediaQuery.of(context).size.height - 100; // หักส่วน AppBar และ margin

    // คำนวณ aspect ratio ของกระดาษ
    final pageFormat = PdfService.getPageFormat(widget.pageSize, widget.orientation);
    final pageAspectRatio = pageFormat.width / pageFormat.height;

    // คำนวณความกว้างที่เหมาะสมสำหรับแต่ละโหมด
    double maxWidth;
    if (_zoomLevel == -1) {
      // โหมดเต็มความสูง: คำนวณความกว้างจากความสูงที่มี
      maxWidth = screenHeight * pageAspectRatio;
    } else if (_zoomLevel == -2) {
      // โหมดพอดีความกว้าง: ใช้ความกว้างเต็ม
      maxWidth = screenWidth;
    } else {
      // โหมด zoom ปกติ
      maxWidth = screenWidth * _zoomLevel;
    }

    return Stack(
      children: [
        // PDF Preview พร้อม zoom - ใช้ PDF ที่โหลดมาแล้ว พร้อม error handling
        RepaintBoundary(
          child: _SafePdfPreview(pdfKey: widget.pdfKey, pdfBytes: widget.pdfBytes, pageSize: widget.pageSize, orientation: widget.orientation, maxWidth: maxWidth),
        ),
        // Zoom Controls - มุมขวาล่าง
        Positioned(
          right: 16,
          bottom: 16,
          child: Container(
            decoration: BoxDecoration(
              color: global.theme.cardColor,
              borderRadius: BorderRadius.circular(8),
              boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.15), blurRadius: 8, offset: const Offset(0, 2))],
            ),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                // Zoom In
                IconButton(
                  icon: Icon(Icons.add, size: 20),
                  tooltip: global.language("zoom_in"),
                  onPressed: _zoomLevel > 0 && _zoomLevel < 3.0 ? _zoomIn : (_zoomLevel < 0 ? () => setState(() => _zoomLevel = 1.2) : null),
                  padding: const EdgeInsets.all(8),
                  constraints: const BoxConstraints(minWidth: 36, minHeight: 36),
                ),
                // แสดงระดับ zoom ปัจจุบัน
                Container(
                  padding: EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  child: Text(_zoomLevel == -1 ? global.language("fit_height_short") : (_zoomLevel == -2 ? global.language("fit_width_short") : '${(_zoomLevel * 100).toInt()}%'), style: TextStyle(fontSize: 11, fontWeight: FontWeight.w500)),
                ),
                // Zoom Out
                IconButton(
                  icon: Icon(Icons.remove, size: 20),
                  tooltip: global.language("zoom_out"),
                  onPressed: _zoomLevel > 0.5 ? _zoomOut : (_zoomLevel < 0 ? () => setState(() => _zoomLevel = 0.8) : null),
                  padding: const EdgeInsets.all(8),
                  constraints: const BoxConstraints(minWidth: 36, minHeight: 36),
                ),
                const Divider(height: 1, thickness: 1),
                // รีเซ็ต zoom
                IconButton(
                  icon: Icon(Icons.fit_screen, size: 20),
                  tooltip: global.language("normal_size"),
                  onPressed: _resetZoom,
                  padding: const EdgeInsets.all(8),
                  constraints: const BoxConstraints(minWidth: 36, minHeight: 36),
                ),
                // พอดีความกว้าง
                IconButton(
                  icon: Icon(Icons.width_full, size: 20),
                  tooltip: global.language("fit_width"),
                  onPressed: _fitWidth,
                  padding: const EdgeInsets.all(8),
                  constraints: const BoxConstraints(minWidth: 36, minHeight: 36),
                ),
                // พอดีความสูง (เต็มหน้า)
                IconButton(
                  icon: Icon(Icons.height, size: 20),
                  tooltip: global.language("fit_height"),
                  onPressed: _fitHeight,
                  padding: const EdgeInsets.all(8),
                  constraints: const BoxConstraints(minWidth: 36, minHeight: 36),
                ),
              ],
            ),
          ),
        ),
      ],
    );
  }
}

/// Widget ที่ห่อ PdfPreview พร้อม error handling
/// ป้องกัน RangeError และ error อื่นๆ ที่อาจเกิดขึ้นขณะแสดง PDF
class _SafePdfPreview extends StatefulWidget {
  final String pdfKey;
  final Uint8List pdfBytes;
  final String pageSize;
  final String orientation;
  final double maxWidth;

  const _SafePdfPreview({required this.pdfKey, required this.pdfBytes, required this.pageSize, required this.orientation, required this.maxWidth});

  @override
  State<_SafePdfPreview> createState() => _SafePdfPreviewState();
}

class _SafePdfPreviewState extends State<_SafePdfPreview> with global.ThemeRefreshMixin {
  bool _hasError = false;
  String _errorMessage = '';

  @override
  void didUpdateWidget(covariant _SafePdfPreview oldWidget) {
    super.didUpdateWidget(oldWidget);
    // รีเซ็ต error state เมื่อ PDF เปลี่ยน
    if (oldWidget.pdfKey != widget.pdfKey) {
      setState(() {
        _hasError = false;
        _errorMessage = '';
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    // แสดง error widget ถ้ามี error
    if (_hasError) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.error_outline, size: 64, color: Colors.orange[300]),
            SizedBox(height: 16),
            Text(
              global.language("error_displaying_pdf"),
              style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.orange),
            ),
            const SizedBox(height: 8),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 32),
              child: Text(
                _errorMessage,
                style: TextStyle(fontSize: 13, color: global.theme.iconSecondaryColor),
                textAlign: TextAlign.center,
              ),
            ),
            SizedBox(height: 16),
            ElevatedButton.icon(
              icon: Icon(Icons.refresh),
              label: Text(global.language("try_again")),
              onPressed: () {
                setState(() {
                  _hasError = false;
                  _errorMessage = '';
                });
              },
            ),
          ],
        ),
      );
    }

    // ใช้ ErrorWidget.builder เพื่อจับ error ขณะ render
    return _ErrorBoundary(
      onError: (error, stackTrace) {
        AppLogger.error('PDF Preview Error: $error');
        if (mounted) {
          setState(() {
            _hasError = true;
            _errorMessage = error.toString();
          });
        }
      },
      child: PdfPreview(
        key: ValueKey(widget.pdfKey),
        build: (_) => Future.value(widget.pdfBytes),
        initialPageFormat: PdfService.getPageFormat(widget.pageSize, widget.orientation),
        canChangePageFormat: false,
        canChangeOrientation: false,
        allowPrinting: false,
        allowSharing: false,
        shouldRepaint: false,
        useActions: false,
        dynamicLayout: true,
        previewPageMargin: const EdgeInsets.all(4),
        maxPageWidth: widget.maxWidth,
        pdfPreviewPageDecoration: BoxDecoration(
          color: global.theme.cardColor,
          boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.2), blurRadius: 4, offset: const Offset(0, 2))],
        ),
        onError: (context, error) {
          // callback เมื่อเกิด error ใน PdfPreview
          AppLogger.error('PdfPreview onError: $error');
          return Center(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(Icons.warning_amber_rounded, size: 48, color: Colors.orange[400]),
                SizedBox(height: 12),
                Text(global.language("cannot_display_pdf"), style: TextStyle(fontSize: 16, fontWeight: FontWeight.w500)),
                SizedBox(height: 8),
                Text(global.language("try_change_settings"), style: TextStyle(fontSize: 13, color: global.theme.iconSecondaryColor)),
              ],
            ),
          );
        },
      ),
    );
  }
}

/// Error Boundary Widget สำหรับจับ error ใน child widget
class _ErrorBoundary extends StatefulWidget {
  final Widget child;
  final void Function(dynamic error, StackTrace stackTrace) onError;

  const _ErrorBoundary({required this.child, required this.onError});

  @override
  State<_ErrorBoundary> createState() => _ErrorBoundaryState();
}

class _ErrorBoundaryState extends State<_ErrorBoundary> with global.ThemeRefreshMixin {
  @override
  Widget build(BuildContext context) {
    // ใช้ FlutterError.onError override แบบ local ไม่ได้
    // แต่สามารถใช้ ErrorWidget.builder ได้
    ErrorWidget.builder = (FlutterErrorDetails details) {
      widget.onError(details.exception, details.stack ?? StackTrace.current);
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.error, size: 48, color: Colors.red[400]),
            SizedBox(height: 12),
            Text(global.language("error_occurred"), style: TextStyle(fontSize: 16, fontWeight: FontWeight.w500)),
          ],
        ),
      );
    };
    return widget.child;
  }
}
