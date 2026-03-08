import 'package:json_annotation/json_annotation.dart';

part 'journal_model.g.dart'; // This will be generated later

@JsonSerializable()
class JournalModel {
  String? accountdescription;
  String? accountgroup;
  int? accountperiod;
  int? accountyear;
  double? amount;
  String? batchid;
  String? bookcode;
  String? docdate;
  String? docformat;
  String? docno;
  String? documentref;
  String? exdocrefdate;
  String? exdocrefno;
  int? journaltype;
  List<JournalDetailModel>? journaldetail;
  List<TaxesModel>? taxes;
  List<VatsModel>? vats;
  JournalModel({
    String? accountdescription,
    String? accountgroup,
    int? accountperiod,
    int? accountyear,
    double? amount,
    String? batchid,
    String? bookcode,
    String? docdate,
    String? docformat,
    String? docno,
    String? documentref,
    String? exdocrefdate,
    String? exdocrefno,
    int? journaltype,
    List<JournalDetailModel>? journaldetail,
    List<TaxesModel>? taxes,
    List<VatsModel>? vats,
  })  : accountdescription = accountdescription ?? '',
        accountgroup = accountgroup ?? '',
        accountperiod = accountperiod ?? 0,
        accountyear = accountyear ?? 0,
        amount = amount ?? 0,
        batchid = batchid ?? '',
        bookcode = bookcode ?? '',
        docdate = docdate ?? '',
        docformat = docformat ?? '',
        docno = docno ?? '',
        documentref = documentref ?? '',
        exdocrefdate = exdocrefdate ?? '',
        exdocrefno = exdocrefno ?? '',
        journaltype = journaltype ?? 0,
        journaldetail = journaldetail ?? [],
        taxes = taxes ?? [],
        vats = vats ?? [];

  factory JournalModel.fromJson(Map<String, dynamic> json) =>
      _$JournalModelFromJson(json);

  Map<String, dynamic> toJson() => _$JournalModelToJson(this);
}

@JsonSerializable()
class JournalDetailModel {
  String? accountcode;
  String? accountname;
  double? creditamount;
  double? debitamount;

  JournalDetailModel({
    String? accountcode,
    String? accountname,
    double? creditamount,
    double? debitamount,
  })  : accountcode = accountcode ?? '',
        accountname = accountname ?? '',
        creditamount = creditamount ?? 0,
        debitamount = debitamount ?? 0;

  factory JournalDetailModel.fromJson(Map<String, dynamic> json) =>
      _$JournalDetailModelFromJson(json);

  Map<String, dynamic> toJson() => _$JournalDetailModelToJson(this);
}

@JsonSerializable()
class TaxesModel {
  String? address;
  String? branchcode;
  String? custname;
  String? custtaxid;
  int? custtype;
  List<DetailsModelModel>? details;
  int? organization;
  double? taxamount;
  String? taxdate;
  String? taxdocno;
  int? taxtype;

  TaxesModel({
    String? address,
    String? branchcode,
    String? custname,
    String? custtaxid,
    int? custtype,
    List<DetailsModelModel>? details,
    int? organization,
    double? taxamount,
    String? taxdate,
    String? taxdocno,
    int? taxtype,
  })  : address = address ?? '',
        branchcode = branchcode ?? '',
        custname = custname ?? '',
        custtaxid = custtaxid ?? '',
        custtype = custtype ?? 0,
        details = details ?? [],
        organization = organization ?? 0,
        taxamount = taxamount ?? 0,
        taxdate = taxdate ?? '',
        taxdocno = taxdocno ?? '',
        taxtype = taxtype ?? 0;

  factory TaxesModel.fromJson(Map<String, dynamic> json) =>
      _$TaxesModelFromJson(json);

  Map<String, dynamic> toJson() => _$TaxesModelToJson(this);
}

@JsonSerializable()
class DetailsModelModel {
  String? description;
  double? taxamount;
  double? taxbase;
  double? taxrate;

  DetailsModelModel({
    String? description,
    double? taxamount,
    double? taxbase,
    double? taxrate,
  })  : description = description ?? '',
        taxamount = taxamount ?? 0,
        taxbase = taxbase ?? 0,
        taxrate = taxrate ?? 0;

  factory DetailsModelModel.fromJson(Map<String, dynamic> json) =>
      _$DetailsModelModelFromJson(json);

  Map<String, dynamic> toJson() => _$DetailsModelModelToJson(this);
}

@JsonSerializable()
class VatsModel {
  String? address;
  String? branchcode;
  String? custname;
  String? custtaxid;
  int? custtype;
  int? exceptvat;
  int? organization;
  String? remark;
  double? vatamount;
  double? vatbase;
  String? vatdate;
  String? vatdocno;
  int? vatmode;
  int? vatperiod;
  int? vatrate;
  bool? vatsubmit;
  int? vattype;
  int? vatyear;

  VatsModel({
    String? address,
    String? branchcode,
    String? custname,
    String? custtaxid,
    int? custtype,
    int? exceptvat,
    int? organization,
    String? remark,
    double? vatamount,
    double? vatbase,
    String? vatdate,
    String? vatdocno,
    int? vatmode,
    int? vatperiod,
    int? vatrate,
    bool? vatsubmit,
    int? vattype,
    int? vatyear,
  })  : address = address ?? '',
        branchcode = branchcode ?? '',
        custname = custname ?? '',
        custtaxid = custtaxid ?? '',
        custtype = custtype ?? 0,
        exceptvat = exceptvat ?? 0,
        organization = organization ?? 0,
        remark = remark ?? '',
        vatamount = vatamount ?? 0,
        vatbase = vatbase ?? 0,
        vatdate = vatdate ?? '',
        vatdocno = vatdocno ?? '',
        vatmode = vatmode ?? 0,
        vatperiod = vatperiod ?? 0,
        vatrate = vatrate ?? 0,
        vatsubmit = vatsubmit ?? false,
        vattype = vattype ?? 0,
        vatyear = vatyear ?? 0;

  factory VatsModel.fromJson(Map<String, dynamic> json) =>
      _$VatsModelFromJson(json);

  Map<String, dynamic> toJson() => _$VatsModelToJson(this);
}
