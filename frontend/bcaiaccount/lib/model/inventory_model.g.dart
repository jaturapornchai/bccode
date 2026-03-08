// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'inventory_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

InventoryModel _$InventoryModelFromJson(Map<String, dynamic> json) =>
    InventoryModel(
      groupcode: json['groupcode'] as String?,
      guidfixed: json['guidfixed'] as String?,
      itemguid: json['itemguid'] as String?,
      barcode: json['barcode'] as String,
      name1: json['name1'] as String,
      name2: json['name2'] as String?,
      name3: json['name3'] as String?,
      name4: json['name4'] as String?,
      name5: json['name5'] as String?,
      description1: json['description1'] as String?,
      description2: json['description2'] as String?,
      description3: json['description3'] as String?,
      description4: json['description4'] as String?,
      description5: json['description5'] as String?,
      itemcode: json['itemcode'] as String?,
      itemunitcode: json['itemunitcode'] as String?,
      itemunitstd: (json['itemunitstd'] as num?)?.toDouble() ?? 0.0,
      itemunitdiv: (json['itemunitdiv'] as num?)?.toDouble() ?? 0.0,
      unitname1: json['unitname1'] as String?,
      unitname2: json['unitname2'] as String?,
      unitname3: json['unitname3'] as String?,
      unitname4: json['unitname4'] as String?,
      unitname5: json['unitname5'] as String?,
      price: (json['price'] as num?)?.toDouble() ?? 0.0,
      options: (json['options'] as List<dynamic>?)
          ?.map((e) => OptionModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      images: (json['images'] as List<dynamic>?)
          ?.map((e) => ImageUpload.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$InventoryModelToJson(InventoryModel instance) =>
    <String, dynamic>{
      'groupcode': instance.groupcode,
      'guidfixed': instance.guidfixed,
      'itemguid': instance.itemguid,
      'barcode': instance.barcode,
      'name1': instance.name1,
      'name2': instance.name2,
      'name3': instance.name3,
      'name4': instance.name4,
      'name5': instance.name5,
      'description1': instance.description1,
      'description2': instance.description2,
      'description3': instance.description3,
      'description4': instance.description4,
      'description5': instance.description5,
      'itemcode': instance.itemcode,
      'itemunitcode': instance.itemunitcode,
      'itemunitstd': instance.itemunitstd,
      'itemunitdiv': instance.itemunitdiv,
      'unitname1': instance.unitname1,
      'unitname2': instance.unitname2,
      'unitname3': instance.unitname3,
      'unitname4': instance.unitname4,
      'unitname5': instance.unitname5,
      'price': instance.price,
      'options': instance.options,
      'images': instance.images,
    };
