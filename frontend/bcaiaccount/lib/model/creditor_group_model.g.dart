// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'creditor_group_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

CreditorGroupModel _$CreditorGroupModelFromJson(Map<String, dynamic> json) =>
    CreditorGroupModel(
      guidfixed: json['guidfixed'] as String,
      groupcode: json['groupcode'] as String,
      names: (json['names'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$CreditorGroupModelToJson(CreditorGroupModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'groupcode': instance.groupcode,
      'names': instance.names.map((e) => e.toJson()).toList(),
    };
