import 'package:json_annotation/json_annotation.dart';
import '../global.dart' as global;

part 'shift_detail_model.g.dart';

@JsonSerializable()
class ShiftDetailModel {
  final String? usercode;
  final String? username;
  final String? posid;
  final String? docno;
  final int? doctype;
  final String? docdate;
  final String? remark;
  final double? amount;
  final double? creditcard;
  final double? promptpay;
  final double? transfer;
  final double? cheque;
  final double? coupon;

  const ShiftDetailModel({
    this.usercode,
    this.username,
    this.posid,
    this.docno,
    this.doctype,
    this.docdate,
    this.remark,
    this.amount,
    this.creditcard,
    this.promptpay,
    this.transfer,
    this.cheque,
    this.coupon,
  });

  factory ShiftDetailModel.fromJson(Map<String, dynamic> json) =>
      _$ShiftDetailModelFromJson(json);

  Map<String, dynamic> toJson() => _$ShiftDetailModelToJson(this);

  String get doctypeText {
    switch (doctype) {
      case 1:
        return global.language('open_shift');
      case 2:
        return global.language('close_shift');
      case 3:
        return 'เพิ่มเงิน';
      case 4:
        return global.language('withdraw_cash');
      default:
        return global.language('alert_not_specified');
    }
  }

  bool get isCashIn => doctype == 1 || doctype == 3;
  bool get isCashOut => doctype == 2 || doctype == 4;
}
