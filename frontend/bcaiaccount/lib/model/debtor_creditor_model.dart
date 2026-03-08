import 'package:smlaicloud/model/customer_address_model.dart';
import 'package:smlaicloud/model/debtor_group_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'debtor_creditor_model.g.dart';

@JsonSerializable(explicitToJson: true)
class DebtorCreditorModel {
  /// ที่อยู่ออกใบกำกับภาษี
  CustomerAddressModel addressforbilling;

  /// ที่อยู่สำหรับจัดส่ง
  List<CustomerAddressModel>? addressforshipping;

  /// เลขที่สาขา
  String? branchnumber;

  /// รหัสลูกค้า
  String? code;

  /// อีเมล
  String? email;

  /// กลุ่มลูกค้า
  List<DebtorGroupModel>? groups;

  /// รูปภาพ
  List<ImagesModel>? images;

  /// ชื่อลูกค้า
  List<LanguageDataModel>? names;

  /// 1 = บุคคลธรรมดา 2 = นิติบุคคล
  int? personaltype;

  /// 1 = สำนักงานใหญ่ 2 = สาขา
  int? customertype;

  /// เลขประจำตัวผู้เสียภาษี
  String? taxid;

  //เลขสมาชิกกองทุน
  String? fundcode;

  //จำนวนวันเครดิต
  int? creditday;

  String? guidfixed;

  bool? ismember;

  DebtorCreditorModel({
    required this.addressforbilling,
    List<CustomerAddressModel>? addressforshipping,
    String? branchnumber,
    String? code,
    String? email,
    List<DebtorGroupModel>? groups,
    List<ImagesModel>? images,
    List<LanguageDataModel>? names,
    int? personaltype,
    int? customertype,
    String? taxid,
    String? fundcode,
    int? creditday,
    String? guidfixed,
    bool? ismember,
  })  : addressforshipping = addressforshipping ?? <CustomerAddressModel>[],
        images = images ?? <ImagesModel>[],
        names = names ?? <LanguageDataModel>[],
        groups = groups ?? <DebtorGroupModel>[],
        personaltype = personaltype ?? 0,
        customertype = customertype ?? 0,
        creditday = creditday ?? 0,
        guidfixed = guidfixed ?? '',
        ismember = ismember ?? false,
        code = code ?? '',
        email = email ?? '',
        taxid = taxid ?? '',
        fundcode = fundcode ?? '',
        branchnumber = branchnumber ?? '';

  factory DebtorCreditorModel.fromJson(Map<String, dynamic> json) =>
      _$DebtorCreditorModelFromJson(json);

  Map<String, dynamic> toJson() => _$DebtorCreditorModelToJson(this);
}
