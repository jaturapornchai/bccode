// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'stock_balance_import_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

UploadSuccessModel _$UploadSuccessModelFromJson(Map<String, dynamic> json) =>
    UploadSuccessModel(
      success: json['success'] as bool,
      id: json['id'] as String,
    );

Map<String, dynamic> _$UploadSuccessModelToJson(UploadSuccessModel instance) =>
    <String, dynamic>{'success': instance.success, 'id': instance.id};

StockBalanceImportModel _$StockBalanceImportModelFromJson(
  Map<String, dynamic> json,
) => StockBalanceImportModel(
  guidfixed: json['guidfixed'] as String?,
  shopid: json['shopid'] as String?,
  taskid: json['taskid'] as String?,
  rownumber: (json['rownumber'] as num?)?.toInt(),
  barcode: json['barcode'] as String?,
  name: json['name'] as String?,
  unitcode: json['unitcode'] as String?,
  warehousecode: json['warehousecode'] as String?,
  shelfcode: json['shelfcode'] as String?,
  qty: (json['qty'] as num?)?.toDouble(),
  price: (json['price'] as num?)?.toDouble(),
  sumamount: (json['sumamount'] as num?)?.toDouble(),
  isnotexist: json['isnotexist'] as bool?,
);

Map<String, dynamic> _$StockBalanceImportModelToJson(
  StockBalanceImportModel instance,
) => <String, dynamic>{
  'guidfixed': instance.guidfixed,
  'shopid': instance.shopid,
  'taskid': instance.taskid,
  'rownumber': instance.rownumber,
  'barcode': instance.barcode,
  'name': instance.name,
  'unitcode': instance.unitcode,
  'warehousecode': instance.warehousecode,
  'shelfcode': instance.shelfcode,
  'qty': instance.qty,
  'price': instance.price,
  'sumamount': instance.sumamount,
  'isnotexist': instance.isnotexist,
};

TotalModel _$TotalModelFromJson(Map<String, dynamic> json) => TotalModel(
  totalitem: (json['totalitem'] as num).toDouble(),
  totalamount: (json['totalamount'] as num).toDouble(),
);

Map<String, dynamic> _$TotalModelToJson(TotalModel instance) =>
    <String, dynamic>{
      'totalitem': instance.totalitem,
      'totalamount': instance.totalamount,
    };
