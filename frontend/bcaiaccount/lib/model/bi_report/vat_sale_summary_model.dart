import 'package:json_annotation/json_annotation.dart';

part 'vat_sale_summary_model.g.dart';

/// Summary model for VAT Sale Report
/// Contains aggregated totals and metadata for the entire report period
@JsonSerializable()
class VatSaleSummary {
  /// Start date of report period
  final String fromdate;

  /// End date of report period
  final String todate;

  /// Total number of records
  @JsonKey(name: 'total_records')
  final int totalRecords;

  /// Sum of all amounts before VAT
  @JsonKey(name: 'total_totalbeforevat')
  final double totalTotalBeforeVat;

  /// Sum of all VAT values
  @JsonKey(name: 'total_totalvatvalue')
  final double totalTotalVatValue;

  /// Sum of all VAT-exempt amounts
  @JsonKey(name: 'total_totalexceptvat')
  final double totalTotalExceptVat;

  /// Sum of all total amounts
  @JsonKey(name: 'total_totalamount')
  final double totalTotalAmount;

  /// Company name (from storage)
  final String companyName;

  /// User who printed the report (from storage)
  final String printedBy;

  const VatSaleSummary({
    required this.fromdate,
    required this.todate,
    required this.totalRecords,
    required this.totalTotalBeforeVat,
    required this.totalTotalVatValue,
    required this.totalTotalExceptVat,
    required this.totalTotalAmount,
    required this.companyName,
    required this.printedBy,
  });

  factory VatSaleSummary.fromJson(Map<String, dynamic> json) =>
      _$VatSaleSummaryFromJson(json);

  Map<String, dynamic> toJson() => _$VatSaleSummaryToJson(this);
}
