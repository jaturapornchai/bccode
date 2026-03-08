import 'form_section.dart';
import 'paper_size.dart';

class FormTemplate {
  String id;
  String name;
  String description;
  PaperSize paperSize;
  FormSection header;
  FormSection detail;
  FormSection footer;
  String? globalBackgroundImagePath;
  DateTime createdAt;
  DateTime updatedAt;

  FormTemplate({
    String? id,
    required this.name,
    this.description = '',
    this.paperSize = PaperSize.a4,
    FormSection? header,
    FormSection? detail,
    FormSection? footer,
    this.globalBackgroundImagePath,
    DateTime? createdAt,
    DateTime? updatedAt,
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

  FormTemplate copyWith({
    String? id,
    String? name,
    String? description,
    PaperSize? paperSize,
    FormSection? header,
    FormSection? detail,
    FormSection? footer,
    String? globalBackgroundImagePath,
    DateTime? createdAt,
    DateTime? updatedAt,
  }) {
    return FormTemplate(
      id: id ?? this.id,
      name: name ?? this.name,
      description: description ?? this.description,
      paperSize: paperSize ?? this.paperSize,
      header: header ?? this.header,
      detail: detail ?? this.detail,
      footer: footer ?? this.footer,
      globalBackgroundImagePath:
          globalBackgroundImagePath ?? this.globalBackgroundImagePath,
      createdAt: createdAt ?? this.createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
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
      'globalBackgroundImagePath': globalBackgroundImagePath,
      'createdAt': createdAt.toIso8601String(),
      'updatedAt': updatedAt.toIso8601String(),
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
      globalBackgroundImagePath: json['globalBackgroundImagePath'] as String?,
      createdAt: DateTime.parse(json['createdAt'] as String),
      updatedAt: DateTime.parse(json['updatedAt'] as String),
    );
  }

  // คำนวณความสูงรวม
  double get totalHeight => header.height + detail.height + footer.height;
}
