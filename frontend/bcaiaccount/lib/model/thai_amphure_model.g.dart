// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'thai_amphure_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ThaiAmphureModel _$ThaiAmphureModelFromJson(Map<String, dynamic> json) =>
    ThaiAmphureModel(
      id: (json['id'] as num).toInt(),
      name_th: json['name_th'] as String,
      name_en: json['name_en'] as String,
      province_id: (json['province_id'] as num).toInt(),
    );

Map<String, dynamic> _$ThaiAmphureModelToJson(ThaiAmphureModel instance) =>
    <String, dynamic>{
      'id': instance.id,
      'name_th': instance.name_th,
      'name_en': instance.name_en,
      'province_id': instance.province_id,
    };
