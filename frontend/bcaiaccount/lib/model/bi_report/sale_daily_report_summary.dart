import 'package:json_annotation/json_annotation.dart';

part 'sale_daily_report_summary.g.dart';

/// Model สำหรับสรุปรายงานยอดขายรายวัน
@JsonSerializable(explicitToJson: true)
class SaleDailyReportSummary {
  final String? fromdate;
  final String? todate;
  @JsonKey(name: 'total_days')
  final int? totalDays;
  @JsonKey(name: 'total_amount')
  final double? totalAmount;
  @JsonKey(name: 'average_daily_amount')
  final double? averageDailyAmount;

  const SaleDailyReportSummary({
    this.fromdate,
    this.todate,
    this.totalDays,
    this.totalAmount,
    this.averageDailyAmount,
  });

  factory SaleDailyReportSummary.fromJson(Map<String, dynamic> json) =>
      _$SaleDailyReportSummaryFromJson(json);

  Map<String, dynamic> toJson() => _$SaleDailyReportSummaryToJson(this);
}
