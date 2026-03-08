// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'thai_province_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ThaiProvinceModel _$ThaiProvinceModelFromJson(Map<String, dynamic> json) =>
    ThaiProvinceModel(
      id: (json['id'] as num).toInt(),
      name_th: json['name_th'] as String,
      name_en: json['name_en'] as String,
      geography_id: (json['geography_id'] as num).toInt(),
    );

Map<String, dynamic> _$ThaiProvinceModelToJson(ThaiProvinceModel instance) =>
    <String, dynamic>{
      'id': instance.id,
      'name_th': instance.name_th,
      'name_en': instance.name_en,
      'geography_id': instance.geography_id,
    };
