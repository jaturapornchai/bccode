// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'promotion_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PromotionModel _$PromotionModelFromJson(
  Map<String, dynamic> json,
) => PromotionModel(
  guidfixed: json['guidfixed'] as String?,
  code: json['code'] as String?,
  promotiontype: (json['promotiontype'] as num?)?.toInt(),
  index: (json['index'] as num?)?.toInt(),
  datebegin: json['datebegin'] as String?,
  dateend: json['dateend'] as String?,
  fromDate: json['fromDate'] == null
      ? null
      : DateTime.parse(json['fromDate'] as String),
  toDate: json['toDate'] == null
      ? null
      : DateTime.parse(json['toDate'] as String),
  name: (json['name'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  discounttext: json['discounttext'] as String?,
  promotionbarcodeinclude: (json['promotionbarcodeinclude'] as List<dynamic>?)
      ?.map(
        (e) => PromotionBarcodeIncludeModel.fromJson(e as Map<String, dynamic>),
      )
      .toList(),
  limitqty: (json['limitqty'] as num?)?.toDouble(),
  promotionqty: (json['promotionqty'] as num?)?.toDouble(),
  limitamount: (json['limitamount'] as num?)?.toDouble(),
  customeronly: (json['customeronly'] as num?)?.toInt(),
);

Map<String, dynamic> _$PromotionModelToJson(PromotionModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'promotiontype': instance.promotiontype,
      'index': instance.index,
      'code': instance.code,
      'datebegin': instance.datebegin,
      'dateend': instance.dateend,
      'fromDate': instance.fromDate.toIso8601String(),
      'toDate': instance.toDate.toIso8601String(),
      'name': instance.name,
      'customeronly': instance.customeronly,
      'discounttext': instance.discounttext,
      'promotionbarcodeinclude': instance.promotionbarcodeinclude,
      'limitqty': instance.limitqty,
      'promotionqty': instance.promotionqty,
      'limitamount': instance.limitamount,
    };

PromotionBarcodeModel _$PromotionBarcodeModelFromJson(
  Map<String, dynamic> json,
) => PromotionBarcodeModel(
  barcode: json['barcode'] as String?,
  name: (json['name'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  unitcode: json['unitcode'] as String?,
  unitname: (json['unitname'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  qty: (json['qty'] as num?)?.toDouble(),
  price: (json['price'] as num?)?.toDouble(),
)..discounttext = json['discounttext'] as String;

Map<String, dynamic> _$PromotionBarcodeModelToJson(
  PromotionBarcodeModel instance,
) => <String, dynamic>{
  'barcode': instance.barcode,
  'name': instance.name,
  'unitcode': instance.unitcode,
  'unitname': instance.unitname,
  'qty': instance.qty,
  'price': instance.price,
  'discounttext': instance.discounttext,
};

PromotionBarcodeIncludeModel _$PromotionBarcodeIncludeModelFromJson(
  Map<String, dynamic> json,
) => PromotionBarcodeIncludeModel(
  promotionproduct: (json['promotionproduct'] as List<dynamic>?)
      ?.map((e) => PromotionBarcodeModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  includeproduct: (json['includeproduct'] as List<dynamic>?)
      ?.map((e) => PromotionBarcodeModel.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$PromotionBarcodeIncludeModelToJson(
  PromotionBarcodeIncludeModel instance,
) => <String, dynamic>{
  'promotionproduct': instance.promotionproduct,
  'includeproduct': instance.includeproduct,
};
