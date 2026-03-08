import 'package:json_annotation/json_annotation.dart';

part 'purchase_partial_summary_model.g.dart';

@JsonSerializable()
class PurchasePartialSummaryModel {
  @JsonKey(name: 'fromdate')
  final String? fromDate;

  @JsonKey(name: 'todate')
  final String? toDate;

  @JsonKey(name: 'total_records')
  final int? totalRecords;

  @JsonKey(name: 'total_value')
  final double? totalValue;

  @JsonKey(name: 'total_except_vat')
  final double? totalExceptVat;

  @JsonKey(name: 'total_before_vat')
  final double? totalBeforeVat;

  @JsonKey(name: 'total_vat_value')
  final double? totalVatValue;

  @JsonKey(name: 'total_amount')
  final double? totalAmount;

  const PurchasePartialSummaryModel({
    this.fromDate,
    this.toDate,
    this.totalRecords,
    this.totalValue,
    this.totalExceptVat,
    this.totalBeforeVat,
    this.totalVatValue,
    this.totalAmount,
  });

  factory PurchasePartialSummaryModel.fromJson(Map<String, dynamic> json) =>
      _$PurchasePartialSummaryModelFromJson(json);

  Map<String, dynamic> toJson() => _$PurchasePartialSummaryModelToJson(this);
}
