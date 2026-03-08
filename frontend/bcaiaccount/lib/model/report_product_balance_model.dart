import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'report_product_balance_model.g.dart';

@JsonSerializable(explicitToJson: true)
class ReportProductBalanceModel {
  String? barcode;
  List<LanguageDataModel>? names;
  String? unitcode;
  String? standunit;
  String? balanceqty;
  String? averagecost;
  String? balanceamount;

  ReportProductBalanceModel({
    String? barcode,
    List<LanguageDataModel>? names,
    String? unitcode,
    String? standunit,
    String? balanceqty,
    String? averagecost,
    String? balanceamount,
  })  : barcode = barcode ?? '',
        names = names ?? [],
        unitcode = unitcode ?? '',
        standunit = standunit ?? '',
        balanceqty = balanceqty ?? '',
        averagecost = averagecost ?? '',
        balanceamount = balanceamount ?? '';

  factory ReportProductBalanceModel.fromJson(Map<String, dynamic> json) =>
      _$ReportProductBalanceModelFromJson(json);

  Map<String, dynamic> toJson() => _$ReportProductBalanceModelToJson(this);
}
