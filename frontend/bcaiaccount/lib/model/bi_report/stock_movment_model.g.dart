// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'stock_movment_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

StockMovementModel _$StockMovementModelFromJson(Map<String, dynamic> json) =>
    StockMovementModel(
      barcode: json['barcode'] as String,
      barcodedoc: json['barcodedoc'] as String,
      docdate: json['docdate'] as String,
      doctime: json['doctime'] as String,
      transFlag: json['trans_flag'] as String,
      docno: json['docno'] as String,
      whCode: json['wh_code'] as String,
      locationCode: json['location_code'] as String,
      unitnames: (json['unitnames'] as List<dynamic>)
          .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      mainunitnames: (json['mainunitnames'] as List<dynamic>)
          .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      qtyIn: (json['qty_in'] as num).toDouble(),
      averageCostIn: (json['average_cost_in'] as num).toDouble(),
      balanceIn: (json['balance_in'] as num).toDouble(),
      qtyOut: (json['qty_out'] as num).toDouble(),
      averageCostOut: (json['average_cost_out'] as num).toDouble(),
      balanceOut: (json['balance_out'] as num).toDouble(),
      balanceQty: (json['balance_qty'] as num).toDouble(),
      averageCost: (json['average_cost'] as num).toDouble(),
      balanceAmount: (json['balance_amount'] as num).toDouble(),
    );

Map<String, dynamic> _$StockMovementModelToJson(StockMovementModel instance) =>
    <String, dynamic>{
      'barcode': instance.barcode,
      'barcodedoc': instance.barcodedoc,
      'docdate': instance.docdate,
      'doctime': instance.doctime,
      'trans_flag': instance.transFlag,
      'docno': instance.docno,
      'wh_code': instance.whCode,
      'location_code': instance.locationCode,
      'unitnames': instance.unitnames,
      'mainunitnames': instance.mainunitnames,
      'qty_in': instance.qtyIn,
      'average_cost_in': instance.averageCostIn,
      'balance_in': instance.balanceIn,
      'qty_out': instance.qtyOut,
      'average_cost_out': instance.averageCostOut,
      'balance_out': instance.balanceOut,
      'balance_qty': instance.balanceQty,
      'average_cost': instance.averageCost,
      'balance_amount': instance.balanceAmount,
    };
