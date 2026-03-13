import 'form_element.dart';
import 'form_layout.dart';

enum SectionType {
  header,
  detail,
  footer,
}

class FormSection {
  String id;
  SectionType type;
  double height; // ความสูงของ section
  List<FormElement> elements; // legacy pixel-based elements (for canvas editor)
  List<FormLayoutRow> rows; // row-based layout (primary format)
  String? backgroundImagePath;

  FormSection({
    required this.id,
    required this.type,
    this.height = 200,
    List<FormElement>? elements,
    List<FormLayoutRow>? rows,
    this.backgroundImagePath,
  })  : elements = elements ?? [],
        rows = rows ?? [];

  /// มี row-based layout หรือไม่
  bool get hasRows => rows.isNotEmpty;

  /// Resolve rows → elements สำหรับ rendering ที่ width กำหนด
  List<FormElement> resolveElements(double sectionWidth) {
    if (!hasRows) return elements;
    final result = FormLayoutResolver.resolve(rows, sectionWidth);
    return result.elements;
  }

  /// คำนวณ height จาก rows (auto)
  double resolveHeight(double sectionWidth) {
    if (!hasRows) return height;
    final result = FormLayoutResolver.resolve(rows, sectionWidth);
    return result.height.clamp(20.0, double.infinity);
  }

  FormSection copyWith({
    String? id,
    SectionType? type,
    double? height,
    List<FormElement>? elements,
    List<FormLayoutRow>? rows,
    String? backgroundImagePath,
  }) {
    return FormSection(
      id: id ?? this.id,
      type: type ?? this.type,
      height: height ?? this.height,
      elements: elements ?? this.elements,
      rows: rows ?? this.rows,
      backgroundImagePath: backgroundImagePath ?? this.backgroundImagePath,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'type': type.name,
      'height': height,
      if (elements.isNotEmpty)
        'elements': elements.map((e) => e.toJson()).toList(),
      if (rows.isNotEmpty) 'rows': rows.map((r) => r.toJson()).toList(),
      'backgroundImagePath': backgroundImagePath,
    };
  }

  factory FormSection.fromJson(Map<String, dynamic> json) {
    return FormSection(
      id: json['id'] as String,
      type: SectionType.values.firstWhere(
        (e) => e.name == json['type'],
        orElse: () => SectionType.detail,
      ),
      height: (json['height'] as num?)?.toDouble() ?? 200,
      elements: json['elements'] != null
          ? (json['elements'] as List)
              .map((e) => FormElement.fromJson(e as Map<String, dynamic>))
              .toList()
          : [],
      rows: json['rows'] != null
          ? (json['rows'] as List)
              .map((r) => FormLayoutRow.fromJson(r as Map<String, dynamic>))
              .toList()
          : [],
      backgroundImagePath: json['backgroundImagePath'] as String?,
    );
  }
}
