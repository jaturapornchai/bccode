import 'package:smlaicloud/model/book_bank_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/qr_model.dart';
import 'package:smlaicloud/model/sale_channel_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'pos_setting_model.g.dart';

@JsonSerializable()
class PosSettingModel {
  late String guidfixed;
  late BranchModel branch;

  /// รหัสรเครื่อง
  late String code;

  /// หมายเลขเครื่อง
  late String devicenumber;

  /// รูปแบบเลขที่เอกสารอย่างย่อ
  late String docformatinv;

  /// รูปแบบเลขที่เอกสารอย่างเต็ม
  late String docformattaxinv;

  /// พื้นที่เก็บสินค้า
  late LocationModel location;

  /// รูปแบบใบเสร็จรับเงิน
  late String receiptform;

  /// รหัสคลังสินค้า
  late WarehouseModel warehouse;

  /// รหัสพนักงานที่ใช้งาน
  late String? activepin;

  /// รายชื่อพนักงานที่ใช้งาน
  late List<EmployeePosModel>? employees;

  /// รหัสเอกสาร
  late String? doccode;

  /// รูปแบบเลขที่เอกสารรับคืน
  late String? docformatesalereturn;

  /// ประเภทภาษี
  late int? vattype;

  /// อัตราภาษี
  late double? vatrate;

  /// slips
  late List<SlipModel>? slips;

  /// vatrate
  late bool? isejournal;

  /// qr code
  late List<QrModel>? qrcodes;

  /// เครดิตการ์ด
  late List<CreditcardsModel>? creditcards;

  /// เงินโอน
  late List<TransfersModel>? transfers;

  /// bill header
  late List<LanguageDataModel>? billheader;

  /// bill footer
  late List<LanguageDataModel>? billfooter;

  /// logo pos
  late String? logourl;

  /// จดทะเบียนภาษีมูลค่าเพิ่ม
  late bool? isvatregister;

  /// รหัสสื่อโฆษณา
  late String? mediaguid;

  /// value timezone
  late String? timezoneoffset;

  /// name timezone
  late String? timezonelabel;

  /// time for sale alcohol
  late List<TimeForsaleModel> timeforsales;

  /// ขายเงินเชื่อ
  late bool isusecreadit;

  /// ประเภทธุรกิจ
  late int? businesstype;

  /// ช่องทางการขาย
  late List<SaleChannelModel>? salechannels;

  /// ค่าบริการ
  late double? servicecharge;

  int? categorygroupnumber;
  int? kitchengroupnumber;
  int? tablegroupnumber;
  int? zonegroupnumber;

  PosSettingModel({
    required this.guidfixed,
    required this.branch,
    required this.code,
    required this.devicenumber,
    required this.docformatinv,
    required this.docformattaxinv,
    required this.location,
    required this.receiptform,
    required this.warehouse,
    String? activepin,
    List<EmployeePosModel>? employees,
    String? doccode,
    String? docformatesalereturn,
    int? vattype,
    double? vatrate,
    List<SlipModel>? slips,
    bool? isejournal,
    List<QrModel>? qrcodes,
    List<CreditcardsModel>? creditcards,
    List<TransfersModel>? transfers,
    bool? isvatregister,
    String? mediaguid,
    String? timezoneoffset,
    String? timezonelabel,
    List<TimeForsaleModel>? timeforsales,
    List<LanguageDataModel>? billheader,
    List<LanguageDataModel>? billfooter,
    String? logourl,
    bool? isusecreadit,
    int? businesstype,
    List<SaleChannelModel>? salechannels,
    double? servicecharge,
    int? categorygroupnumber,
    int? kitchengroupnumber,
    int? tablegroupnumber,
    int? zonegroupnumber,
  })  : activepin = activepin ?? '',
        employees = employees ?? <EmployeePosModel>[],
        doccode = doccode ?? '',
        docformatesalereturn = docformatesalereturn ?? '',
        vattype = vattype ?? 0,
        vatrate = vatrate ?? 0.0,
        slips = slips ?? <SlipModel>[],
        isejournal = isejournal ?? false,
        qrcodes = qrcodes ?? <QrModel>[],
        creditcards = creditcards ?? <CreditcardsModel>[],
        transfers = transfers ?? <TransfersModel>[],
        isvatregister = isvatregister ?? true,
        mediaguid = mediaguid ?? '',
        timezoneoffset = timezoneoffset ?? '',
        timezonelabel = timezonelabel ?? '',
        timeforsales = timeforsales ?? <TimeForsaleModel>[],
        billheader = billheader ?? <LanguageDataModel>[],
        billfooter = billfooter ?? <LanguageDataModel>[],
        logourl = logourl ?? '',
        isusecreadit = isusecreadit ?? false,
        businesstype = businesstype ?? 0,
        salechannels = salechannels ?? <SaleChannelModel>[],
        servicecharge = servicecharge ?? 0.0,
        categorygroupnumber = categorygroupnumber ?? 0,
        kitchengroupnumber = kitchengroupnumber ?? 0,
        tablegroupnumber = tablegroupnumber ?? 0,
        zonegroupnumber = zonegroupnumber ?? 0;

  factory PosSettingModel.fromJson(Map<String, dynamic> json) =>
      _$PosSettingModelFromJson(json);
  Map<String, dynamic> toJson() => _$PosSettingModelToJson(this);
}

@JsonSerializable()
class BranchModel {
  late String code;
  late String guidfixed;
  late List<LanguageDataModel> names;

  BranchModel({
    required this.code,
    required this.guidfixed,
    required this.names,
  });

  factory BranchModel.fromJson(Map<String, dynamic> json) =>
      _$BranchModelFromJson(json);
  Map<String, dynamic> toJson() => _$BranchModelToJson(this);
}

@JsonSerializable()
class LocationModel {
  late String code;
  late List<LanguageDataModel> names;

  LocationModel({
    required this.code,
    required this.names,
  });

  factory LocationModel.fromJson(Map<String, dynamic> json) =>
      _$LocationModelFromJson(json);
  Map<String, dynamic> toJson() => _$LocationModelToJson(this);
}

@JsonSerializable()
class WarehouseModel {
  late String code;
  late String guidfixed;
  late List<LanguageDataModel> names;

  WarehouseModel({
    required this.code,
    required this.guidfixed,
    required this.names,
  });

  factory WarehouseModel.fromJson(Map<String, dynamic> json) =>
      _$WarehouseModelFromJson(json);
  Map<String, dynamic> toJson() => _$WarehouseModelToJson(this);
}

@JsonSerializable()
class EmployeePosModel {
  String? code;
  String? guidfixed;
  String? name;
  List<String>? permissions;

  EmployeePosModel({
    String? code,
    String? guidfixed,
    String? name,
    List<String>? permissions,
  })  : code = code ?? '',
        guidfixed = guidfixed ?? '',
        name = name ?? '',
        permissions = permissions ?? <String>[];

  factory EmployeePosModel.fromJson(Map<String, dynamic> json) =>
      _$EmployeePosModelFromJson(json);
  Map<String, dynamic> toJson() => _$EmployeePosModelToJson(this);
}

/// slips
@JsonSerializable()
class SlipModel {
  late String code;
  late bool isrequire;
  late String name;
  late String formcode;
  late List<LanguageDataModel> formnames = <LanguageDataModel>[];
  late List<LanguageDataModel> headernames = <LanguageDataModel>[];

  SlipModel({
    required this.code,
    required this.isrequire,
    required this.name,
    String? formcode,
    List<LanguageDataModel>? formnames,
    List<LanguageDataModel>? headernames,
  })  : formnames = formnames ?? <LanguageDataModel>[],
        formcode = formcode ?? '',
        headernames = headernames ?? <LanguageDataModel>[];

  factory SlipModel.fromJson(Map<String, dynamic> json) =>
      _$SlipModelFromJson(json);
  Map<String, dynamic> toJson() => _$SlipModelToJson(this);
}

/// เวลาขายสุรา
@JsonSerializable()
class TimeForsaleModel {
  late String from;
  late String to;
  late List<LanguageDataModel> names;

  TimeForsaleModel({
    required this.from,
    required this.to,
    required this.names,
  });

  factory TimeForsaleModel.fromJson(Map<String, dynamic> json) =>
      _$TimeForsaleModelFromJson(json);
  Map<String, dynamic> toJson() => _$TimeForsaleModelToJson(this);
}

/// เครดิตการ์ด
@JsonSerializable()
class CreditcardsModel {
  List<LanguageDataModel>? names;
  BookBankModel? bookbank;

  CreditcardsModel({
    List<LanguageDataModel>? names,
    BookBankModel? bookbank,
  })  : names = names ?? <LanguageDataModel>[],
        bookbank = bookbank ?? BookBankModel();

  factory CreditcardsModel.fromJson(Map<String, dynamic> json) =>
      _$CreditcardsModelFromJson(json);
  Map<String, dynamic> toJson() => _$CreditcardsModelToJson(this);
}

/// เงินโอน
@JsonSerializable()
class TransfersModel {
  List<LanguageDataModel>? names;
  BookBankModel? bookbank;

  TransfersModel({
    List<LanguageDataModel>? names,
    BookBankModel? bookbank,
  })  : names = names ?? <LanguageDataModel>[],
        bookbank = bookbank ?? BookBankModel();

  factory TransfersModel.fromJson(Map<String, dynamic> json) =>
      _$TransfersModelFromJson(json);
  Map<String, dynamic> toJson() => _$TransfersModelToJson(this);
}
