// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'master_group_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

MasterGroupModel _$MasterGroupModelFromJson(Map<String, dynamic> json) =>
    MasterGroupModel(
      guidfixed: json['guidfixed'] as String,
      code: json['code'] as String,
      names: (json['names'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$MasterGroupModelToJson(MasterGroupModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'code': instance.code,
      'names': instance.names.map((e) => e.toJson()).toList(),
    };
