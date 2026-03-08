// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'gross_profit_by_document_summary_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

GrossProfitByDocumentSummary _$GrossProfitByDocumentSummaryFromJson(
  Map<String, dynamic> json,
) => GrossProfitByDocumentSummary(
  fromdate: json['fromdate'] as String?,
  todate: json['todate'] as String?,
  totalRecords: (json['total_records'] as num?)?.toInt(),
  totalTotalBeforeVat: (json['total_totalbeforevat'] as num?)?.toDouble(),
  totalDetailTotalDiscount: (json['total_detailtotaldiscount'] as num?)
      ?.toDouble(),
  totalTotalAfterDiscount: (json['total_totalafterdiscount'] as num?)
      ?.toDouble(),
  totalTotalCost: (json['total_totalcost'] as num?)?.toDouble(),
  totalPal: (json['total_pal'] as num?)?.toDouble(),
  companyName: json['companyName'] as String?,
);

Map<String, dynamic> _$GrossProfitByDocumentSummaryToJson(
  GrossProfitByDocumentSummary instance,
) => <String, dynamic>{
  'fromdate': instance.fromdate,
  'todate': instance.todate,
  'total_records': instance.totalRecords,
  'total_totalbeforevat': instance.totalTotalBeforeVat,
  'total_detailtotaldiscount': instance.totalDetailTotalDiscount,
  'total_totalafterdiscount': instance.totalTotalAfterDiscount,
  'total_totalcost': instance.totalTotalCost,
  'total_pal': instance.totalPal,
  'companyName': instance.companyName,
};
