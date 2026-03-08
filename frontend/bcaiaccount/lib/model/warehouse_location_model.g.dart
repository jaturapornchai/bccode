// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'warehouse_location_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

WarehouseLocationModel _$WarehouseLocationModelFromJson(
  Map<String, dynamic> json,
) => WarehouseLocationModel(
  guidfixed: json['guidfixed'] as String,
  warehousecode: json['warehousecode'] as String,
  warehousenames: (json['warehousenames'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  locationcode: json['locationcode'] as String,
  locationnames: (json['locationnames'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  shelf: (json['shelf'] as List<dynamic>?)
      ?.map((e) => ShelfModel.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$WarehouseLocationModelToJson(
  WarehouseLocationModel instance,
) => <String, dynamic>{
  'guidfixed': instance.guidfixed,
  'warehousecode': instance.warehousecode,
  'warehousenames': instance.warehousenames,
  'locationcode': instance.locationcode,
  'locationnames': instance.locationnames,
  'shelf': instance.shelf,
};
