import 'package:json_annotation/json_annotation.dart';
import 'package:smlaicloud/model/global_model.dart';

part 'stock_balance_model.g.dart';

@JsonSerializable()
class StockBalanceModel {
  @JsonKey(name: 'barcode')
  final String barcode;

  @JsonKey(name: 'itemcode', includeIfNull: false)
  final String? itemcode;

  @JsonKey(name: 'mainbarcoderef')
  final String mainBarcodeRef;

  @JsonKey(name: 'names')
  final List<LanguageDataModel> names;

  @JsonKey(name: 'unitnames')
  final List<LanguageDataModel> unitNames;

  @JsonKey(name: 'docdate')
  final String docDate;

  @JsonKey(name: 'docno')
  final String docNo;

  @JsonKey(name: 'balance_qty')
  final double balanceQty;

  @JsonKey(name: 'average_cost')
  final double? averageCost;

  @JsonKey(name: 'balance_amount')
  final double? balanceAmount;

  const StockBalanceModel({
    required this.barcode,
    this.itemcode = '',
    required this.mainBarcodeRef,
    required this.names,
    required this.unitNames,
    required this.docDate,
    required this.docNo,
    required this.balanceQty,
    this.averageCost,
    this.balanceAmount,
  });

  factory StockBalanceModel.fromJson(Map<String, dynamic> json) =>
      _$StockBalanceModelFromJson(json);
  Map<String, dynamic> toJson() => _$StockBalanceModelToJson(this);
}

/// StockBalanceSummaryModel
@JsonSerializable()
class StockBalanceSummaryModel {
  @JsonKey(name: 'todate')
  final String toDate;

  @JsonKey(name: 'barcode')
  final String barcode;

  @JsonKey(name: 'total_records')
  final int totalRecords;

  @JsonKey(name: 'total_balance_qty')
  final double totalBalanceQty;

  @JsonKey(name: 'average_cost')
  final double? averageCost;

  @JsonKey(name: 'total_balance_amount')
  final double? totalBalanceAmount;

  const StockBalanceSummaryModel({
    required this.toDate,
    required this.barcode,
    required this.totalRecords,
    required this.totalBalanceQty,
    this.averageCost,
    this.totalBalanceAmount,
  });

  factory StockBalanceSummaryModel.fromJson(Map<String, dynamic> json) =>
      _$StockBalanceSummaryModelFromJson(json);
  Map<String, dynamic> toJson() => _$StockBalanceSummaryModelToJson(this);
}
