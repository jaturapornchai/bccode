import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;

enum ElementType {
  text,
  image,
  line,
  rectangle,
  table,
  dataField,
  barcode,
  qrCode,
  separator,
  pageNumber,
  dateTime,
  signature,
}

extension ElementTypeExtension on ElementType {
  String get label {
    switch (this) {
      case ElementType.text:
        return global.language('fd_text');
      case ElementType.image:
        return global.language('fd_image');
      case ElementType.line:
        return global.language('fd_line');
      case ElementType.rectangle:
        return global.language('fd_rectangle');
      case ElementType.table:
        return global.language('fd_table');
      case ElementType.dataField:
        return global.language('fd_data_field');
      case ElementType.barcode:
        return global.language('fd_barcode');
      case ElementType.qrCode:
        return global.language('fd_qr_code');
      case ElementType.separator:
        return global.language('fd_separator');
      case ElementType.pageNumber:
        return global.language('fd_page_number');
      case ElementType.dateTime:
        return global.language('fd_date_time');
      case ElementType.signature:
        return global.language('fd_signature');
    }
  }

  IconData get icon {
    switch (this) {
      case ElementType.text:
        return Icons.text_fields;
      case ElementType.image:
        return Icons.image;
      case ElementType.line:
        return Icons.horizontal_rule;
      case ElementType.rectangle:
        return Icons.crop_square;
      case ElementType.table:
        return Icons.table_chart;
      case ElementType.dataField:
        return Icons.data_object;
      case ElementType.barcode:
        return Icons.view_week;
      case ElementType.qrCode:
        return Icons.qr_code;
      case ElementType.separator:
        return Icons.space_bar;
      case ElementType.pageNumber:
        return Icons.numbers;
      case ElementType.dateTime:
        return Icons.calendar_today;
      case ElementType.signature:
        return Icons.draw;
    }
  }

  String get category {
    switch (this) {
      case ElementType.text:
      case ElementType.image:
      case ElementType.line:
      case ElementType.rectangle:
      case ElementType.separator:
        return 'static';
      case ElementType.dataField:
        return 'data';
      case ElementType.table:
        return 'table';
      case ElementType.barcode:
      case ElementType.qrCode:
      case ElementType.pageNumber:
      case ElementType.dateTime:
      case ElementType.signature:
        return 'special';
    }
  }
}

class FormElement {
  String id;
  ElementType type;
  double x;
  double y;
  double width;
  double height;

  // Text
  String? text;
  String? fontFamily;
  double? fontSize;
  bool fontBold;
  bool fontItalic;
  bool fontUnderline;
  Color? textColor;
  TextAlign? textAlign;

  // Image
  String? imagePath;

  // Appearance
  Color? backgroundColor;
  Color? borderColor;
  double? borderWidth;
  double? borderRadius;
  double opacity;

  // Layout
  double rotation;
  int zIndex;
  bool isLocked;
  bool isVisible;

  // Data Binding
  String? dataBindingKey;
  String? dataBindingFormat;

  // Barcode/QR
  String? barcodeFormat;
  String? barcodeValue;

  // Separator/Line
  String? lineStyle; // solid, dashed, dotted

  // Table
  int? rows;
  int? columns;
  List<List<String>>? tableData;
  List<double>? columnWidths;
  List<int>? columnAligns; // 0=left, 1=center, 2=right per column
  bool tableHasHeader;

  // Legacy textStyle support (for backward compat reading)
  TextStyle? _legacyTextStyle;

  FormElement({
    required this.id,
    required this.type,
    this.x = 0,
    this.y = 0,
    this.width = 100,
    this.height = 50,
    this.text,
    this.fontFamily,
    this.fontSize,
    this.fontBold = false,
    this.fontItalic = false,
    this.fontUnderline = false,
    this.textColor,
    this.textAlign,
    this.imagePath,
    this.backgroundColor,
    this.borderColor,
    this.borderWidth,
    this.borderRadius,
    this.opacity = 1.0,
    this.rotation = 0,
    this.zIndex = 0,
    this.isLocked = false,
    this.isVisible = true,
    this.dataBindingKey,
    this.dataBindingFormat,
    this.barcodeFormat,
    this.barcodeValue,
    this.lineStyle,
    this.rows,
    this.columns,
    this.tableData,
    this.columnWidths,
    this.columnAligns,
    this.tableHasHeader = false,
    TextStyle? textStyle,
  }) : _legacyTextStyle = textStyle;

  /// Build TextStyle from individual properties (preferred) or legacy textStyle
  TextStyle get textStyle {
    if (fontSize != null || fontFamily != null || textColor != null) {
      return TextStyle(
        fontFamily: fontFamily,
        fontSize: fontSize ?? 14,
        fontWeight: fontBold ? FontWeight.bold : FontWeight.normal,
        fontStyle: fontItalic ? FontStyle.italic : FontStyle.normal,
        decoration: fontUnderline ? TextDecoration.underline : TextDecoration.none,
        color: textColor,
      );
    }
    return _legacyTextStyle ?? const TextStyle(fontSize: 14);
  }

  /// Display text for the element (used in layers panel, etc.)
  String get displayName {
    if (text != null && text!.isNotEmpty) {
      return text!.length > 20 ? '${text!.substring(0, 20)}...' : text!;
    }
    if (dataBindingKey != null) return '{${dataBindingKey!}}';
    return type.label;
  }

  FormElement copyWith({
    String? id,
    ElementType? type,
    double? x,
    double? y,
    double? width,
    double? height,
    String? text,
    String? fontFamily,
    double? fontSize,
    bool? fontBold,
    bool? fontItalic,
    bool? fontUnderline,
    Color? textColor,
    TextAlign? textAlign,
    String? imagePath,
    Color? backgroundColor,
    Color? borderColor,
    double? borderWidth,
    double? borderRadius,
    double? opacity,
    double? rotation,
    int? zIndex,
    bool? isLocked,
    bool? isVisible,
    String? dataBindingKey,
    String? dataBindingFormat,
    String? barcodeFormat,
    String? barcodeValue,
    String? lineStyle,
    int? rows,
    int? columns,
    List<List<String>>? tableData,
    List<double>? columnWidths,
    List<int>? columnAligns,
    bool? tableHasHeader,
  }) {
    return FormElement(
      id: id ?? this.id,
      type: type ?? this.type,
      x: x ?? this.x,
      y: y ?? this.y,
      width: width ?? this.width,
      height: height ?? this.height,
      text: text ?? this.text,
      fontFamily: fontFamily ?? this.fontFamily,
      fontSize: fontSize ?? this.fontSize,
      fontBold: fontBold ?? this.fontBold,
      fontItalic: fontItalic ?? this.fontItalic,
      fontUnderline: fontUnderline ?? this.fontUnderline,
      textColor: textColor ?? this.textColor,
      textAlign: textAlign ?? this.textAlign,
      imagePath: imagePath ?? this.imagePath,
      backgroundColor: backgroundColor ?? this.backgroundColor,
      borderColor: borderColor ?? this.borderColor,
      borderWidth: borderWidth ?? this.borderWidth,
      borderRadius: borderRadius ?? this.borderRadius,
      opacity: opacity ?? this.opacity,
      rotation: rotation ?? this.rotation,
      zIndex: zIndex ?? this.zIndex,
      isLocked: isLocked ?? this.isLocked,
      isVisible: isVisible ?? this.isVisible,
      dataBindingKey: dataBindingKey ?? this.dataBindingKey,
      dataBindingFormat: dataBindingFormat ?? this.dataBindingFormat,
      barcodeFormat: barcodeFormat ?? this.barcodeFormat,
      barcodeValue: barcodeValue ?? this.barcodeValue,
      lineStyle: lineStyle ?? this.lineStyle,
      rows: rows ?? this.rows,
      columns: columns ?? this.columns,
      tableData: tableData ?? this.tableData,
      columnWidths: columnWidths ?? this.columnWidths,
      columnAligns: columnAligns ?? this.columnAligns,
      tableHasHeader: tableHasHeader ?? this.tableHasHeader,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'type': type.name,
      'x': x,
      'y': y,
      'width': width,
      'height': height,
      'text': text,
      'fontFamily': fontFamily,
      'fontSize': fontSize ?? textStyle.fontSize,
      'fontWeight': fontBold ? FontWeight.bold.value : (textStyle.fontWeight?.value),
      'fontBold': fontBold,
      'fontItalic': fontItalic,
      'fontUnderline': fontUnderline,
      'color': textColor?.toARGB32() ?? textStyle.color?.toARGB32(),
      'textAlign': textAlign?.index,
      'imagePath': imagePath,
      'backgroundColor': backgroundColor?.toARGB32(),
      'borderColor': borderColor?.toARGB32(),
      'borderWidth': borderWidth,
      'borderRadius': borderRadius,
      'opacity': opacity,
      'rotation': rotation,
      'zIndex': zIndex,
      'isLocked': isLocked,
      'isVisible': isVisible,
      'dataBindingKey': dataBindingKey,
      'dataBindingFormat': dataBindingFormat,
      'barcodeFormat': barcodeFormat,
      'barcodeValue': barcodeValue,
      'lineStyle': lineStyle,
      'rows': rows,
      'columns': columns,
      'tableData': tableData,
      'columnWidths': columnWidths,
      'columnAligns': columnAligns,
      'tableHasHeader': tableHasHeader,
    };
  }

  factory FormElement.fromJson(Map<String, dynamic> json) {
    // Backward compat: read fontBold from either new field or legacy fontWeight
    final legacyFontWeight = json['fontWeight'] as int?;
    final fontBold = json['fontBold'] as bool? ??
        (legacyFontWeight != null && legacyFontWeight >= 5);

    return FormElement(
      id: json['id'] as String,
      type: ElementType.values.firstWhere(
        (e) => e.name == json['type'],
        orElse: () => ElementType.text,
      ),
      x: (json['x'] as num?)?.toDouble() ?? 0,
      y: (json['y'] as num?)?.toDouble() ?? 0,
      width: (json['width'] as num?)?.toDouble() ?? 100,
      height: (json['height'] as num?)?.toDouble() ?? 50,
      text: json['text'] as String?,
      fontFamily: json['fontFamily'] as String?,
      fontSize: (json['fontSize'] as num?)?.toDouble(),
      fontBold: fontBold,
      fontItalic: json['fontItalic'] as bool? ?? false,
      fontUnderline: json['fontUnderline'] as bool? ?? false,
      textColor: json['color'] != null ? Color(json['color'] as int) : null,
      textAlign: json['textAlign'] != null
          ? TextAlign.values[json['textAlign'] as int]
          : null,
      imagePath: json['imagePath'] as String?,
      backgroundColor: json['backgroundColor'] != null
          ? Color(json['backgroundColor'] as int)
          : null,
      borderColor: json['borderColor'] != null
          ? Color(json['borderColor'] as int)
          : null,
      borderWidth: (json['borderWidth'] as num?)?.toDouble(),
      borderRadius: (json['borderRadius'] as num?)?.toDouble(),
      opacity: (json['opacity'] as num?)?.toDouble() ?? 1.0,
      rotation: (json['rotation'] as num?)?.toDouble() ?? 0,
      zIndex: json['zIndex'] as int? ?? 0,
      isLocked: json['isLocked'] as bool? ?? false,
      isVisible: json['isVisible'] as bool? ?? true,
      dataBindingKey: json['dataBindingKey'] as String?,
      dataBindingFormat: json['dataBindingFormat'] as String?,
      barcodeFormat: json['barcodeFormat'] as String?,
      barcodeValue: json['barcodeValue'] as String?,
      lineStyle: json['lineStyle'] as String?,
      rows: json['rows'] as int?,
      columns: json['columns'] as int?,
      tableData: json['tableData'] != null
          ? (json['tableData'] as List)
              .map((row) =>
                  (row as List).map((cell) => cell.toString()).toList())
              .toList()
          : null,
      columnWidths: json['columnWidths'] != null
          ? (json['columnWidths'] as List)
              .map((w) => (w as num).toDouble())
              .toList()
          : null,
      columnAligns: json['columnAligns'] != null
          ? (json['columnAligns'] as List)
              .map((a) => (a as num).toInt())
              .toList()
          : null,
      tableHasHeader: json['tableHasHeader'] as bool? ?? false,
    );
  }
}
