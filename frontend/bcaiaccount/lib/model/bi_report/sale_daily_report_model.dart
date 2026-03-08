import 'package:json_annotation/json_annotation.dart';
import 'package:smlaicloud/model/bi_report/sale_daily_report_summary.dart';

part 'sale_daily_report_model.g.dart';

/// Model สำหรับข้อมูลรายการแต่ละ transaction ในวันนั้น
@JsonSerializable(explicitToJson: true)
class SaleDailyTransaction {
  @JsonKey(name: 'doc_date')
  final String docDate;
  @JsonKey(name: 'docno')
  final String docno;
  @JsonKey(name: 'total_value')
  final double totalValue;
  @JsonKey(name: 'detail_total_discount')
  final double detailTotalDiscount;
  @JsonKey(name: 'total_after_discount')
  final double totalAfterDiscount;
  @JsonKey(name: 'total_except_vat')
  final double totalExceptVat;
  @JsonKey(name: 'total_before_vat')
  final double totalBeforeVat;
  @JsonKey(name: 'total_vat_value')
  final double totalVatValue;
  @JsonKey(name: 'detail_total_amount')
  final double detailTotalAmount;
  @JsonKey(name: 'total_discount')
  final double totalDiscount;
  @JsonKey(name: 'total_amount')
  final double totalAmount;

  const SaleDailyTransaction({
    required this.docDate,
    required this.docno,
    required this.totalValue,
    required this.detailTotalDiscount,
    required this.totalAfterDiscount,
    required this.totalExceptVat,
    required this.totalBeforeVat,
    required this.totalVatValue,
    required this.detailTotalAmount,
    required this.totalDiscount,
    required this.totalAmount,
  });

  factory SaleDailyTransaction.fromJson(Map<String, dynamic> json) =>
      _$SaleDailyTransactionFromJson(json);

  Map<String, dynamic> toJson() => _$SaleDailyTransactionToJson(this);
}

/// Model สำหรับข้อมูลรายงานยอดขายรายวัน
/// ใช้สำหรับแสดงผลยอดขายแบบจัดกลุ่มตามวันที่
@JsonSerializable(explicitToJson: true)
class SaleDailyReportData {
  @JsonKey(name: 'doc_date')
  final String docDate;
  @JsonKey(name: 'total_value')
  final double totalValue;
  @JsonKey(name: 'detail_total_discount')
  final double detailTotalDiscount;
  @JsonKey(name: 'total_after_discount')
  final double totalAfterDiscount;
  @JsonKey(name: 'total_except_vat')
  final double totalExceptVat;
  @JsonKey(name: 'total_before_vat')
  final double totalBeforeVat;
  @JsonKey(name: 'total_vat_value')
  final double totalVatValue;
  @JsonKey(name: 'detail_total_amount')
  final double detailTotalAmount;
  @JsonKey(name: 'total_discount')
  final double totalDiscount;
  @JsonKey(name: 'total_amount')
  final double totalAmount;
  @JsonKey(name: 'transactions', includeIfNull: false)
  final List<SaleDailyTransaction>? transactions;

  const SaleDailyReportData({
    required this.docDate,
    required this.totalValue,
    required this.detailTotalDiscount,
    required this.totalAfterDiscount,
    required this.totalExceptVat,
    required this.totalBeforeVat,
    required this.totalVatValue,
    required this.detailTotalAmount,
    required this.totalDiscount,
    required this.totalAmount,
    this.transactions = const [],
  });

  factory SaleDailyReportData.fromJson(Map<String, dynamic> json) =>
      _$SaleDailyReportDataFromJson(json);

  Map<String, dynamic> toJson() => _$SaleDailyReportDataToJson(this);
}

/// Response wrapper สำหรับ Summary API
@JsonSerializable(explicitToJson: true)
class SaleDailyReportSummaryResponse {
  final String status;
  final SaleDailyReportSummary data;

  const SaleDailyReportSummaryResponse({
    required this.status,
    required this.data,
  });

  factory SaleDailyReportSummaryResponse.fromJson(Map<String, dynamic> json) =>
      _$SaleDailyReportSummaryResponseFromJson(json);

  Map<String, dynamic> toJson() => _$SaleDailyReportSummaryResponseToJson(this);
}
