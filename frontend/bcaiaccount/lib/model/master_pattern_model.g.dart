// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'master_pattern_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

MasterPatternModel _$MasterPatternModelFromJson(Map<String, dynamic> json) =>
    MasterPatternModel(
      guidfixed: json['guidfixed'] as String,
      code: json['code'] as String,
      names: (json['names'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$MasterPatternModelToJson(MasterPatternModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'code': instance.code,
      'names': instance.names.map((e) => e.toJson()).toList(),
    };
