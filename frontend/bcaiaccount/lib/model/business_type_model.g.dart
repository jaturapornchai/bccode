// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'business_type_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

BusinessTypeModel _$BusinessTypeModelFromJson(Map<String, dynamic> json) =>
    BusinessTypeModel(
      guidfixed: json['guidfixed'] as String?,
      code: json['code'] as String?,
      names: (json['names'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      isdefault: json['isdefault'] as bool?,
    );

Map<String, dynamic> _$BusinessTypeModelToJson(BusinessTypeModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'code': instance.code,
      'names': instance.names?.map((e) => e.toJson()).toList(),
      'isdefault': instance.isdefault,
    };
