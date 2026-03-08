import 'package:json_annotation/json_annotation.dart';

part 'report_main_model.g.dart';

@JsonSerializable(explicitToJson: true)
class ReportSaleByDateModel {
  String docdatetime;
  double detailtotalamount;
  double totaldiscount;
  double totalexceptvat;
  double totalbeforevat;
  double totalvatvalue;
  double totalamount;

  ReportSaleByDateModel({
    String? docdatetime,
    double? detailtotalamount,
    double? totaldiscount,
    double? totalexceptvat,
    double? totalbeforevat,
    double? totalvatvalue,
    double? totalamount,
  })  : docdatetime = docdatetime ?? '',
        detailtotalamount = detailtotalamount ?? 0.0,
        totaldiscount = totaldiscount ?? 0.0,
        totalexceptvat = totalexceptvat ?? 0.0,
        totalbeforevat = totalbeforevat ?? 0.0,
        totalvatvalue = totalvatvalue ?? 0.0,
        totalamount = totalamount ?? 0.0;

  factory ReportSaleByDateModel.fromJson(Map<String, dynamic> json) =>
      _$ReportSaleByDateModelFromJson(json);
  Map<String, dynamic> toJson() => _$ReportSaleByDateModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class ReportReceiveMoneyModel {
  String date;
  ReportReceiveMoneyDetailModel data;

  ReportReceiveMoneyModel({
    String? date,
    ReportReceiveMoneyDetailModel? data,
  })  : date = date ?? '',
        data = data ?? ReportReceiveMoneyDetailModel();

  factory ReportReceiveMoneyModel.fromJson(Map<String, dynamic> json) =>
      _$ReportReceiveMoneyModelFromJson(json);
  Map<String, dynamic> toJson() => _$ReportReceiveMoneyModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class ReportReceiveMoneyDetailModel {
  double cashAmount;
  double creditAmount;
  double transferAmount;
  double couponAmount;
  double chequeAmount;
  double totalAmount;

  ReportReceiveMoneyDetailModel({
    double? cashAmount,
    double? creditAmount,
    double? transferAmount,
    double? couponAmount,
    double? chequeAmount,
    double? totalAmount,
  })  : cashAmount = cashAmount ?? 0.0,
        creditAmount = creditAmount ?? 0.0,
        transferAmount = transferAmount ?? 0.0,
        couponAmount = couponAmount ?? 0.0,
        chequeAmount = chequeAmount ?? 0.0,
        totalAmount = totalAmount ?? 0.0;

  factory ReportReceiveMoneyDetailModel.fromJson(Map<String, dynamic> json) =>
      _$ReportReceiveMoneyDetailModelFromJson(json);
  Map<String, dynamic> toJson() => _$ReportReceiveMoneyDetailModelToJson(this);
}
