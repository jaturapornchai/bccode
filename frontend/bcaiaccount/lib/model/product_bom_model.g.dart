// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'product_bom_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ProductBomModel _$ProductBomModelFromJson(Map<String, dynamic> json) =>
    ProductBomModel(
      guidfixed: json['guidfixed'] as String?,
      names: (json['names'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      itemunitcode: json['itemunitcode'] as String?,
      itemunitnames: (json['itemunitnames'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      barcode: json['barcode'] as String?,
      condition: json['condition'] as bool?,
      dividevalue: (json['dividevalue'] as num?)?.toInt(),
      standvalue: (json['standvalue'] as num?)?.toInt(),
      qty: (json['qty'] as num?)?.toDouble(),
      imageuri: json['imageuri'] as String?,
      bom: (json['bom'] as List<dynamic>?)
          ?.map((e) => ProductBomModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$ProductBomModelToJson(ProductBomModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'names': instance.names,
      'itemunitcode': instance.itemunitcode,
      'itemunitnames': instance.itemunitnames,
      'barcode': instance.barcode,
      'condition': instance.condition,
      'dividevalue': instance.dividevalue,
      'standvalue': instance.standvalue,
      'qty': instance.qty,
      'imageuri': instance.imageuri,
      'bom': instance.bom,
    };
