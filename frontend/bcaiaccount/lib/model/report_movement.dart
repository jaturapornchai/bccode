import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'report_movement.g.dart';

@JsonSerializable(explicitToJson: true)
class ReportMovementModel {
  double? balance;
  List<ReportMovementDetail>? details;

  ReportMovementModel({
    double? balance,
    List<ReportMovementDetail>? details,
  })  : balance = balance ?? 0,
        details = details ?? [];

  factory ReportMovementModel.fromJson(Map<String, dynamic> json) =>
      _$ReportMovementModelFromJson(json);

  Map<String, dynamic> toJson() => _$ReportMovementModelToJson(this);
}

@JsonSerializable()
class ReportMovementDetail {
  String docdate;
  String docno;
  String barcode;
  String qty;
  int transflag;
  int calcflag;
  double balance;

  ReportMovementDetail({
    String? docdate,
    String? docno,
    String? barcode,
    String? qty,
    int? transflag,
    int? calcflag,
    List<LanguageDataModel>? names,
    double? balance,
  })  : docdate = docdate ?? "",
        docno = docno ?? "",
        barcode = barcode ?? "",
        qty = qty ?? "0",
        transflag = transflag ?? 0,
        calcflag = calcflag ?? 0,
        balance = balance ?? 0.0;

  factory ReportMovementDetail.fromJson(Map<String, dynamic> json) =>
      _$ReportMovementDetailFromJson(json);
  Map<String, dynamic> toJson() => _$ReportMovementDetailToJson(this);
}
