// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'debtor_group_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

DebtorGroupModel _$DebtorGroupModelFromJson(Map<String, dynamic> json) =>
    DebtorGroupModel(
      guidfixed: json['guidfixed'] as String,
      groupcode: json['groupcode'] as String,
      names: (json['names'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$DebtorGroupModelToJson(DebtorGroupModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'groupcode': instance.groupcode,
      'names': instance.names.map((e) => e.toJson()).toList(),
    };
