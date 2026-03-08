// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'kitchen_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

KitchenModel _$KitchenModelFromJson(Map<String, dynamic> json) => KitchenModel(
  guidfixed: json['guidfixed'] as String?,
  code: json['code'] as String,
  names: (json['names'] as List<dynamic>)
      .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  printers: (json['printers'] as List<dynamic>?)
      ?.map((e) => e as String)
      .toList(),
  products: (json['products'] as List<dynamic>?)
      ?.map((e) => e as String)
      .toList(),
  zones: (json['zones'] as List<dynamic>?)?.map((e) => e as String).toList(),
  groupnumber: (json['groupnumber'] as num?)?.toInt(),
  zonenumber: (json['zonenumber'] as num?)?.toInt(),
);

Map<String, dynamic> _$KitchenModelToJson(KitchenModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'code': instance.code,
      'names': instance.names,
      'printers': instance.printers,
      'products': instance.products,
      'zones': instance.zones,
      'groupnumber': instance.groupnumber,
      'zonenumber': instance.zonenumber,
    };

GetKitchenModel _$GetKitchenModelFromJson(Map<String, dynamic> json) =>
    GetKitchenModel(
      guidfixed: json['guidfixed'] as String,
      code: json['code'] as String,
      names: (json['names'] as List<dynamic>)
          .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      printers: (json['printers'] as List<dynamic>?)
          ?.map((e) => e as String)
          .toList(),
      products: (json['products'] as List<dynamic>?)
          ?.map((e) => KitchenProductModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      zones: (json['zones'] as List<dynamic>?)
          ?.map((e) => e as String)
          .toList(),
    );

Map<String, dynamic> _$GetKitchenModelToJson(GetKitchenModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'code': instance.code,
      'names': instance.names,
      'printers': instance.printers,
      'products': instance.products,
      'zones': instance.zones,
    };
