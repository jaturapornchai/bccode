// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'form_design_struct.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

FormDesignStruct _$FormDesignStructFromJson(Map<String, dynamic> json) =>
    FormDesignStruct(
      guid_fixed: json['guid_fixed'] as String,
      code: json['code'] as String,
      type: (json['type'] as num).toInt(),
      names_json: json['names_json'] as String,
      header_json: json['header_json'] as String,
      detail_json: json['detail_json'] as String,
      detail_extra_json: json['detail_extra_json'] as String,
      detail_total_json: json['detail_total_json'] as String,
      detail_footer_json: json['detail_footer_json'] as String,
      footer_json: json['footer_json'] as String,
    )..id = (json['id'] as num).toInt();

Map<String, dynamic> _$FormDesignStructToJson(FormDesignStruct instance) =>
    <String, dynamic>{
      'id': instance.id,
      'guid_fixed': instance.guid_fixed,
      'code': instance.code,
      'type': instance.type,
      'names_json': instance.names_json,
      'header_json': instance.header_json,
      'detail_json': instance.detail_json,
      'detail_extra_json': instance.detail_extra_json,
      'detail_total_json': instance.detail_total_json,
      'detail_footer_json': instance.detail_footer_json,
      'footer_json': instance.footer_json,
    };

FormDesignHeaderModel _$FormDesignHeaderModelFromJson(
  Map<String, dynamic> json,
) => FormDesignHeaderModel(
  description: (json['description'] as List<dynamic>)
      .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  headerType: (json['headerType'] as num?)?.toInt() ?? 0,
  value: json['value'] as String? ?? "",
  imgUri: json['imgUri'] as String? ?? "",
);

Map<String, dynamic> _$FormDesignHeaderModelToJson(
  FormDesignHeaderModel instance,
) => <String, dynamic>{
  'description': instance.description.map((e) => e.toJson()).toList(),
  'headerType': instance.headerType,
  'value': instance.value,
  'imgUri': instance.imgUri,
};

FormDesignFooterModel _$FormDesignFooterModelFromJson(
  Map<String, dynamic> json,
) => FormDesignFooterModel(
  description: (json['description'] as List<dynamic>)
      .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  footerType: (json['footerType'] as num?)?.toInt() ?? 0,
  value: json['value'] as String? ?? "",
  imgUri: json['imgUri'] as String? ?? "",
  print_qr_doc_no: json['print_qr_doc_no'] as bool,
);

Map<String, dynamic> _$FormDesignFooterModelToJson(
  FormDesignFooterModel instance,
) => <String, dynamic>{
  'description': instance.description.map((e) => e.toJson()).toList(),
  'print_qr_doc_no': instance.print_qr_doc_no,
  'footerType': instance.footerType,
  'value': instance.value,
  'imgUri': instance.imgUri,
};

FormDesignColumnModel _$FormDesignColumnModelFromJson(
  Map<String, dynamic> json,
) => FormDesignColumnModel(
  command: json['command'] as String,
  width: (json['width'] as num?)?.toDouble() ?? 1,
  header_names:
      (json['header_names'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList() ??
      const [],
  font_size: (json['font_size'] as num?)?.toDouble() ?? 12,
  font_family: json['font_family'] as String? ?? "Arial",
  font_weight_bold: json['font_weight_bold'] as bool? ?? false,
  font_style_italic: json['font_style_italic'] as bool? ?? false,
  decoration_underline: json['decoration_underline'] as bool? ?? false,
  text_align:
      $enumDecodeNullable(_$PrintColumnAlignEnumMap, json['text_align']) ??
      PrintColumnAlign.left,
  background_color: json['background_color'] as String? ?? "",
  font_color: json['font_color'] as String? ?? "",
);

Map<String, dynamic> _$FormDesignColumnModelToJson(
  FormDesignColumnModel instance,
) => <String, dynamic>{
  'header_names': instance.header_names.map((e) => e.toJson()).toList(),
  'command': instance.command,
  'width': instance.width,
  'font_size': instance.font_size,
  'font_family': instance.font_family,
  'font_weight_bold': instance.font_weight_bold,
  'font_style_italic': instance.font_style_italic,
  'decoration_underline': instance.decoration_underline,
  'text_align': _$PrintColumnAlignEnumMap[instance.text_align]!,
  'background_color': instance.background_color,
  'font_color': instance.font_color,
};

const _$PrintColumnAlignEnumMap = {
  PrintColumnAlign.left: 'left',
  PrintColumnAlign.right: 'right',
  PrintColumnAlign.center: 'center',
};
