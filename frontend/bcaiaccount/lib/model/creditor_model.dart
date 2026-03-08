import 'package:smlaicloud/model/creditor_group_model.dart';
import 'package:smlaicloud/model/customer_address_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'creditor_model.g.dart';

/// ใช้ GET DATA
@JsonSerializable(explicitToJson: true)
class CreditorModel {
  /// ที่อยู่ออกใบกำกับภาษี
  CustomerAddressModel addressforbilling;

  /// ที่อยู่สำหรับจัดส่ง
  List<CustomerAddressModel> addressforshipping;

  /// เลขที่สาขา
  String branchnumber;

  /// รหัสลูกค้า
  String code;

  /// อีเมล
  String email;

  /// กลุ่มเจ้าหนี้
  List<CreditorGroupModel> groups;

  /// รูปภาพ
  List<ImagesModel> images;

  /// ชื่อลูกค้า
  List<LanguageDataModel> names;

  /// 1 = บุคคลธรรมดา 2 = นิติบุคคล
  int personaltype;

  /// 1 = สำนักงานใหญ่ 2 = สาขา
  int customertype;

  /// เลขประจำตัวผู้เสียภาษี
  String taxid;

  //เลขสมาชิกกองทุน
  String fundcode;

  //จำนวนวันเครดิต
  int creditday;

  String guidfixed;

  CreditorModel(
      {CustomerAddressModel? addressforbilling,
      List<CustomerAddressModel>? addressforshipping,
      String? branchnumber,
      String? code,
      String? email,
      List<CreditorGroupModel>? groups,
      List<ImagesModel>? images,
      List<LanguageDataModel>? names,
      int? personaltype,
      int? customertype,
      String? taxid,
      String? guidfixed,
      String? fundcode,
      int? creditday})
      : addressforbilling = addressforbilling ?? CustomerAddressModel(),
        addressforshipping = addressforshipping ?? <CustomerAddressModel>[],
        branchnumber = branchnumber ?? "",
        code = code ?? "",
        email = email ?? "",
        groups = groups ?? <CreditorGroupModel>[],
        images = images ?? <ImagesModel>[],
        names = names ?? <LanguageDataModel>[],
        personaltype = personaltype ?? 0,
        customertype = customertype ?? 0,
        taxid = taxid ?? "",
        guidfixed = guidfixed ?? "",
        fundcode = fundcode ?? "",
        creditday = creditday ?? 0;

  factory CreditorModel.fromJson(Map<String, dynamic> json) =>
      _$CreditorModelFromJson(json);

  Map<String, dynamic> toJson() => _$CreditorModelToJson(this);
}

/// ใช้ SAVE AND UPDATE
@JsonSerializable(explicitToJson: true)
class CreditorRequestModel {
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

  CreditorRequestModel(
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
      String? fundcode,
      int? creditday})
      : addressforshipping = addressforshipping ?? <CustomerAddressModel>[],
        images = images ?? <ImagesModel>[],
        names = names ?? <LanguageDataModel>[],
        groups = groups ?? <String>[],
        fundcode = fundcode ?? "",
        creditday = creditday ?? 0;

  factory CreditorRequestModel.fromJson(Map<String, dynamic> json) =>
      _$CreditorRequestModelFromJson(json);

  Map<String, dynamic> toJson() => _$CreditorRequestModelToJson(this);
}
