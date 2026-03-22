// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'job_project_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

JobProjectModel _$JobProjectModelFromJson(Map<String, dynamic> json) =>
    JobProjectModel(
      guidfixed: json['guidfixed'] as String?,
      code: json['code'] as String,
      parentcode: json['parentcode'] as String?,
      names: (json['names'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$JobProjectModelToJson(JobProjectModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'code': instance.code,
      'parentcode': instance.parentcode,
      'names': instance.names.map((e) => e.toJson()).toList(),
    };
