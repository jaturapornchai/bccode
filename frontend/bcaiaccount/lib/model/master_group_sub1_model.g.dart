// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'master_group_sub1_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

MasterGroupSub1Model _$MasterGroupSub1ModelFromJson(
  Map<String, dynamic> json,
) => MasterGroupSub1Model(
  guidfixed: json['guidfixed'] as String,
  code: json['code'] as String,
  names: (json['names'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  groupMainGuid: json['groupMainGuid'] as String,
  groupMainNames: (json['groupMainNames'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$MasterGroupSub1ModelToJson(
  MasterGroupSub1Model instance,
) => <String, dynamic>{
  'guidfixed': instance.guidfixed,
  'code': instance.code,
  'names': instance.names.map((e) => e.toJson()).toList(),
  'groupMainGuid': instance.groupMainGuid,
  'groupMainNames': instance.groupMainNames.map((e) => e.toJson()).toList(),
};
