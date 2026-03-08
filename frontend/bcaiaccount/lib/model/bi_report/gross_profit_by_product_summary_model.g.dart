// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'gross_profit_by_product_summary_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

GrossProfitByProductSummary _$GrossProfitByProductSummaryFromJson(
  Map<String, dynamic> json,
) => GrossProfitByProductSummary(
  fromdate: json['fromdate'] as String?,
  todate: json['todate'] as String?,
  totalRecords: (json['total_records'] as num?)?.toInt(),
  totalQty: (json['total_qty'] as num?)?.toDouble(),
  totalSumAmountExcludeVat: (json['total_sumamountexcludevat'] as num?)
      ?.toDouble(),
  totalTotalCost: (json['total_totalcost'] as num?)?.toDouble(),
  totalPal: (json['total_pal'] as num?)?.toDouble(),
  companyName: json['companyName'] as String?,
);

Map<String, dynamic> _$GrossProfitByProductSummaryToJson(
  GrossProfitByProductSummary instance,
) => <String, dynamic>{
  'fromdate': instance.fromdate,
  'todate': instance.todate,
  'total_records': instance.totalRecords,
  'total_qty': instance.totalQty,
  'total_sumamountexcludevat': instance.totalSumAmountExcludeVat,
  'total_totalcost': instance.totalTotalCost,
  'total_pal': instance.totalPal,
  'companyName': instance.companyName,
};
