// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'stock_movment_summary_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

StockMovmentSummaryModel _$StockMovmentSummaryModelFromJson(
  Map<String, dynamic> json,
) => StockMovmentSummaryModel(
  barcode: json['barcode'] as String?,
  fromdate: json['fromdate'] as String?,
  todate: json['todate'] as String?,
  totalRecords: (json['total_records'] as num?)?.toInt(),
  totalQtyIn: (json['total_qty_in'] as num?)?.toDouble(),
  averageCostIn: (json['average_cost_in'] as num?)?.toDouble(),
  totalBalanceIn: (json['total_balance_in'] as num?)?.toDouble(),
  totalQtyOut: (json['total_qty_out'] as num?)?.toDouble(),
  averageCostOut: (json['average_cost_out'] as num?)?.toDouble(),
  totalBalanceOut: (json['total_balance_out'] as num?)?.toDouble(),
  finalBalanceQty: (json['final_balance_qty'] as num?)?.toDouble(),
  finalAverageCost: (json['final_average_cost'] as num?)?.toDouble(),
  finalBalanceAmount: (json['final_balance_amount'] as num?)?.toDouble(),
);

Map<String, dynamic> _$StockMovmentSummaryModelToJson(
  StockMovmentSummaryModel instance,
) => <String, dynamic>{
  'barcode': instance.barcode,
  'fromdate': instance.fromdate,
  'todate': instance.todate,
  'total_records': instance.totalRecords,
  'total_qty_in': instance.totalQtyIn,
  'average_cost_in': instance.averageCostIn,
  'total_balance_in': instance.totalBalanceIn,
  'total_qty_out': instance.totalQtyOut,
  'average_cost_out': instance.averageCostOut,
  'total_balance_out': instance.totalBalanceOut,
  'final_balance_qty': instance.finalBalanceQty,
  'final_average_cost': instance.finalAverageCost,
  'final_balance_amount': instance.finalBalanceAmount,
};
