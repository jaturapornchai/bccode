// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'accountbook_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

AccountBookModel _$AccountBookModelFromJson(Map<String, dynamic> json) =>
    AccountBookModel(
      guidfixed: json['guidfixed'] as String?,
      code: json['code'] as String?,
      name1: json['name1'] as String?,
      name2: json['name2'] as String?,
      name3: json['name3'] as String?,
      name4: json['name4'] as String?,
      name5: json['name5'] as String?,
      iscenterbook: json['iscenterbook'] as bool?,
    );

Map<String, dynamic> _$AccountBookModelToJson(AccountBookModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'code': instance.code,
      'name1': instance.name1,
      'name2': instance.name2,
      'name3': instance.name3,
      'name4': instance.name4,
      'name5': instance.name5,
      'iscenterbook': instance.iscenterbook,
    };
