// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'bank_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

BankModel _$BankModelFromJson(Map<String, dynamic> json) => BankModel(
  guidfixed: json['guidfixed'] as String?,
  code: json['code'] as String?,
  logo: json['logo'] as String?,
  names: (json['names'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$BankModelToJson(BankModel instance) => <String, dynamic>{
  'code': instance.code,
  'guidfixed': instance.guidfixed,
  'logo': instance.logo,
  'names': instance.names.map((e) => e.toJson()).toList(),
};
