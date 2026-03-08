// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'master_grade_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

MasterGradeModel _$MasterGradeModelFromJson(Map<String, dynamic> json) =>
    MasterGradeModel(
      guidfixed: json['guidfixed'] as String,
      code: json['code'] as String,
      names: (json['names'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$MasterGradeModelToJson(MasterGradeModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'code': instance.code,
      'names': instance.names.map((e) => e.toJson()).toList(),
    };
