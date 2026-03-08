// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'master_model_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

MasterModelModel _$MasterModelModelFromJson(Map<String, dynamic> json) =>
    MasterModelModel(
      guidfixed: json['guidfixed'] as String,
      code: json['code'] as String,
      names: (json['names'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$MasterModelModelToJson(MasterModelModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'code': instance.code,
      'names': instance.names.map((e) => e.toJson()).toList(),
    };
