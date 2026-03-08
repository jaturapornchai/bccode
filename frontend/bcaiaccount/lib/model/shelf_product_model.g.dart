// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'shelf_product_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ShelfProductBulkAddModel _$ShelfProductBulkAddModelFromJson(
  Map<String, dynamic> json,
) => ShelfProductBulkAddModel(
  products: (json['products'] as List<dynamic>)
      .map((e) => ShelfProductItemModel.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$ShelfProductBulkAddModelToJson(
  ShelfProductBulkAddModel instance,
) => <String, dynamic>{'products': instance.products};

ShelfProductItemModel _$ShelfProductItemModelFromJson(
  Map<String, dynamic> json,
) => ShelfProductItemModel(
  barcode: json['barcode'] as String,
  guidfixed: json['guidfixed'] as String,
  names: (json['names'] as List<dynamic>)
      .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  unitcode: json['unitcode'] as String?,
  unitnames: (json['unitnames'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$ShelfProductItemModelToJson(
  ShelfProductItemModel instance,
) => <String, dynamic>{
  'barcode': instance.barcode,
  'guidfixed': instance.guidfixed,
  'names': instance.names,
  'unitcode': instance.unitcode,
  'unitnames': instance.unitnames,
};

ShelfProductBulkDeleteModel _$ShelfProductBulkDeleteModelFromJson(
  Map<String, dynamic> json,
) => ShelfProductBulkDeleteModel(
  productguidfixedlist: (json['productguidfixedlist'] as List<dynamic>)
      .map((e) => e as String)
      .toList(),
);

Map<String, dynamic> _$ShelfProductBulkDeleteModelToJson(
  ShelfProductBulkDeleteModel instance,
) => <String, dynamic>{'productguidfixedlist': instance.productguidfixedlist};

ShelfProductDisplayModel _$ShelfProductDisplayModelFromJson(
  Map<String, dynamic> json,
) => ShelfProductDisplayModel(
  barcode: json['barcode'] as String,
  guidfixed: json['guidfixed'] as String,
  names: (json['names'] as List<dynamic>)
      .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  isSelected: json['isSelected'] as bool? ?? false,
  unitcode: json['unitcode'] as String,
  unitnames: (json['unitnames'] as List<dynamic>)
      .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$ShelfProductDisplayModelToJson(
  ShelfProductDisplayModel instance,
) => <String, dynamic>{
  'barcode': instance.barcode,
  'guidfixed': instance.guidfixed,
  'names': instance.names,
  'isSelected': instance.isSelected,
  'unitcode': instance.unitcode,
  'unitnames': instance.unitnames,
};
