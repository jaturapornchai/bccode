import 'package:smlaicloud/model/customer_address_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'customer_request_model.g.dart';

@JsonSerializable(explicitToJson: true)
class CustomerRequestModel {
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
  List<String> groups;

  /// รูปภาพ
  List<ImagesModel> images;

  /// เป็นเจ้าหนี้
  bool iscreditor;

  /// เป็นลูกหนี้
  bool isdebtor;

  /// ชื่อลูกค้า
  List<LanguageDataModel> names;

  /// 1 = บุคคลธรรมดา 2 = นิติบุคคล
  int personaltype;

  /// 1 = สำนักงานใหญ่ 2 = สาขา
  int customertype;

  /// เลขประจำตัวผู้เสียภาษี
  String? taxid;
  String guidfixed;

  //เลขสมาชิกกองทุน
  String? fundcode;

  //จำนวนวันเครดิต
  int? creditday;

  CustomerRequestModel(
      {required this.addressforbilling,
      List<CustomerAddressModel>? addressforshipping,
      this.branchnumber,
      required this.code,
      this.email,
      List<String>? groups,
      List<ImagesModel>? images,
      required this.iscreditor,
      required this.isdebtor,
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

  factory CustomerRequestModel.fromJson(Map<String, dynamic> json) =>
      _$CustomerRequestModelFromJson(json);

  Map<String, dynamic> toJson() => _$CustomerRequestModelToJson(this);
}
