// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'stock_balance_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

StockBalanceModel _$StockBalanceModelFromJson(Map<String, dynamic> json) =>
    StockBalanceModel(
      barcode: json['barcode'] as String,
      itemcode: json['itemcode'] as String? ?? '',
      mainBarcodeRef: json['mainbarcoderef'] as String,
      names: (json['names'] as List<dynamic>)
          .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      unitNames: (json['unitnames'] as List<dynamic>)
          .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      docDate: json['docdate'] as String,
      docNo: json['docno'] as String,
      balanceQty: (json['balance_qty'] as num).toDouble(),
      averageCost: (json['average_cost'] as num?)?.toDouble(),
      balanceAmount: (json['balance_amount'] as num?)?.toDouble(),
    );

Map<String, dynamic> _$StockBalanceModelToJson(StockBalanceModel instance) =>
    <String, dynamic>{
      'barcode': instance.barcode,
      'itemcode': ?instance.itemcode,
      'mainbarcoderef': instance.mainBarcodeRef,
      'names': instance.names,
      'unitnames': instance.unitNames,
      'docdate': instance.docDate,
      'docno': instance.docNo,
      'balance_qty': instance.balanceQty,
      'average_cost': instance.averageCost,
      'balance_amount': instance.balanceAmount,
    };

StockBalanceSummaryModel _$StockBalanceSummaryModelFromJson(
  Map<String, dynamic> json,
) => StockBalanceSummaryModel(
  toDate: json['todate'] as String,
  barcode: json['barcode'] as String,
  totalRecords: (json['total_records'] as num).toInt(),
  totalBalanceQty: (json['total_balance_qty'] as num).toDouble(),
  averageCost: (json['average_cost'] as num?)?.toDouble(),
  totalBalanceAmount: (json['total_balance_amount'] as num?)?.toDouble(),
);

Map<String, dynamic> _$StockBalanceSummaryModelToJson(
  StockBalanceSummaryModel instance,
) => <String, dynamic>{
  'todate': instance.toDate,
  'barcode': instance.barcode,
  'total_records': instance.totalRecords,
  'total_balance_qty': instance.totalBalanceQty,
  'average_cost': instance.averageCost,
  'total_balance_amount': instance.totalBalanceAmount,
};
