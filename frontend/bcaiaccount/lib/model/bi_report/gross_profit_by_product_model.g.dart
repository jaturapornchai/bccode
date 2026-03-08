// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'gross_profit_by_product_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

GrossProfitByProductModel _$GrossProfitByProductModelFromJson(
  Map<String, dynamic> json,
) => GrossProfitByProductModel(
  docdate: json['docdate'] as String?,
  docno: json['docno'] as String?,
  barcode: json['barcode'] as String?,
  mainbarcoderef: json['mainbarcoderef'] as String?,
  itemcode: json['itemcode'] as String?,
  names: (json['names'] as List<dynamic>?)
      ?.map((e) => ProductName.fromJson(e as Map<String, dynamic>))
      .toList(),
  unitcode: json['unitcode'] as String?,
  unitnames: (json['unitnames'] as List<dynamic>?)
      ?.map((e) => UnitName.fromJson(e as Map<String, dynamic>))
      .toList(),
  qty: (json['qty'] as num?)?.toDouble(),
  sumamountexcludevat: (json['sumamountexcludevat'] as num?)?.toDouble(),
  totalcost: (json['totalcost'] as num?)?.toDouble(),
  pal: (json['pal'] as num?)?.toDouble(),
  perPal: (json['per_pal'] as num?)?.toDouble(),
  unitcodeDoc: json['unitcode_doc'] as String?,
  unitnamesDoc: (json['unitnames_doc'] as List<dynamic>?)
      ?.map((e) => UnitName.fromJson(e as Map<String, dynamic>))
      .toList(),
  qtySale: (json['qty_sale'] as num?)?.toDouble(),
  price: (json['price'] as num?)?.toDouble(),
  sumamount: (json['sumamount'] as num?)?.toDouble(),
);

Map<String, dynamic> _$GrossProfitByProductModelToJson(
  GrossProfitByProductModel instance,
) => <String, dynamic>{
  'docdate': instance.docdate,
  'docno': instance.docno,
  'barcode': instance.barcode,
  'mainbarcoderef': instance.mainbarcoderef,
  'itemcode': instance.itemcode,
  'names': instance.names,
  'unitcode': instance.unitcode,
  'unitnames': instance.unitnames,
  'qty': instance.qty,
  'sumamountexcludevat': instance.sumamountexcludevat,
  'totalcost': instance.totalcost,
  'pal': instance.pal,
  'per_pal': instance.perPal,
  'unitcode_doc': instance.unitcodeDoc,
  'unitnames_doc': instance.unitnamesDoc,
  'qty_sale': instance.qtySale,
  'price': instance.price,
  'sumamount': instance.sumamount,
};

ProductName _$ProductNameFromJson(Map<String, dynamic> json) => ProductName(
  code: json['code'] as String?,
  name: json['name'] as String?,
  isauto: json['isauto'] as bool?,
  isdelete: json['isdelete'] as bool?,
);

Map<String, dynamic> _$ProductNameToJson(ProductName instance) =>
    <String, dynamic>{
      'code': instance.code,
      'name': instance.name,
      'isauto': instance.isauto,
      'isdelete': instance.isdelete,
    };

UnitName _$UnitNameFromJson(Map<String, dynamic> json) => UnitName(
  code: json['code'] as String?,
  name: json['name'] as String?,
  isauto: json['isauto'] as bool?,
  isdelete: json['isdelete'] as bool?,
);

Map<String, dynamic> _$UnitNameToJson(UnitName instance) => <String, dynamic>{
  'code': instance.code,
  'name': instance.name,
  'isauto': instance.isauto,
  'isdelete': instance.isdelete,
};
