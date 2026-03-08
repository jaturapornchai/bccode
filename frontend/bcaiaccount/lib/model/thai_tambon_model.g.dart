// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'thai_tambon_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ThaiTambonModel _$ThaiTambonModelFromJson(Map<String, dynamic> json) =>
    ThaiTambonModel(
      id: (json['id'] as num).toInt(),
      name_th: json['name_th'] as String,
      name_en: json['name_en'] as String,
      amphure_id: (json['amphure_id'] as num).toInt(),
      zip_code: (json['zip_code'] as num).toInt(),
    );

Map<String, dynamic> _$ThaiTambonModelToJson(ThaiTambonModel instance) =>
    <String, dynamic>{
      'id': instance.id,
      'name_th': instance.name_th,
      'name_en': instance.name_en,
      'amphure_id': instance.amphure_id,
      'zip_code': instance.zip_code,
    };
