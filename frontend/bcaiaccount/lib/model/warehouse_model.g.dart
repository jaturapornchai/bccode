// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'warehouse_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

WarehouseModel _$WarehouseModelFromJson(Map<String, dynamic> json) =>
    WarehouseModel(
      guidfixed: json['guidfixed'] as String,
      code: json['code'] as String,
      names: (json['names'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      location: (json['location'] as List<dynamic>?)
          ?.map((e) => LocationModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$WarehouseModelToJson(WarehouseModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'code': instance.code,
      'names': instance.names.map((e) => e.toJson()).toList(),
      'location': instance.location.map((e) => e.toJson()).toList(),
    };
