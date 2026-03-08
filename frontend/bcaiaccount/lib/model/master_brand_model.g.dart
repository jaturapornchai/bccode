// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'master_brand_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

MasterBrandModel _$MasterBrandModelFromJson(Map<String, dynamic> json) =>
    MasterBrandModel(
      guidfixed: json['guidfixed'] as String,
      code: json['code'] as String,
      names: (json['names'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$MasterBrandModelToJson(MasterBrandModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'code': instance.code,
      'names': instance.names.map((e) => e.toJson()).toList(),
    };
