// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'devices_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

DevicesModel _$DevicesModelFromJson(Map<String, dynamic> json) => DevicesModel(
  guidfixed: json['guidfixed'] as String,
  code: json['code'] as String,
  name1: json['name1'] as String,
  type: (json['type'] as num).toInt(),
);

Map<String, dynamic> _$DevicesModelToJson(DevicesModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'code': instance.code,
      'name1': instance.name1,
      'type': instance.type,
    };
