import 'package:flutter/material.dart';
import 'form_section.dart';
import 'paper_size.dart';

/// ประเภทหน้าเอกสาร สำหรับเลือก section ที่จะแสดง/แก้ไข
enum PageType {
  first,
  middle,
  last,
}

extension PageTypeExtension on PageType {
  String get label {
    switch (this) {
      case PageType.first:
        return 'หน้าแรก';
      case PageType.middle:
        return 'หน้ากลาง';
      case PageType.last:
        return 'หน้าสุดท้าย';
    }
  }

  IconData get icon {
    switch (this) {
      case PageType.first:
        return Icons.first_page;
      case PageType.middle:
        return Icons.menu_book;
      case PageType.last:
        return Icons.last_page;
    }
  }
}

enum PageOrientation {
  portrait,
  landscape,
}

extension PageOrientationExtension on PageOrientation {
  String get label {
    switch (this) {
      case PageOrientation.portrait:
        return 'แนวตั้ง';
      case PageOrientation.landscape:
        return 'แนวนอน';
    }
  }

  IconData get icon {
    switch (this) {
      case PageOrientation.portrait:
        return Icons.stay_current_portrait;
      case PageOrientation.landscape:
        return Icons.stay_current_landscape;
    }
  }

  String toJson() => name;

  static PageOrientation fromJson(String? json) {
    if (json == null) return PageOrientation.portrait;
    for (final o in PageOrientation.values) {
      if (o.name == json) return o;
    }
    return PageOrientation.portrait;
  }
}

enum PrintMode {
  color,
  grayscale,
  blackAndWhite,
}

extension PrintModeExtension on PrintMode {
  String get label {
    switch (this) {
      case PrintMode.color:
        return 'สี';
      case PrintMode.grayscale:
        return 'ขาวดำ (Grayscale)';
      case PrintMode.blackAndWhite:
        return 'ขาวดำ (B&W)';
    }
  }

  String toJson() => name;

  static PrintMode fromJson(String? json) {
    if (json == null) return PrintMode.color;
    for (final mode in PrintMode.values) {
      if (mode.name == json) return mode;
    }
    return PrintMode.color;
  }
}

class FormTemplate {
  String id;
  String name;
  String description;
  PaperSize paperSize;

  // ── Page sections ──
  // หน้าแรก: header + detail + footer
  // หน้ากลาง: headerMiddle + detail + footerMiddle
  // หน้าสุดท้าย: headerLast + detail + footer (= last page footer)
  FormSection header; // = first page header (always required)
  FormSection detail; // = shared detail section
  FormSection footer; // = last page footer (always required)

  // Optional multi-page overrides (null = ใช้ค่า default)
  FormSection? headerMiddle; // header หน้า 2+ (null → เหมือน header)
  FormSection? headerLast; // header หน้าสุดท้าย (null → เหมือน headerMiddle)
  FormSection? footerFirst; // footer หน้าแรก (null → ว่าง/เรียบง่าย)
  FormSection? footerMiddle; // footer หน้ากลาง (null → ว่าง/เรียบง่าย)

  String? globalBackgroundImagePath;
  DateTime createdAt;
  DateTime updatedAt;

  PrintMode printMode;
  bool multiPageEnabled;
  int detailRowsPerPage;
  String? documentType;
  double marginTop;
  double marginBottom;
  double marginLeft;
  double marginRight;
  PageOrientation orientation;

  FormTemplate({
    String? id,
    required this.name,
    this.description = '',
    this.paperSize = PaperSize.a4,
    FormSection? header,
    FormSection? detail,
    FormSection? footer,
    this.headerMiddle,
    this.headerLast,
    this.footerFirst,
    this.footerMiddle,
    this.globalBackgroundImagePath,
    DateTime? createdAt,
    DateTime? updatedAt,
    this.printMode = PrintMode.color,
    this.multiPageEnabled = true,
    this.detailRowsPerPage = 20,
    this.documentType,
    this.marginTop = 20,
    this.marginBottom = 20,
    this.marginLeft = 20,
    this.marginRight = 20,
    this.orientation = PageOrientation.portrait,
  })  : id = id ?? DateTime.now().millisecondsSinceEpoch.toString(),
        header = header ??
            FormSection(
              id: 'header',
              type: SectionType.header,
              height: 150,
            ),
        detail = detail ??
            FormSection(
              id: 'detail',
              type: SectionType.detail,
              height: 400,
            ),
        footer = footer ??
            FormSection(
              id: 'footer',
              type: SectionType.footer,
              height: 100,
            ),
        createdAt = createdAt ?? DateTime.now(),
        updatedAt = updatedAt ?? DateTime.now();

  // ── Page-aware section getters ──

  /// ดึง header section ตาม PageType
  FormSection headerForPage(PageType type) {
    switch (type) {
      case PageType.first:
        return header;
      case PageType.middle:
        return headerMiddle ?? header;
      case PageType.last:
        return headerLast ?? headerMiddle ?? header;
    }
  }

  /// ดึง footer section ตาม PageType
  FormSection footerForPage(PageType type) {
    switch (type) {
      case PageType.first:
        return footerFirst ?? footer;
      case PageType.middle:
        return footerMiddle ?? footer;
      case PageType.last:
        return footer;
    }
  }

  /// totalHeight สำหรับ page type ที่กำลังแสดง
  double totalHeightForPage(PageType type) {
    return headerForPage(type).height + detail.height + footerForPage(type).height;
  }

  /// กว้างจริงตาม orientation (landscape สลับ w↔h)
  double get effectiveWidth =>
      orientation == PageOrientation.landscape
          ? paperSize.height
          : paperSize.width;

  /// สูงจริงตาม orientation
  double get effectiveHeight =>
      orientation == PageOrientation.landscape
          ? paperSize.width
          : paperSize.height;

  double get totalHeight => header.height + detail.height + footer.height;

  /// sections ทั้งหมดที่มี (สำหรับ auto-fit ทุก section)
  List<FormSection> get allSections => [
        header,
        if (headerMiddle != null) headerMiddle!,
        if (headerLast != null) headerLast!,
        detail,
        footer,
        if (footerFirst != null) footerFirst!,
        if (footerMiddle != null) footerMiddle!,
      ];

  FormTemplate copyWith({
    String? id,
    String? name,
    String? description,
    PaperSize? paperSize,
    FormSection? header,
    FormSection? detail,
    FormSection? footer,
    FormSection? headerMiddle,
    FormSection? headerLast,
    FormSection? footerFirst,
    FormSection? footerMiddle,
    bool clearHeaderMiddle = false,
    bool clearHeaderLast = false,
    bool clearFooterFirst = false,
    bool clearFooterMiddle = false,
    String? globalBackgroundImagePath,
    DateTime? createdAt,
    DateTime? updatedAt,
    PrintMode? printMode,
    bool? multiPageEnabled,
    int? detailRowsPerPage,
    String? documentType,
    double? marginTop,
    double? marginBottom,
    double? marginLeft,
    double? marginRight,
    PageOrientation? orientation,
  }) {
    return FormTemplate(
      id: id ?? this.id,
      name: name ?? this.name,
      description: description ?? this.description,
      paperSize: paperSize ?? this.paperSize,
      header: header ?? this.header,
      detail: detail ?? this.detail,
      footer: footer ?? this.footer,
      headerMiddle:
          clearHeaderMiddle ? null : (headerMiddle ?? this.headerMiddle),
      headerLast: clearHeaderLast ? null : (headerLast ?? this.headerLast),
      footerFirst:
          clearFooterFirst ? null : (footerFirst ?? this.footerFirst),
      footerMiddle:
          clearFooterMiddle ? null : (footerMiddle ?? this.footerMiddle),
      globalBackgroundImagePath:
          globalBackgroundImagePath ?? this.globalBackgroundImagePath,
      createdAt: createdAt ?? this.createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
      printMode: printMode ?? this.printMode,
      multiPageEnabled: multiPageEnabled ?? this.multiPageEnabled,
      detailRowsPerPage: detailRowsPerPage ?? this.detailRowsPerPage,
      documentType: documentType ?? this.documentType,
      marginTop: marginTop ?? this.marginTop,
      marginBottom: marginBottom ?? this.marginBottom,
      marginLeft: marginLeft ?? this.marginLeft,
      marginRight: marginRight ?? this.marginRight,
      orientation: orientation ?? this.orientation,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name': name,
      'description': description,
      'paperSize': paperSize.toJson(),
      'header': header.toJson(),
      'detail': detail.toJson(),
      'footer': footer.toJson(),
      if (headerMiddle != null) 'headerMiddle': headerMiddle!.toJson(),
      if (headerLast != null) 'headerLast': headerLast!.toJson(),
      if (footerFirst != null) 'footerFirst': footerFirst!.toJson(),
      if (footerMiddle != null) 'footerMiddle': footerMiddle!.toJson(),
      'globalBackgroundImagePath': globalBackgroundImagePath,
      'createdAt': createdAt.toIso8601String(),
      'updatedAt': updatedAt.toIso8601String(),
      'printMode': printMode.toJson(),
      'multiPageEnabled': multiPageEnabled,
      'detailRowsPerPage': detailRowsPerPage,
      'documentType': documentType,
      'marginTop': marginTop,
      'marginBottom': marginBottom,
      'marginLeft': marginLeft,
      'marginRight': marginRight,
      'orientation': orientation.toJson(),
    };
  }

  factory FormTemplate.fromJson(Map<String, dynamic> json) {
    return FormTemplate(
      id: json['id'] as String,
      name: json['name'] as String,
      description: json['description'] as String? ?? '',
      paperSize: PaperSizeExtension.fromJson(json['paperSize'] as String),
      header: FormSection.fromJson(json['header'] as Map<String, dynamic>),
      detail: FormSection.fromJson(json['detail'] as Map<String, dynamic>),
      footer: FormSection.fromJson(json['footer'] as Map<String, dynamic>),
      headerMiddle: json['headerMiddle'] != null
          ? FormSection.fromJson(json['headerMiddle'] as Map<String, dynamic>)
          : null,
      headerLast: json['headerLast'] != null
          ? FormSection.fromJson(json['headerLast'] as Map<String, dynamic>)
          : null,
      footerFirst: json['footerFirst'] != null
          ? FormSection.fromJson(json['footerFirst'] as Map<String, dynamic>)
          : null,
      footerMiddle: json['footerMiddle'] != null
          ? FormSection.fromJson(json['footerMiddle'] as Map<String, dynamic>)
          : null,
      globalBackgroundImagePath: json['globalBackgroundImagePath'] as String?,
      createdAt: DateTime.parse(json['createdAt'] as String),
      updatedAt: DateTime.parse(json['updatedAt'] as String),
      printMode: PrintModeExtension.fromJson(json['printMode'] as String?),
      multiPageEnabled: json['multiPageEnabled'] as bool? ?? true,
      detailRowsPerPage: json['detailRowsPerPage'] as int? ?? 20,
      documentType: json['documentType'] as String?,
      marginTop: (json['marginTop'] as num?)?.toDouble() ?? 20,
      marginBottom: (json['marginBottom'] as num?)?.toDouble() ?? 20,
      marginLeft: (json['marginLeft'] as num?)?.toDouble() ?? 20,
      marginRight: (json['marginRight'] as num?)?.toDouble() ?? 20,
      orientation:
          PageOrientationExtension.fromJson(json['orientation'] as String?),
    );
  }
}
