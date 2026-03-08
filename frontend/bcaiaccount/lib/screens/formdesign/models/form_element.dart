import 'package:flutter/material.dart';

enum ElementType {
  text,
  image,
  line,
  rectangle,
  table,
}

class FormElement {
  String id;
  ElementType type;
  double x; // ตำแหน่ง X
  double y; // ตำแหน่ง Y
  double width;
  double height;
  String? text;
  String? imagePath;
  TextStyle? textStyle;
  Color? backgroundColor;
  Color? borderColor;
  double? borderWidth;
  TextAlign? textAlign;

  // สำหรับตาราง
  int? rows;
  int? columns;
  List<List<String>>? tableData;

  FormElement({
    required this.id,
    required this.type,
    this.x = 0,
    this.y = 0,
    this.width = 100,
    this.height = 50,
    this.text,
    this.imagePath,
    this.textStyle,
    this.backgroundColor,
    this.borderColor,
    this.borderWidth,
    this.textAlign,
    this.rows,
    this.columns,
    this.tableData,
  });

  FormElement copyWith({
    String? id,
    ElementType? type,
    double? x,
    double? y,
    double? width,
    double? height,
    String? text,
    String? imagePath,
    TextStyle? textStyle,
    Color? backgroundColor,
    Color? borderColor,
    double? borderWidth,
    TextAlign? textAlign,
    int? rows,
    int? columns,
    List<List<String>>? tableData,
  }) {
    return FormElement(
      id: id ?? this.id,
      type: type ?? this.type,
      x: x ?? this.x,
      y: y ?? this.y,
      width: width ?? this.width,
      height: height ?? this.height,
      text: text ?? this.text,
      imagePath: imagePath ?? this.imagePath,
      textStyle: textStyle ?? this.textStyle,
      backgroundColor: backgroundColor ?? this.backgroundColor,
      borderColor: borderColor ?? this.borderColor,
      borderWidth: borderWidth ?? this.borderWidth,
      textAlign: textAlign ?? this.textAlign,
      rows: rows ?? this.rows,
      columns: columns ?? this.columns,
      tableData: tableData ?? this.tableData,
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
      'imagePath': imagePath,
      'fontSize': textStyle?.fontSize,
      'fontWeight': textStyle?.fontWeight?.index,
      'color': textStyle?.color?.value,
      'backgroundColor': backgroundColor?.value,
      'borderColor': borderColor?.value,
      'borderWidth': borderWidth,
      'textAlign': textAlign?.index,
      'rows': rows,
      'columns': columns,
      'tableData': tableData,
    };
  }

  factory FormElement.fromJson(Map<String, dynamic> json) {
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
      imagePath: json['imagePath'] as String?,
      textStyle: json['fontSize'] != null
          ? TextStyle(
              fontSize: (json['fontSize'] as num).toDouble(),
              fontWeight: json['fontWeight'] != null
                  ? FontWeight.values[json['fontWeight'] as int]
                  : null,
              color: json['color'] != null ? Color(json['color'] as int) : null,
            )
          : null,
      backgroundColor: json['backgroundColor'] != null
          ? Color(json['backgroundColor'] as int)
          : null,
      borderColor: json['borderColor'] != null
          ? Color(json['borderColor'] as int)
          : null,
      borderWidth: (json['borderWidth'] as num?)?.toDouble(),
      textAlign: json['textAlign'] != null
          ? TextAlign.values[json['textAlign'] as int]
          : null,
      rows: json['rows'] as int?,
      columns: json['columns'] as int?,
      tableData: json['tableData'] != null
          ? (json['tableData'] as List)
              .map((row) =>
                  (row as List).map((cell) => cell.toString()).toList())
              .toList()
          : null,
    );
  }
}
