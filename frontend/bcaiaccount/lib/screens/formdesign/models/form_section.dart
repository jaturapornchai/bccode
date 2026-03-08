import 'form_element.dart';

enum SectionType {
  header,
  detail,
  footer,
}

class FormSection {
  String id;
  SectionType type;
  double height; // ความสูงของ section
  List<FormElement> elements;
  String? backgroundImagePath;

  FormSection({
    required this.id,
    required this.type,
    this.height = 200,
    List<FormElement>? elements,
    this.backgroundImagePath,
  }) : elements = elements ?? [];

  FormSection copyWith({
    String? id,
    SectionType? type,
    double? height,
    List<FormElement>? elements,
    String? backgroundImagePath,
  }) {
    return FormSection(
      id: id ?? this.id,
      type: type ?? this.type,
      height: height ?? this.height,
      elements: elements ?? this.elements,
      backgroundImagePath: backgroundImagePath ?? this.backgroundImagePath,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'type': type.name,
      'height': height,
      'elements': elements.map((e) => e.toJson()).toList(),
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
      backgroundImagePath: json['backgroundImagePath'] as String?,
    );
  }
}
