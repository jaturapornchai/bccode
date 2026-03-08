import 'package:json_annotation/json_annotation.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/product_condition_model.dart';
import '../global.dart' as global;

part 'coupon_model.g.dart';

@JsonSerializable(explicitToJson: true)
class CouponModel {
  String? guidfixed;
  String? couponcode;
  List<LanguageDataModel>? names;
  double? couponvalue;
  String? issueddate;
  String? expirydate;
  int? coupontype; // 0 = ลดตามมูลค่า, 1 = ลดตามเปอร์เซ็นต์, 2 = คูปองแทนเงินสด
  List<String>? customercodes;
  String? remark;
  int? status; // 0 = ใช้งาน, 1 = ยกเลิก
  bool? isonetimeuse;
  int? maxusagecount;
  int? maxusagecountpercustomer;

  @JsonKey(name: 'product_condition')
  ProductConditionModel? productCondition;

  CouponModel({
    this.guidfixed,
    this.couponcode,
    this.names,
    this.couponvalue,
    this.issueddate,
    this.expirydate,
    this.coupontype,
    this.customercodes,
    this.remark,
    this.status,
    this.isonetimeuse,
    this.maxusagecount,
    this.maxusagecountpercustomer,
    this.productCondition,
  }) {
    names ??= <LanguageDataModel>[];
    customercodes ??= <String>[];
  }

  factory CouponModel.fromJson(Map<String, dynamic> json) =>
      _$CouponModelFromJson(json);

  Map<String, dynamic> toJson() => _$CouponModelToJson(this);

  CouponModel copyWith({
    String? guidfixed,
    String? couponcode,
    List<LanguageDataModel>? names,
    double? couponvalue,
    String? issueddate,
    String? expirydate,
    int? coupontype,
    List<String>? customercodes,
    String? remark,
    int? status,
    bool? isonetimeuse,
    int? maxusagecount,
    int? maxusagecountpercustomer,
    ProductConditionModel? productCondition,
  }) {
    return CouponModel(
      guidfixed: guidfixed ?? this.guidfixed,
      couponcode: couponcode ?? this.couponcode,
      names: names ?? this.names,
      couponvalue: couponvalue ?? this.couponvalue,
      issueddate: issueddate ?? this.issueddate,
      expirydate: expirydate ?? this.expirydate,
      coupontype: coupontype ?? this.coupontype,
      customercodes: customercodes ?? this.customercodes,
      remark: remark ?? this.remark,
      status: status ?? this.status,
      isonetimeuse: isonetimeuse ?? this.isonetimeuse,
      maxusagecount: maxusagecount ?? this.maxusagecount,
      maxusagecountpercustomer:
          maxusagecountpercustomer ?? this.maxusagecountpercustomer,
      productCondition: productCondition ?? this.productCondition,
    );
  }

  // Helper methods สำหรับแสดงข้อมูล
  String getCouponTypeText() {
    switch (coupontype) {
      case 0:
        return global.language('reduce_by_value');
      case 1:
        return 'ลดตามเปอร์เซ็นต์';
      case 2:
        return 'คูปองแทนเงินสด';
      default:
        return global.language('alert_not_specified');
    }
  }

  String getStatusText() {
    switch (status) {
      case 0:
        return global.language('active');
      case 1:
        return global.language('cancel');
      default:
        return global.language('alert_not_specified');
    }
  }

  String getCouponValueText() {
    if (couponvalue == null) return '';

    switch (coupontype) {
      case 0:
      case 2:
        return '${couponvalue!.toStringAsFixed(0)} บาท';
      case 1:
        return '${couponvalue!.toStringAsFixed(0)}%';
      default:
        return couponvalue!.toString();
    }
  }
}

@JsonSerializable(explicitToJson: true)
class CouponResponseModel {
  bool? success;
  List<CouponModel>? data;

  CouponResponseModel({this.success, this.data});

  factory CouponResponseModel.fromJson(Map<String, dynamic> json) =>
      _$CouponResponseModelFromJson(json);

  Map<String, dynamic> toJson() => _$CouponResponseModelToJson(this);
}
