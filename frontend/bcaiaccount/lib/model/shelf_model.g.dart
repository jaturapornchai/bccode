// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'shelf_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ShelfModel _$ShelfModelFromJson(Map<String, dynamic> json) => ShelfModel(
  code: json['code'] as String,
  name: json['name'] as String,
  productitems: (json['productitems'] as List<dynamic>?)
      ?.map((e) => ShelfProductItemModel.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$ShelfModelToJson(ShelfModel instance) =>
    <String, dynamic>{
      'code': instance.code,
      'name': instance.name,
      'productitems': instance.productitems?.map((e) => e.toJson()).toList(),
    };
