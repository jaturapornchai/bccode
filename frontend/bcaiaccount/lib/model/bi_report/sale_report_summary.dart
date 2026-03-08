import 'package:json_annotation/json_annotation.dart';

part 'sale_report_summary.g.dart';

@JsonSerializable(explicitToJson: true)
class SaleReportSummary {
  final String? fromdate;
  final String? todate;
  @JsonKey(name: 'total_records')
  final int? totalRecords;
  @JsonKey(name: 'total_amount')
  final double? totalAmount;
  @JsonKey(name: 'total_value')
  final double? totalValue;
  @JsonKey(name: 'total_after_discount')
  final double? totalAfterDiscount;
  @JsonKey(name: 'total_except_vat')
  final double? totalExceptVat;
  @JsonKey(name: 'total_before_vat')
  final double? totalBeforeVat;
  @JsonKey(name: 'total_vat_value')
  final double? totalVatValue;
  @JsonKey(name: 'total_detail_discount')
  final double? totalDetailDiscount;
  @JsonKey(name: 'total_by_branch')
  final List<SaleReportSummaryByBranch>? totalByBranch;

  const SaleReportSummary({
    this.fromdate,
    this.todate,
    this.totalRecords,
    this.totalAmount,
    this.totalValue,
    this.totalAfterDiscount,
    this.totalExceptVat,
    this.totalBeforeVat,
    this.totalVatValue,
    this.totalDetailDiscount,
    this.totalByBranch,
  });

  factory SaleReportSummary.fromJson(Map<String, dynamic> json) =>
      _$SaleReportSummaryFromJson(json);

  Map<String, dynamic> toJson() => _$SaleReportSummaryToJson(this);
}

@JsonSerializable(explicitToJson: true)
class SaleReportSummaryByBranch {
  final String? branchcode;
  @JsonKey(name: 'total_records')
  final int? totalRecords;
  @JsonKey(name: 'total_amount')
  final double? totalAmount;
  @JsonKey(name: 'total_value')
  final double? totalValue;
  @JsonKey(name: 'total_after_discount')
  final double? totalAfterDiscount;
  @JsonKey(name: 'total_except_vat')
  final double? totalExceptVat;
  @JsonKey(name: 'total_before_vat')
  final double? totalBeforeVat;
  @JsonKey(name: 'total_vat_value')
  final double? totalVatValue;
  @JsonKey(name: 'total_detail_discount')
  final double? totalDetailDiscount;

  const SaleReportSummaryByBranch({
    this.branchcode,
    this.totalRecords,
    this.totalAmount,
    this.totalValue,
    this.totalAfterDiscount,
    this.totalExceptVat,
    this.totalBeforeVat,
    this.totalVatValue,
    this.totalDetailDiscount,
  });

  factory SaleReportSummaryByBranch.fromJson(Map<String, dynamic> json) =>
      _$SaleReportSummaryByBranchFromJson(json);

  Map<String, dynamic> toJson() => _$SaleReportSummaryByBranchToJson(this);
}
