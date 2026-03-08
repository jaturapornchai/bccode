// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'color_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ColorModel _$ColorModelFromJson(Map<String, dynamic> json) => ColorModel(
  guidfixed: json['guidfixed'] as String,
  code: json['code'] as String,
  colorselect: json['colorselect'] as String,
  colorsystem: json['colorsystem'] as String,
  colorselecthex: json['colorselecthex'] as String,
  colorsystemhex: json['colorsystemhex'] as String,
  names: (json['names'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$ColorModelToJson(ColorModel instance) =>
    <String, dynamic>{
      'code': instance.code,
      'guidfixed': instance.guidfixed,
      'colorselect': instance.colorselect,
      'colorselecthex': instance.colorselecthex,
      'colorsystem': instance.colorsystem,
      'colorsystemhex': instance.colorsystemhex,
      'names': instance.names.map((e) => e.toJson()).toList(),
    };
