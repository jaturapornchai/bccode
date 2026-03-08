// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'master_class_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

MasterClassModel _$MasterClassModelFromJson(Map<String, dynamic> json) =>
    MasterClassModel(
      guidfixed: json['guidfixed'] as String,
      code: json['code'] as String,
      names: (json['names'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$MasterClassModelToJson(MasterClassModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'code': instance.code,
      'names': instance.names.map((e) => e.toJson()).toList(),
    };
