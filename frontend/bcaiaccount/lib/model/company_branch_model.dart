import 'package:smlaicloud/model/contact_model.dart';
import 'package:smlaicloud/model/business_type_model.dart';
import 'package:smlaicloud/model/department_model.dart';
import 'package:json_annotation/json_annotation.dart';
import 'package:smlaicloud/model/global_model.dart';

part 'company_branch_model.g.dart';

@JsonSerializable(explicitToJson: true)
class CompanyBranchModel {
  String guidfixed;
  List<LanguageDataModel> companynames = <LanguageDataModel>[];
  String code;
  List<LanguageDataModel> names = <LanguageDataModel>[];
  List<DepartmentModel> departments = <DepartmentModel>[];
  List<String> languages;
  ContactModel? contact;
  String? imageuri;
  String? logouri;
  PosModel? pos;
  BusinessTypeModel? businesstype;
  PaymentRoundingModel? paymentrounding;
  PointConfigModel? pointconfig;
  int? machinetype;

  /// 0 = ใช้ได้หลายใบ , 1 = ใช้ได้ใบเดียว
  int? couponusetype;

  /// สกุลเงินหลักของสาขา (เช่น THB, USD, VND)
  @JsonKey(name: 'base_currency')
  String? baseCurrency;

  /// ภาษาที่ใช้ในสาขา (th, en, vi, lo, km, my, cn, ja, ko)
  @JsonKey(name: 'language')
  String? language;

  /// Timezone ของสาขา (เช่น "Asia/Bangkok", "Asia/Ho_Chi_Minh")
  @JsonKey(name: 'timezone')
  String? timezone;

  /// รูปแบบวันที่ของสาขา (เช่น "dd/MM/yyyy", "yyyy-MM-dd", "MM/dd/yyyy")
  @JsonKey(name: 'date_format')
  String? dateFormat;

  /// ประเภทปี (christian = ค.ศ., buddhist = พ.ศ.)
  @JsonKey(name: 'yeartype')
  String? yeartype;

  /// Timezone label สำหรับแสดงผล (เช่น "(UTC+07:00) Bangkok, Hanoi, Jakarta")
  @JsonKey(name: 'timezonelabel')
  String? timezonelabel;

  /// Timezone offset (เช่น "+07:00")
  @JsonKey(name: 'timezoneoffset')
  String? timezoneoffset;

  /// ทศนิยมจำนวนสินค้า (2-8, default=2)
  @JsonKey(name: 'decimal_quantity')
  int? decimalQuantity;

  /// ทศนิยมมูลค่าสินค้า (2-8, default=2)
  @JsonKey(name: 'decimal_price')
  int? decimalPrice;

  /// ทศนิยมมูลค่าเอกสาร (2-8, default=2)
  @JsonKey(name: 'decimal_document')
  int? decimalDocument;

  /// คุณสมบัติธุรกิจ
  @JsonKey(name: 'is_restaurant')
  bool? isRestaurant;

  @JsonKey(name: 'is_tire')
  bool? isTire;

  @JsonKey(name: 'is_agriculture')
  bool? isAgriculture;

  @JsonKey(name: 'is_pharmacy')
  bool? isPharmacy;

  CompanyBranchModel({
    required this.guidfixed,
    required this.code,
    List<LanguageDataModel>? companynames,
    List<LanguageDataModel>? names,
    List<DepartmentModel>? departments,
    List<String>? languages,
    ContactModel? contact,
    String? imageuri,
    String? logouri,
    PosModel? pos,
    BusinessTypeModel? businesstype,
    PaymentRoundingModel? paymentrounding,
    PointConfigModel? pointconfig,
    int? machinetype,
    int? couponusetype,
    this.baseCurrency,
    this.language,
    this.timezone,
    this.dateFormat,
    this.yeartype,
    this.timezonelabel,
    this.timezoneoffset,
    this.decimalQuantity,
    this.decimalPrice,
    this.decimalDocument,
    this.isRestaurant,
    this.isTire,
    this.isAgriculture,
    this.isPharmacy,
  }) : names = names ?? <LanguageDataModel>[],
       companynames = companynames ?? <LanguageDataModel>[],
       departments = departments ?? <DepartmentModel>[],
       languages = languages ?? <String>[],
       contact = contact ?? ContactModel(),
       logouri = logouri ?? "",
       imageuri = imageuri ?? "",
       pos = pos ?? PosModel(),
       businesstype = businesstype ?? BusinessTypeModel(),
       paymentrounding = paymentrounding ?? PaymentRoundingModel(),
       pointconfig = pointconfig ?? PointConfigModel(),
       machinetype = machinetype ?? 0,
       couponusetype = couponusetype ?? 0;

  factory CompanyBranchModel.fromJson(Map<String, dynamic> json) =>
      _$CompanyBranchModelFromJson(json);

  Map<String, dynamic> toJson() => _$CompanyBranchModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class PosModel {
  String? taxid;
  double? vatrate;
  int? vattypesale;
  int? vattypepurchase;
  int? inquirytypesale;
  int? inquirytypepurchase;
  String? headerreceiptpos;
  String? footerreceiptpos;
  bool? isbom;

  PosModel({
    String? taxid,
    double? vatrate,
    int? vattypesale,
    int? vattypepurchase,
    int? inquirytypesale,
    int? inquirytypepurchase,
    String? headerreceiptpos,
    String? footerreceiptpos,
    bool? isbom,
  }) : taxid = taxid ?? "",
       vatrate = vatrate ?? 0.0,
       vattypesale = vattypesale ?? 0,
       vattypepurchase = vattypepurchase ?? 0,
       inquirytypesale = inquirytypesale ?? 0,
       inquirytypepurchase = inquirytypepurchase ?? 0,
       headerreceiptpos = headerreceiptpos ?? "",
       footerreceiptpos = footerreceiptpos ?? "",
       isbom = isbom ?? false;

  factory PosModel.fromJson(Map<String, dynamic> json) =>
      _$PosModelFromJson(json);

  Map<String, dynamic> toJson() => _$PosModelToJson(this);
}

@JsonSerializable()
class BranchModel {
  String? code;
  String? guidfixed;
  List<LanguageDataModel>? names;
  bool? isignore;

  BranchModel({
    String? code,
    String? guidfixed,
    List<LanguageDataModel>? names,
    bool? isignore,
  }) : code = code ?? '',
       guidfixed = guidfixed ?? '',
       names = names ?? [],
       isignore = isignore ?? false;

  factory BranchModel.fromJson(Map<String, dynamic> json) =>
      _$BranchModelFromJson(json);
  Map<String, dynamic> toJson() => _$BranchModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class PaymentRoundingModel {
  PaymentMethodRoundingModel banktransfer;
  PaymentMethodRoundingModel cash;
  PaymentMethodRoundingModel cheque;
  PaymentMethodRoundingModel coupon;
  PaymentMethodRoundingModel creditcard;
  PaymentMethodRoundingModel delivery;
  PaymentMethodRoundingModel qrcode;

  PaymentRoundingModel({
    PaymentMethodRoundingModel? banktransfer,
    PaymentMethodRoundingModel? cash,
    PaymentMethodRoundingModel? cheque,
    PaymentMethodRoundingModel? coupon,
    PaymentMethodRoundingModel? creditcard,
    PaymentMethodRoundingModel? delivery,
    PaymentMethodRoundingModel? qrcode,
  }) : banktransfer = banktransfer ?? PaymentMethodRoundingModel(),
       cash = cash ?? PaymentMethodRoundingModel(),
       cheque = cheque ?? PaymentMethodRoundingModel(),
       coupon = coupon ?? PaymentMethodRoundingModel(),
       creditcard = creditcard ?? PaymentMethodRoundingModel(),
       delivery = delivery ?? PaymentMethodRoundingModel(),
       qrcode = qrcode ?? PaymentMethodRoundingModel();

  factory PaymentRoundingModel.fromJson(Map<String, dynamic> json) =>
      _$PaymentRoundingModelFromJson(json);

  Map<String, dynamic> toJson() => _$PaymentRoundingModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class PaymentMethodRoundingModel {
  bool enabled;
  List<RoundingRuleModel> rules;

  PaymentMethodRoundingModel({bool? enabled, List<RoundingRuleModel>? rules})
    : enabled = enabled ?? true,
      rules = rules ?? [RoundingRuleModel()];

  factory PaymentMethodRoundingModel.fromJson(Map<String, dynamic> json) =>
      _$PaymentMethodRoundingModelFromJson(json);

  Map<String, dynamic> toJson() => _$PaymentMethodRoundingModelToJson(this);
}

@JsonSerializable()
class RoundingRuleModel {
  double lowerbound;
  double roundto;
  double upperbound;

  RoundingRuleModel({double? lowerbound, double? roundto, double? upperbound})
    : lowerbound = lowerbound ?? 0,
      roundto = roundto ?? 0,
      upperbound = upperbound ?? 0;

  factory RoundingRuleModel.fromJson(Map<String, dynamic> json) =>
      _$RoundingRuleModelFromJson(json);

  Map<String, dynamic> toJson() => _$RoundingRuleModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class PointConfigModel {
  List<GeneralRuleModel> generalrules;
  List<SpecialRuleModel> specialrules;
  int pointusagetype;
  PointConfigModel({
    List<GeneralRuleModel>? generalrules,
    List<SpecialRuleModel>? specialrules,
    int? pointusagetype,
  }) : generalrules = generalrules ?? <GeneralRuleModel>[],
       specialrules = specialrules ?? <SpecialRuleModel>[],
       pointusagetype = pointusagetype ?? 1;

  factory PointConfigModel.fromJson(Map<String, dynamic> json) =>
      _$PointConfigModelFromJson(json);

  Map<String, dynamic> toJson() => _$PointConfigModelToJson(this);
}

@JsonSerializable()
class GeneralRuleModel {
  String startdate;
  String enddate;
  double payperpoint;
  double pointvalue;
  GeneralRuleModel({
    String? startdate,
    String? enddate,
    double? payperpoint,
    double? pointvalue,
  }) : startdate = startdate ?? DateTime.now().toUtc().toIso8601String(),
       enddate =
           enddate ??
           DateTime.now()
               .toUtc()
               .add(const Duration(days: 365))
               .toIso8601String(),
       payperpoint = payperpoint ?? 20,
       pointvalue = pointvalue ?? 1;

  factory GeneralRuleModel.fromJson(Map<String, dynamic> json) =>
      _$GeneralRuleModelFromJson(json);

  Map<String, dynamic> toJson() => _$GeneralRuleModelToJson(this);
}

@JsonSerializable()
class SpecialRuleModel {
  String startdate;
  String enddate;
  double multiplier;
  bool sunday;
  bool monday;
  bool tuesday;
  bool wednesday;
  bool thursday;
  bool friday;
  bool saturday;
  double maxpointperbill;

  SpecialRuleModel({
    String? startdate,
    String? enddate,
    double? multiplier,
    bool? sunday,
    bool? monday,
    bool? tuesday,
    bool? wednesday,
    bool? thursday,
    bool? friday,
    bool? saturday,
    double? maxpointperbill,
  }) : startdate = startdate ?? DateTime.now().toUtc().toIso8601String(),
       enddate =
           enddate ??
           DateTime.now()
               .toUtc()
               .add(const Duration(days: 30))
               .toIso8601String(),
       multiplier = multiplier ?? 2,
       sunday = sunday ?? false,
       monday = monday ?? true,
       tuesday = tuesday ?? true,
       wednesday = wednesday ?? true,
       thursday = thursday ?? true,
       friday = friday ?? true,
       saturday = saturday ?? false,
       maxpointperbill = maxpointperbill ?? 100;

  factory SpecialRuleModel.fromJson(Map<String, dynamic> json) =>
      _$SpecialRuleModelFromJson(json);

  Map<String, dynamic> toJson() => _$SpecialRuleModelToJson(this);
}
