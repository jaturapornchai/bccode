// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'sale_daily_report_summary.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

SaleDailyReportSummary _$SaleDailyReportSummaryFromJson(
  Map<String, dynamic> json,
) => SaleDailyReportSummary(
  fromdate: json['fromdate'] as String?,
  todate: json['todate'] as String?,
  totalDays: (json['total_days'] as num?)?.toInt(),
  totalAmount: (json['total_amount'] as num?)?.toDouble(),
  averageDailyAmount: (json['average_daily_amount'] as num?)?.toDouble(),
);

Map<String, dynamic> _$SaleDailyReportSummaryToJson(
  SaleDailyReportSummary instance,
) => <String, dynamic>{
  'fromdate': instance.fromdate,
  'todate': instance.todate,
  'total_days': instance.totalDays,
  'total_amount': instance.totalAmount,
  'average_daily_amount': instance.averageDailyAmount,
};
