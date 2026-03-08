// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'master_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

WareHouseModel _$WareHouseModelFromJson(Map<String, dynamic> json) =>
    WareHouseModel(
      code: json['code'] as String,
      imageuri: json['imageuri'] as String,
      locatecodes: (json['locatecodes'] as List<dynamic>)
          .map((e) => LocationModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$WareHouseModelToJson(WareHouseModel instance) =>
    <String, dynamic>{
      'code': instance.code,
      'imageuri': instance.imageuri,
      'locatecodes': instance.locatecodes.map((e) => e.toJson()).toList(),
    };

LocationModel _$LocationModelFromJson(Map<String, dynamic> json) =>
    LocationModel(
      code: json['code'] as String,
      imageuri: json['imageuri'] as String,
      width: (json['width'] as num).toDouble(),
      height: (json['height'] as num).toDouble(),
      depth: (json['depth'] as num).toDouble(),
      description: json['description'] as String,
      recommendshelfs: (json['recommendshelfs'] as List<dynamic>)
          .map((e) => e as String)
          .toList(),
    );

Map<String, dynamic> _$LocationModelToJson(LocationModel instance) =>
    <String, dynamic>{
      'code': instance.code,
      'imageuri': instance.imageuri,
      'width': instance.width,
      'height': instance.height,
      'depth': instance.depth,
      'description': instance.description,
      'recommendshelfs': instance.recommendshelfs,
    };

ProductShelfModel _$ProductShelfModelFromJson(Map<String, dynamic> json) =>
    ProductShelfModel(
      code: json['code'] as String,
      imageuri: json['imageuri'] as String,
      width: (json['width'] as num).toDouble(),
      height: (json['height'] as num).toDouble(),
      depth: (json['depth'] as num).toDouble(),
      description: json['description'] as String,
    );

Map<String, dynamic> _$ProductShelfModelToJson(ProductShelfModel instance) =>
    <String, dynamic>{
      'code': instance.code,
      'imageuri': instance.imageuri,
      'width': instance.width,
      'height': instance.height,
      'depth': instance.depth,
      'description': instance.description,
    };

UploadImageModel _$UploadImageModelFromJson(Map<String, dynamic> json) =>
    UploadImageModel(uri: json['uri'] as String);

Map<String, dynamic> _$UploadImageModelToJson(UploadImageModel instance) =>
    <String, dynamic>{'uri': instance.uri};
