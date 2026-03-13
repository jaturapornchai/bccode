import 'package:flutter/material.dart';
import 'form_element.dart';

/// Layout cell — เหมือน Widget ใน Flutter Row
/// ใช้ flex หรือ fixedWidth กำหนดความกว้าง
class FormLayoutCell {
  String id;
  ElementType type;

  // ── Size ──
  double flex; // flex ratio (0 = ใช้ fixedWidth)
  double? fixedWidth; // px — ถ้าตั้งค่า flex จะถูกละเว้น
  double? minHeight; // ความสูงขั้นต่ำ

  // ── Content ──
  String? text;
  String? dataBindingKey;
  String? dataBindingFormat;
  String? imagePath;

  // ── Text Style ──
  String fontFamily;
  double fontSize;
  bool fontBold;
  bool fontItalic;
  bool fontUnderline;
  TextAlign textAlign;
  Color? textColor;

  // ── Appearance ──
  Color? backgroundColor;
  Color? borderColor;
  double? borderWidth;
  double? borderRadius;
  double opacity;

  // ── Barcode/QR ──
  String? barcodeFormat;
  String? barcodeValue;

  // ── Line/Separator ──
  String? lineStyle;

  // ── Table ──
  int? rows;
  int? columns;
  List<List<String>>? tableData;
  List<int>? columnAligns; // 0=left, 1=center, 2=right per column
  bool tableHasHeader;

  FormLayoutCell({
    String? id,
    required this.type,
    this.flex = 1.0,
    this.fixedWidth,
    this.minHeight,
    this.text,
    this.dataBindingKey,
    this.dataBindingFormat,
    this.imagePath,
    this.fontFamily = 'Sarabun',
    this.fontSize = 7.0,
    this.fontBold = false,
    this.fontItalic = false,
    this.fontUnderline = false,
    this.textAlign = TextAlign.left,
    this.textColor,
    this.backgroundColor,
    this.borderColor,
    this.borderWidth,
    this.borderRadius,
    this.opacity = 1.0,
    this.barcodeFormat,
    this.barcodeValue,
    this.lineStyle,
    this.rows,
    this.columns,
    this.tableData,
    this.columnAligns,
    this.tableHasHeader = false,
  }) : id = id ?? DateTime.now().microsecondsSinceEpoch.toString();

  bool get useFixedWidth => fixedWidth != null && fixedWidth! > 0;

  /// คำนวณ height จาก content ที่ width กำหนด
  double estimateHeight(double cellWidth) {
    if (type == ElementType.image || type == ElementType.signature) {
      return minHeight ?? 40;
    }
    if (type == ElementType.line || type == ElementType.separator) {
      return 2;
    }
    if (type == ElementType.table) {
      final rowCount = rows ?? 3;
      final rowH = fontSize * 1.5;
      final headerH = fontSize * 1.8;
      return headerH + rowCount * rowH + 4;
    }
    // Text-like: estimate from content
    final content = text ?? dataBindingKey ?? '';
    if (content.isEmpty) return fontSize * 1.5 + 4;
    final charWidth = fontSize * 0.52;
    final charsPerLine = (cellWidth / charWidth).floor().clamp(1, 9999);
    final numLines = (content.length / charsPerLine).ceil().clamp(1, 999);
    final h = numLines * fontSize * 1.45 + 4;
    return minHeight != null && minHeight! > h ? minHeight! : h;
  }

  FormLayoutCell copyWith({
    String? id,
    ElementType? type,
    double? flex,
    double? fixedWidth,
    bool clearFixedWidth = false,
    double? minHeight,
    String? text,
    String? dataBindingKey,
    String? dataBindingFormat,
    String? imagePath,
    String? fontFamily,
    double? fontSize,
    bool? fontBold,
    bool? fontItalic,
    bool? fontUnderline,
    TextAlign? textAlign,
    Color? textColor,
    Color? backgroundColor,
    Color? borderColor,
    double? borderWidth,
    double? borderRadius,
    double? opacity,
    String? barcodeFormat,
    String? barcodeValue,
    String? lineStyle,
    int? rows,
    int? columns,
    List<List<String>>? tableData,
    List<int>? columnAligns,
    bool? tableHasHeader,
  }) {
    return FormLayoutCell(
      id: id ?? this.id,
      type: type ?? this.type,
      flex: flex ?? this.flex,
      fixedWidth: clearFixedWidth ? null : (fixedWidth ?? this.fixedWidth),
      minHeight: minHeight ?? this.minHeight,
      text: text ?? this.text,
      dataBindingKey: dataBindingKey ?? this.dataBindingKey,
      dataBindingFormat: dataBindingFormat ?? this.dataBindingFormat,
      imagePath: imagePath ?? this.imagePath,
      fontFamily: fontFamily ?? this.fontFamily,
      fontSize: fontSize ?? this.fontSize,
      fontBold: fontBold ?? this.fontBold,
      fontItalic: fontItalic ?? this.fontItalic,
      fontUnderline: fontUnderline ?? this.fontUnderline,
      textAlign: textAlign ?? this.textAlign,
      textColor: textColor ?? this.textColor,
      backgroundColor: backgroundColor ?? this.backgroundColor,
      borderColor: borderColor ?? this.borderColor,
      borderWidth: borderWidth ?? this.borderWidth,
      borderRadius: borderRadius ?? this.borderRadius,
      opacity: opacity ?? this.opacity,
      barcodeFormat: barcodeFormat ?? this.barcodeFormat,
      barcodeValue: barcodeValue ?? this.barcodeValue,
      lineStyle: lineStyle ?? this.lineStyle,
      rows: rows ?? this.rows,
      columns: columns ?? this.columns,
      tableData: tableData ?? this.tableData,
      columnAligns: columnAligns ?? this.columnAligns,
      tableHasHeader: tableHasHeader ?? this.tableHasHeader,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'type': type.name,
      'flex': flex,
      if (fixedWidth != null) 'fixedWidth': fixedWidth,
      if (minHeight != null) 'minHeight': minHeight,
      if (text != null) 'text': text,
      if (dataBindingKey != null) 'dataBindingKey': dataBindingKey,
      if (dataBindingFormat != null) 'dataBindingFormat': dataBindingFormat,
      if (imagePath != null) 'imagePath': imagePath,
      'fontFamily': fontFamily,
      'fontSize': fontSize,
      'fontBold': fontBold,
      'fontItalic': fontItalic,
      'fontUnderline': fontUnderline,
      'textAlign': textAlign.index,
      if (textColor != null) 'textColor': textColor!.toARGB32(),
      if (backgroundColor != null)
        'backgroundColor': backgroundColor!.toARGB32(),
      if (borderColor != null) 'borderColor': borderColor!.toARGB32(),
      if (borderWidth != null) 'borderWidth': borderWidth,
      if (borderRadius != null) 'borderRadius': borderRadius,
      if (opacity != 1.0) 'opacity': opacity,
      if (barcodeFormat != null) 'barcodeFormat': barcodeFormat,
      if (barcodeValue != null) 'barcodeValue': barcodeValue,
      if (lineStyle != null) 'lineStyle': lineStyle,
      if (rows != null) 'rows': rows,
      if (columns != null) 'columns': columns,
      if (tableData != null) 'tableData': tableData,
      if (columnAligns != null) 'columnAligns': columnAligns,
      if (tableHasHeader) 'tableHasHeader': tableHasHeader,
    };
  }

  factory FormLayoutCell.fromJson(Map<String, dynamic> json) {
    return FormLayoutCell(
      id: json['id'] as String? ??
          DateTime.now().microsecondsSinceEpoch.toString(),
      type: ElementType.values.firstWhere(
        (e) => e.name == json['type'],
        orElse: () => ElementType.text,
      ),
      flex: (json['flex'] as num?)?.toDouble() ?? 1.0,
      fixedWidth: (json['fixedWidth'] as num?)?.toDouble(),
      minHeight: (json['minHeight'] as num?)?.toDouble(),
      text: json['text'] as String?,
      dataBindingKey: json['dataBindingKey'] as String?,
      dataBindingFormat: json['dataBindingFormat'] as String?,
      imagePath: json['imagePath'] as String?,
      fontFamily: json['fontFamily'] as String? ?? 'Sarabun',
      fontSize: (json['fontSize'] as num?)?.toDouble() ?? 7.0,
      fontBold: json['fontBold'] as bool? ?? false,
      fontItalic: json['fontItalic'] as bool? ?? false,
      fontUnderline: json['fontUnderline'] as bool? ?? false,
      textAlign: json['textAlign'] != null
          ? TextAlign.values[(json['textAlign'] as int).clamp(0, 5)]
          : TextAlign.left,
      textColor:
          json['textColor'] != null ? Color(json['textColor'] as int) : null,
      backgroundColor: json['backgroundColor'] != null
          ? Color(json['backgroundColor'] as int)
          : null,
      borderColor: json['borderColor'] != null
          ? Color(json['borderColor'] as int)
          : null,
      borderWidth: (json['borderWidth'] as num?)?.toDouble(),
      borderRadius: (json['borderRadius'] as num?)?.toDouble(),
      opacity: (json['opacity'] as num?)?.toDouble() ?? 1.0,
      barcodeFormat: json['barcodeFormat'] as String?,
      barcodeValue: json['barcodeValue'] as String?,
      lineStyle: json['lineStyle'] as String?,
      rows: json['rows'] as int?,
      columns: json['columns'] as int?,
      tableData: json['tableData'] != null
          ? (json['tableData'] as List)
              .map(
                  (row) => (row as List).map((c) => c.toString()).toList())
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

/// Layout row — เหมือน Flutter Row
class FormLayoutRow {
  String id;
  List<FormLayoutCell> cells;
  double gap; // horizontal gap between cells
  double paddingTop;
  double paddingBottom;
  double paddingLeft;
  double paddingRight;

  FormLayoutRow({
    String? id,
    List<FormLayoutCell>? cells,
    this.gap = 4,
    this.paddingTop = 0,
    this.paddingBottom = 0,
    this.paddingLeft = 0,
    this.paddingRight = 0,
  })  : id = id ?? DateTime.now().microsecondsSinceEpoch.toString(),
        cells = cells ?? [];

  FormLayoutRow copyWith({
    String? id,
    List<FormLayoutCell>? cells,
    double? gap,
    double? paddingTop,
    double? paddingBottom,
    double? paddingLeft,
    double? paddingRight,
  }) {
    return FormLayoutRow(
      id: id ?? this.id,
      cells: cells ?? this.cells,
      gap: gap ?? this.gap,
      paddingTop: paddingTop ?? this.paddingTop,
      paddingBottom: paddingBottom ?? this.paddingBottom,
      paddingLeft: paddingLeft ?? this.paddingLeft,
      paddingRight: paddingRight ?? this.paddingRight,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'cells': cells.map((c) => c.toJson()).toList(),
      if (gap != 4) 'gap': gap,
      if (paddingTop != 0) 'paddingTop': paddingTop,
      if (paddingBottom != 0) 'paddingBottom': paddingBottom,
      if (paddingLeft != 0) 'paddingLeft': paddingLeft,
      if (paddingRight != 0) 'paddingRight': paddingRight,
    };
  }

  factory FormLayoutRow.fromJson(Map<String, dynamic> json) {
    return FormLayoutRow(
      id: json['id'] as String? ??
          DateTime.now().microsecondsSinceEpoch.toString(),
      cells: json['cells'] != null
          ? (json['cells'] as List)
              .map((c) => FormLayoutCell.fromJson(c as Map<String, dynamic>))
              .toList()
          : [],
      gap: (json['gap'] as num?)?.toDouble() ?? 4,
      paddingTop: (json['paddingTop'] as num?)?.toDouble() ?? 0,
      paddingBottom: (json['paddingBottom'] as num?)?.toDouble() ?? 0,
      paddingLeft: (json['paddingLeft'] as num?)?.toDouble() ?? 0,
      paddingRight: (json['paddingRight'] as num?)?.toDouble() ?? 0,
    );
  }
}

/// Layout resolver — แปลง rows → positioned FormElements
/// เหมือน Flutter layout engine: วัดขนาด → จัด layout → กำหนดตำแหน่ง
class FormLayoutResolver {
  const FormLayoutResolver._();

  /// Resolve rows into positioned FormElements for rendering
  /// [sectionWidth] = ความกว้างของ section (effectiveWidth ของกระดาษ)
  /// Returns: (elements, totalHeight)
  static ({List<FormElement> elements, double height}) resolve(
    List<FormLayoutRow> rows,
    double sectionWidth,
  ) {
    final List<FormElement> result = [];
    double cursorY = 0;

    for (final row in rows) {
      cursorY += row.paddingTop;
      final usableWidth =
          sectionWidth - row.paddingLeft - row.paddingRight;

      // ── 1. คำนวณ cell widths (flex algorithm เหมือน Flutter) ──
      double fixedTotal = 0;
      double flexTotal = 0;
      int gapCount = 0;
      for (final cell in row.cells) {
        if (cell.useFixedWidth) {
          fixedTotal += cell.fixedWidth!;
        } else {
          flexTotal += cell.flex;
        }
        gapCount++;
      }
      gapCount = (gapCount - 1).clamp(0, 9999);
      final gapTotal = gapCount * row.gap;
      final availableForFlex = (usableWidth - fixedTotal - gapTotal)
          .clamp(0.0, double.infinity);

      // ── 2. คำนวณ cell widths + row height ──
      final cellWidths = <double>[];
      double rowHeight = 0;
      for (final cell in row.cells) {
        final w = cell.useFixedWidth
            ? cell.fixedWidth!
            : (flexTotal > 0
                ? availableForFlex * cell.flex / flexTotal
                : availableForFlex);
        cellWidths.add(w);
        final h = cell.estimateHeight(w);
        if (h > rowHeight) rowHeight = h;
      }

      // ── 3. สร้าง FormElement สำหรับแต่ละ cell ──
      double cursorX = row.paddingLeft;
      for (int i = 0; i < row.cells.length; i++) {
        final cell = row.cells[i];
        final w = cellWidths[i];

        result.add(FormElement(
          id: cell.id,
          type: cell.type,
          x: cursorX,
          y: cursorY,
          width: w,
          height: rowHeight,
          text: cell.text,
          fontFamily: cell.fontFamily,
          fontSize: cell.fontSize,
          fontBold: cell.fontBold,
          fontItalic: cell.fontItalic,
          fontUnderline: cell.fontUnderline,
          textColor: cell.textColor,
          textAlign: cell.textAlign,
          imagePath: cell.imagePath,
          backgroundColor: cell.backgroundColor,
          borderColor: cell.borderColor,
          borderWidth: cell.borderWidth,
          borderRadius: cell.borderRadius,
          opacity: cell.opacity,
          dataBindingKey: cell.dataBindingKey,
          dataBindingFormat: cell.dataBindingFormat,
          barcodeFormat: cell.barcodeFormat,
          barcodeValue: cell.barcodeValue,
          lineStyle: cell.lineStyle,
          rows: cell.rows,
          columns: cell.columns,
          tableData: cell.tableData,
          columnAligns: cell.columnAligns,
          tableHasHeader: cell.tableHasHeader,
        ));

        cursorX += w + row.gap;
      }

      cursorY += rowHeight + row.paddingBottom;
    }

    return (elements: result, height: cursorY);
  }
}
