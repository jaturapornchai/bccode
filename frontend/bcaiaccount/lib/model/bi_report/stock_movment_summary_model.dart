import 'package:json_annotation/json_annotation.dart';

part 'stock_movment_summary_model.g.dart';

@JsonSerializable()
class StockMovmentSummaryModel {
  @JsonKey(name: 'barcode')
  final String? barcode;

  @JsonKey(name: 'fromdate')
  final String? fromdate;

  @JsonKey(name: 'todate')
  final String? todate;

  @JsonKey(name: 'total_records')
  final int? totalRecords;

  @JsonKey(name: 'total_qty_in')
  final double? totalQtyIn;

  @JsonKey(name: 'average_cost_in')
  final double? averageCostIn;

  @JsonKey(name: 'total_balance_in')
  final double? totalBalanceIn;

  @JsonKey(name: 'total_qty_out')
  final double? totalQtyOut;

  @JsonKey(name: 'average_cost_out')
  final double? averageCostOut;

  @JsonKey(name: 'total_balance_out')
  final double? totalBalanceOut;

  @JsonKey(name: 'final_balance_qty')
  final double? finalBalanceQty;

  @JsonKey(name: 'final_average_cost')
  final double? finalAverageCost;

  @JsonKey(name: 'final_balance_amount')
  final double? finalBalanceAmount;

  const StockMovmentSummaryModel({
    String? barcode,
    String? fromdate,
    String? todate,
    int? totalRecords,
    double? totalQtyIn,
    double? averageCostIn,
    double? totalBalanceIn,
    double? totalQtyOut,
    double? averageCostOut,
    double? totalBalanceOut,
    double? finalBalanceQty,
    double? finalAverageCost,
    double? finalBalanceAmount,
  })  : barcode = barcode ?? '',
        fromdate = fromdate ?? '',
        todate = todate ?? '',
        totalRecords = totalRecords ?? 0,
        totalQtyIn = totalQtyIn ?? 0.0,
        averageCostIn = averageCostIn ?? 0.0,
        totalBalanceIn = totalBalanceIn ?? 0.0,
        totalQtyOut = totalQtyOut ?? 0.0,
        averageCostOut = averageCostOut ?? 0.0,
        totalBalanceOut = totalBalanceOut ?? 0.0,
        finalBalanceQty = finalBalanceQty ?? 0.0,
        finalAverageCost = finalAverageCost ?? 0.0,
        finalBalanceAmount = finalBalanceAmount ?? 0.0;

  factory StockMovmentSummaryModel.fromJson(Map<String, dynamic> json) =>
      _$StockMovmentSummaryModelFromJson(json);
  Map<String, dynamic> toJson() => _$StockMovmentSummaryModelToJson(this);
}
