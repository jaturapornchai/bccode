// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'product_type_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ProductTypeModel _$ProductTypeModelFromJson(Map<String, dynamic> json) =>
    ProductTypeModel(
      guidfixed: json['guidfixed'] as String?,
      code: json['code'] as String?,
      names: (json['names'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$ProductTypeModelToJson(ProductTypeModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'code': instance.code,
      'names': instance.names,
    };
