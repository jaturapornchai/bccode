import 'package:smlaicloud/model/customer_address_model.dart';
import 'package:smlaicloud/model/debtor_group_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'debtor_model.g.dart';

/// ใช้ GET DATA
@JsonSerializable(explicitToJson: true)
class DebtorModel {
  /// ที่อยู่ออกใบกำกับภาษี
  CustomerAddressModel addressforbilling;

  /// ที่อยู่สำหรับจัดส่ง
  List<CustomerAddressModel> addressforshipping;

  /// เลขที่สาขา
  String? branchnumber;

  /// รหัสลูกค้า
  String code;

  /// อีเมล
  String? email;

  /// กลุ่มลูกค้า
  List<DebtorGroupModel> groups;

  /// รูปภาพ
  List<ImagesModel> images;

  /// ชื่อลูกค้า
  List<LanguageDataModel> names;

  /// 1 = บุคคลธรรมดา 2 = นิติบุคคล
  int personaltype;

  /// 1 = สำนักงานใหญ่ 2 = สาขา
  int customertype;

  /// เลขประจำตัวผู้เสียภาษี
  String? taxid;

  //เลขสมาชิกกองทุน
  String? fundcode;

  //จำนวนวันเครดิต
  int? creditday;

  String guidfixed;

  bool? ismember;

  //แต้มคงเหลือ
  int? pointbalance;

  //รหัสสะสมแต้ม
  String? pointscode;

  // ระดับราคา
  String? pricelevel;

  DebtorModel({
    required this.addressforbilling,
    List<CustomerAddressModel>? addressforshipping,
    this.branchnumber,
    required this.code,
    this.email,
    List<DebtorGroupModel>? groups,
    List<ImagesModel>? images,
    List<LanguageDataModel>? names,
    required this.personaltype,
    required this.customertype,
    this.taxid,
    required this.guidfixed,
    String? fundcode,
    bool? ismember,
    int? creditday,
    int? pointbalance,
    String? pointscode,
    String? pricelevel,
  })  : addressforshipping = addressforshipping ?? <CustomerAddressModel>[],
        images = images ?? <ImagesModel>[],
        names = names ?? <LanguageDataModel>[],
        groups = groups ?? <DebtorGroupModel>[],
        fundcode = fundcode ?? "",
        creditday = creditday ?? 0,
        ismember = ismember ?? false,
        pointbalance = pointbalance ?? 0,
        pointscode = pointscode ?? "",
        pricelevel = pricelevel ?? "1";

  factory DebtorModel.fromJson(Map<String, dynamic> json) =>
      _$DebtorModelFromJson(json);

  Map<String, dynamic> toJson() => _$DebtorModelToJson(this);
}

/// ใช้ SAVE AND UPDATE
@JsonSerializable(explicitToJson: true)
class DebtorRequestModel {
  CustomerAddressModel addressforbilling;
  List<CustomerAddressModel> addressforshipping;
  String? branchnumber;
  String code;
  String? email;
  List<String> groups;
  List<ImagesModel> images;
  List<LanguageDataModel> names;
  int personaltype;
  int customertype;
  String? taxid;
  String guidfixed;
  String? fundcode;
  int? creditday;
  bool? ismember;
  int? pointbalance;
  String? pointscode;
  String? pricelevel;

  DebtorRequestModel(
      {required this.addressforbilling,
      List<CustomerAddressModel>? addressforshipping,
      this.branchnumber,
      required this.code,
      this.email,
      List<String>? groups,
      List<ImagesModel>? images,
      List<LanguageDataModel>? names,
      required this.personaltype,
      required this.customertype,
      this.taxid,
      required this.guidfixed,
      bool? ismember,
      String? fundcode,
      int? creditday,
      int? pointbalance,
      String? pointscode,
      String? pricelevel})
      : addressforshipping = addressforshipping ?? <CustomerAddressModel>[],
        images = images ?? <ImagesModel>[],
        names = names ?? <LanguageDataModel>[],
        groups = groups ?? <String>[],
        fundcode = fundcode ?? "",
        ismember = ismember ?? false,
        creditday = creditday ?? 0,
        pointbalance = pointbalance ?? 0,
        pointscode = pointscode ?? "",
        pricelevel = pricelevel ?? "1";

  factory DebtorRequestModel.fromJson(Map<String, dynamic> json) =>
      _$DebtorRequestModelFromJson(json);

  Map<String, dynamic> toJson() => _$DebtorRequestModelToJson(this);
}
