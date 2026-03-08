// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'promotion_detail_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PromotionDetailModel _$PromotionDetailModelFromJson(
  Map<String, dynamic> json,
) => PromotionDetailModel(
  detailtype: (json['detailtype'] as num).toInt(),
  minimum: (json['minimum'] as num).toDouble(),
  discount: (json['discount'] as num).toDouble(),
  productbarcode: ProductBarcodeModel.fromJson(
    json['productbarcode'] as Map<String, dynamic>,
  ),
);

Map<String, dynamic> _$PromotionDetailModelToJson(
  PromotionDetailModel instance,
) => <String, dynamic>{
  'detailtype': instance.detailtype,
  'minimum': instance.minimum,
  'discount': instance.discount,
  'productbarcode': instance.productbarcode,
};
