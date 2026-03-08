// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'master_group_sub2_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

MasterGroupSub2Model _$MasterGroupSub2ModelFromJson(
  Map<String, dynamic> json,
) => MasterGroupSub2Model(
  guidfixed: json['guidfixed'] as String,
  code: json['code'] as String,
  names: (json['names'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  groupMainGuid: json['groupMainGuid'] as String,
  groupMainNames: (json['groupMainNames'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  groupSubGuid: json['groupSubGuid'] as String,
  groupSubNames: (json['groupSubNames'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$MasterGroupSub2ModelToJson(
  MasterGroupSub2Model instance,
) => <String, dynamic>{
  'guidfixed': instance.guidfixed,
  'code': instance.code,
  'names': instance.names.map((e) => e.toJson()).toList(),
  'groupMainGuid': instance.groupMainGuid,
  'groupMainNames': instance.groupMainNames.map((e) => e.toJson()).toList(),
  'groupSubGuid': instance.groupSubGuid,
  'groupSubNames': instance.groupSubNames.map((e) => e.toJson()).toList(),
};
