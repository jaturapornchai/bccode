// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'warehouse_location_update_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

WarehouseLocationUpdateModel _$WarehouseLocationUpdateModelFromJson(
  Map<String, dynamic> json,
) => WarehouseLocationUpdateModel(
  warehousecode: json['warehousecode'] as String,
  locationcode: json['locationcode'] as String,
  locationnames: (json['locationnames'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  shelf: (json['shelf'] as List<dynamic>?)
      ?.map((e) => ShelfModel.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$WarehouseLocationUpdateModelToJson(
  WarehouseLocationUpdateModel instance,
) => <String, dynamic>{
  'warehousecode': instance.warehousecode,
  'locationcode': instance.locationcode,
  'locationnames': instance.locationnames,
  'shelf': instance.shelf,
};
