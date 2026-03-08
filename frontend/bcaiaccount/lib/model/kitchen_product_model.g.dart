// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'kitchen_product_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

KitchenProductModel _$KitchenProductModelFromJson(Map<String, dynamic> json) =>
    KitchenProductModel(
      guidfixed: json['guidfixed'] as String,
      barcode: json['barcode'] as String,
      names: (json['names'] as List<dynamic>)
          .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      isdisable: json['isdisable'] as bool?,
      kitchens: (json['kitchens'] as List<dynamic>?)
          ?.map((e) => KitchenModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$KitchenProductModelToJson(
  KitchenProductModel instance,
) => <String, dynamic>{
  'guidfixed': instance.guidfixed,
  'barcode': instance.barcode,
  'names': instance.names,
  'isdisable': instance.isdisable,
  'kitchens': instance.kitchens,
};

ProductInKitchenModel _$ProductInKitchenModelFromJson(
  Map<String, dynamic> json,
) => ProductInKitchenModel(
  barcode: json['barcode'] as String?,
  kitchens: (json['kitchens'] as List<dynamic>?)
      ?.map((e) => KitchenModel.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$ProductInKitchenModelToJson(
  ProductInKitchenModel instance,
) => <String, dynamic>{
  'barcode': instance.barcode,
  'kitchens': instance.kitchens,
};
