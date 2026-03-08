import 'package:json_annotation/json_annotation.dart';
import 'package:smlaicloud/model/global_model.dart';

part 'stock_movment_model.g.dart';

@JsonSerializable()
class StockMovementModel {
  @JsonKey(name: 'barcode')
  final String barcode;

  @JsonKey(name: 'barcodedoc')
  final String barcodedoc;

  @JsonKey(name: 'docdate')
  final String docdate;

  @JsonKey(name: 'doctime')
  final String doctime;

  @JsonKey(name: 'trans_flag')
  final String transFlag;

  @JsonKey(name: 'docno')
  final String docno;

  @JsonKey(name: 'wh_code')
  final String whCode;

  @JsonKey(name: 'location_code')
  final String locationCode;

  @JsonKey(name: 'unitnames')
  final List<LanguageDataModel> unitnames;

  @JsonKey(name: 'mainunitnames')
  final List<LanguageDataModel> mainunitnames;

  @JsonKey(name: 'qty_in')
  final double qtyIn;

  @JsonKey(name: 'average_cost_in')
  final double averageCostIn;

  @JsonKey(name: 'balance_in')
  final double balanceIn;

  @JsonKey(name: 'qty_out')
  final double qtyOut;

  @JsonKey(name: 'average_cost_out')
  final double averageCostOut;

  @JsonKey(name: 'balance_out')
  final double balanceOut;

  @JsonKey(name: 'balance_qty')
  final double balanceQty;

  @JsonKey(name: 'average_cost')
  final double averageCost;

  @JsonKey(name: 'balance_amount')
  final double balanceAmount;

  const StockMovementModel({
    required this.barcode,
    required this.barcodedoc,
    required this.docdate,
    required this.doctime,
    required this.transFlag,
    required this.docno,
    required this.whCode,
    required this.locationCode,
    required this.unitnames,
    required this.mainunitnames,
    required this.qtyIn,
    required this.averageCostIn,
    required this.balanceIn,
    required this.qtyOut,
    required this.averageCostOut,
    required this.balanceOut,
    required this.balanceQty,
    required this.averageCost,
    required this.balanceAmount,
  });

  factory StockMovementModel.fromJson(Map<String, dynamic> json) =>
      _$StockMovementModelFromJson(json);
  Map<String, dynamic> toJson() => _$StockMovementModelToJson(this);
}
