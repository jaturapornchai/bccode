// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'product_group_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ProductGroupModel _$ProductGroupModelFromJson(Map<String, dynamic> json) =>
    ProductGroupModel(
      guidfixed: json['guidfixed'] as String,
      code: json['code'] as String,
      names: (json['names'] as List<dynamic>)
          .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$ProductGroupModelToJson(ProductGroupModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'code': instance.code,
      'names': instance.names.map((e) => e.toJson()).toList(),
    };
