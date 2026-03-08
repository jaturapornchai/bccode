import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'book_bank_model.g.dart';

@JsonSerializable(explicitToJson: true)
class BookBankModel {
  /// เลขที่บัญชี
  String? passbook;

  /// รหัสสมุดเงินฝาก
  String? bookcode;

  /// รหัสธนาคาร
  String? bankcode;
  String? guidfixed;

  /// รูปภาพสมุดเงินฝาก
  List<ImagesModel>? images;

  /// ชื่อสมุดเงินฝาก
  List<LanguageDataModel>? names;

  /// ชื่อธนาคาร
  List<LanguageDataModel>? banknames;

  /// รหัสผังบัญชี
  String? accountcode;

  /// ชื่อผังบัญชี
  String? accountname;

  /// รหัสสาขา
  String? bankbranch;

  BookBankModel({
    String? passbook,
    String? bookcode,
    String? bankcode,
    String? guidfixed,
    List<ImagesModel>? images,
    List<LanguageDataModel>? names,
    List<LanguageDataModel>? banknames,
    String? accountcode,
    String? accountname,
    String? bankbranch,
  })  : passbook = passbook ?? '',
        bookcode = bookcode ?? '',
        bankcode = bankcode ?? '',
        guidfixed = guidfixed ?? '',
        images = images ?? [],
        names = names ?? [],
        banknames = banknames ?? [],
        accountcode = accountcode ?? '',
        accountname = accountname ?? '',
        bankbranch = bankbranch ?? '';

  factory BookBankModel.fromJson(Map<String, dynamic> json) =>
      _$BookBankModelFromJson(json);

  Map<String, dynamic> toJson() => _$BookBankModelToJson(this);
}
