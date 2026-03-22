// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'company_branch_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

CompanyBranchModel _$CompanyBranchModelFromJson(Map<String, dynamic> json) =>
    CompanyBranchModel(
      guidfixed: json['guidfixed'] as String,
      code: json['code'] as String,
      companynames: (json['companynames'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      names: (json['names'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      departments: (json['departments'] as List<dynamic>?)
          ?.map((e) => DepartmentModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      languages: (json['languages'] as List<dynamic>?)
          ?.map((e) => e as String)
          .toList(),
      contact: json['contact'] == null
          ? null
          : ContactModel.fromJson(json['contact'] as Map<String, dynamic>),
      imageuri: json['imageuri'] as String?,
      logouri: json['logouri'] as String?,
      pos: json['pos'] == null
          ? null
          : PosModel.fromJson(json['pos'] as Map<String, dynamic>),
      businesstype: json['businesstype'] == null
          ? null
          : BusinessTypeModel.fromJson(
              json['businesstype'] as Map<String, dynamic>,
            ),
      paymentrounding: json['paymentrounding'] == null
          ? null
          : PaymentRoundingModel.fromJson(
              json['paymentrounding'] as Map<String, dynamic>,
            ),
      pointconfig: json['pointconfig'] == null
          ? null
          : PointConfigModel.fromJson(
              json['pointconfig'] as Map<String, dynamic>,
            ),
      machinetype: (json['machinetype'] as num?)?.toInt(),
      couponusetype: (json['couponusetype'] as num?)?.toInt(),
      baseCurrency: json['base_currency'] as String?,
      language: json['language'] as String?,
      timezone: json['timezone'] as String?,
      dateFormat: json['date_format'] as String?,
      yeartype: json['yeartype'] as String?,
      timezonelabel: json['timezonelabel'] as String?,
      timezoneoffset: json['timezoneoffset'] as String?,
      decimalQuantity: (json['decimal_quantity'] as num?)?.toInt(),
      decimalPrice: (json['decimal_price'] as num?)?.toInt(),
      decimalDocument: (json['decimal_document'] as num?)?.toInt(),
      isRestaurant: json['is_restaurant'] as bool?,
      isTire: json['is_tire'] as bool?,
      isAgriculture: json['is_agriculture'] as bool?,
      isPharmacy: json['is_pharmacy'] as bool?,
      isService: json['is_service'] as bool?,
      isManufacturing: json['is_manufacturing'] as bool?,
      isImportExport: json['is_import_export'] as bool?,
      isContractor: json['is_contractor'] as bool?,
      isRental: json['is_rental'] as bool?,
      isWholesale: json['is_wholesale'] as bool?,
      isEcommerce: json['is_ecommerce'] as bool?,
      isLogistics: json['is_logistics'] as bool?,
      isEducation: json['is_education'] as bool?,
      isHotel: json['is_hotel'] as bool?,
      isBeauty: json['is_beauty'] as bool?,
      isGoldShop: json['is_gold_shop'] as bool?,
      isAccountingFirm: json['is_accounting_firm'] as bool?,
      isRetail: json['is_retail'] as bool?,
      isConstruction: json['is_construction'] as bool?,
      isElectronics: json['is_electronics'] as bool?,
      isMobileShop: json['is_mobile_shop'] as bool?,
    );

Map<String, dynamic> _$CompanyBranchModelToJson(CompanyBranchModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'companynames': instance.companynames.map((e) => e.toJson()).toList(),
      'code': instance.code,
      'names': instance.names.map((e) => e.toJson()).toList(),
      'departments': instance.departments.map((e) => e.toJson()).toList(),
      'languages': instance.languages,
      'contact': instance.contact?.toJson(),
      'imageuri': instance.imageuri,
      'logouri': instance.logouri,
      'pos': instance.pos?.toJson(),
      'businesstype': instance.businesstype?.toJson(),
      'paymentrounding': instance.paymentrounding?.toJson(),
      'pointconfig': instance.pointconfig?.toJson(),
      'machinetype': instance.machinetype,
      'couponusetype': instance.couponusetype,
      'base_currency': instance.baseCurrency,
      'language': instance.language,
      'timezone': instance.timezone,
      'date_format': instance.dateFormat,
      'yeartype': instance.yeartype,
      'timezonelabel': instance.timezonelabel,
      'timezoneoffset': instance.timezoneoffset,
      'decimal_quantity': instance.decimalQuantity,
      'decimal_price': instance.decimalPrice,
      'decimal_document': instance.decimalDocument,
      'is_restaurant': instance.isRestaurant,
      'is_tire': instance.isTire,
      'is_agriculture': instance.isAgriculture,
      'is_pharmacy': instance.isPharmacy,
      'is_service': instance.isService,
      'is_manufacturing': instance.isManufacturing,
      'is_import_export': instance.isImportExport,
      'is_contractor': instance.isContractor,
      'is_rental': instance.isRental,
      'is_wholesale': instance.isWholesale,
      'is_ecommerce': instance.isEcommerce,
      'is_logistics': instance.isLogistics,
      'is_education': instance.isEducation,
      'is_hotel': instance.isHotel,
      'is_beauty': instance.isBeauty,
      'is_gold_shop': instance.isGoldShop,
      'is_accounting_firm': instance.isAccountingFirm,
      'is_retail': instance.isRetail,
      'is_construction': instance.isConstruction,
      'is_electronics': instance.isElectronics,
      'is_mobile_shop': instance.isMobileShop,
    };

PosModel _$PosModelFromJson(Map<String, dynamic> json) => PosModel(
  taxid: json['taxid'] as String?,
  vatrate: (json['vatrate'] as num?)?.toDouble(),
  vattypesale: (json['vattypesale'] as num?)?.toInt(),
  vattypepurchase: (json['vattypepurchase'] as num?)?.toInt(),
  inquirytypesale: (json['inquirytypesale'] as num?)?.toInt(),
  inquirytypepurchase: (json['inquirytypepurchase'] as num?)?.toInt(),
  headerreceiptpos: json['headerreceiptpos'] as String?,
  footerreceiptpos: json['footerreceiptpos'] as String?,
  isbom: json['isbom'] as bool?,
);

Map<String, dynamic> _$PosModelToJson(PosModel instance) => <String, dynamic>{
  'taxid': instance.taxid,
  'vatrate': instance.vatrate,
  'vattypesale': instance.vattypesale,
  'vattypepurchase': instance.vattypepurchase,
  'inquirytypesale': instance.inquirytypesale,
  'inquirytypepurchase': instance.inquirytypepurchase,
  'headerreceiptpos': instance.headerreceiptpos,
  'footerreceiptpos': instance.footerreceiptpos,
  'isbom': instance.isbom,
};

BranchModel _$BranchModelFromJson(Map<String, dynamic> json) => BranchModel(
  code: json['code'] as String?,
  guidfixed: json['guidfixed'] as String?,
  names: (json['names'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  isignore: json['isignore'] as bool?,
);

Map<String, dynamic> _$BranchModelToJson(BranchModel instance) =>
    <String, dynamic>{
      'code': instance.code,
      'guidfixed': instance.guidfixed,
      'names': instance.names,
      'isignore': instance.isignore,
    };

PaymentRoundingModel _$PaymentRoundingModelFromJson(
  Map<String, dynamic> json,
) => PaymentRoundingModel(
  banktransfer: json['banktransfer'] == null
      ? null
      : PaymentMethodRoundingModel.fromJson(
          json['banktransfer'] as Map<String, dynamic>,
        ),
  cash: json['cash'] == null
      ? null
      : PaymentMethodRoundingModel.fromJson(
          json['cash'] as Map<String, dynamic>,
        ),
  cheque: json['cheque'] == null
      ? null
      : PaymentMethodRoundingModel.fromJson(
          json['cheque'] as Map<String, dynamic>,
        ),
  coupon: json['coupon'] == null
      ? null
      : PaymentMethodRoundingModel.fromJson(
          json['coupon'] as Map<String, dynamic>,
        ),
  creditcard: json['creditcard'] == null
      ? null
      : PaymentMethodRoundingModel.fromJson(
          json['creditcard'] as Map<String, dynamic>,
        ),
  delivery: json['delivery'] == null
      ? null
      : PaymentMethodRoundingModel.fromJson(
          json['delivery'] as Map<String, dynamic>,
        ),
  qrcode: json['qrcode'] == null
      ? null
      : PaymentMethodRoundingModel.fromJson(
          json['qrcode'] as Map<String, dynamic>,
        ),
);

Map<String, dynamic> _$PaymentRoundingModelToJson(
  PaymentRoundingModel instance,
) => <String, dynamic>{
  'banktransfer': instance.banktransfer.toJson(),
  'cash': instance.cash.toJson(),
  'cheque': instance.cheque.toJson(),
  'coupon': instance.coupon.toJson(),
  'creditcard': instance.creditcard.toJson(),
  'delivery': instance.delivery.toJson(),
  'qrcode': instance.qrcode.toJson(),
};

PaymentMethodRoundingModel _$PaymentMethodRoundingModelFromJson(
  Map<String, dynamic> json,
) => PaymentMethodRoundingModel(
  enabled: json['enabled'] as bool?,
  rules: (json['rules'] as List<dynamic>?)
      ?.map((e) => RoundingRuleModel.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$PaymentMethodRoundingModelToJson(
  PaymentMethodRoundingModel instance,
) => <String, dynamic>{
  'enabled': instance.enabled,
  'rules': instance.rules.map((e) => e.toJson()).toList(),
};

RoundingRuleModel _$RoundingRuleModelFromJson(Map<String, dynamic> json) =>
    RoundingRuleModel(
      lowerbound: (json['lowerbound'] as num?)?.toDouble(),
      roundto: (json['roundto'] as num?)?.toDouble(),
      upperbound: (json['upperbound'] as num?)?.toDouble(),
    );

Map<String, dynamic> _$RoundingRuleModelToJson(RoundingRuleModel instance) =>
    <String, dynamic>{
      'lowerbound': instance.lowerbound,
      'roundto': instance.roundto,
      'upperbound': instance.upperbound,
    };

PointConfigModel _$PointConfigModelFromJson(Map<String, dynamic> json) =>
    PointConfigModel(
      generalrules: (json['generalrules'] as List<dynamic>?)
          ?.map((e) => GeneralRuleModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      specialrules: (json['specialrules'] as List<dynamic>?)
          ?.map((e) => SpecialRuleModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      pointusagetype: (json['pointusagetype'] as num?)?.toInt(),
    );

Map<String, dynamic> _$PointConfigModelToJson(PointConfigModel instance) =>
    <String, dynamic>{
      'generalrules': instance.generalrules.map((e) => e.toJson()).toList(),
      'specialrules': instance.specialrules.map((e) => e.toJson()).toList(),
      'pointusagetype': instance.pointusagetype,
    };

GeneralRuleModel _$GeneralRuleModelFromJson(Map<String, dynamic> json) =>
    GeneralRuleModel(
      startdate: json['startdate'] as String?,
      enddate: json['enddate'] as String?,
      payperpoint: (json['payperpoint'] as num?)?.toDouble(),
      pointvalue: (json['pointvalue'] as num?)?.toDouble(),
    );

Map<String, dynamic> _$GeneralRuleModelToJson(GeneralRuleModel instance) =>
    <String, dynamic>{
      'startdate': instance.startdate,
      'enddate': instance.enddate,
      'payperpoint': instance.payperpoint,
      'pointvalue': instance.pointvalue,
    };

SpecialRuleModel _$SpecialRuleModelFromJson(Map<String, dynamic> json) =>
    SpecialRuleModel(
      startdate: json['startdate'] as String?,
      enddate: json['enddate'] as String?,
      multiplier: (json['multiplier'] as num?)?.toDouble(),
      sunday: json['sunday'] as bool?,
      monday: json['monday'] as bool?,
      tuesday: json['tuesday'] as bool?,
      wednesday: json['wednesday'] as bool?,
      thursday: json['thursday'] as bool?,
      friday: json['friday'] as bool?,
      saturday: json['saturday'] as bool?,
      maxpointperbill: (json['maxpointperbill'] as num?)?.toDouble(),
    );

Map<String, dynamic> _$SpecialRuleModelToJson(SpecialRuleModel instance) =>
    <String, dynamic>{
      'startdate': instance.startdate,
      'enddate': instance.enddate,
      'multiplier': instance.multiplier,
      'sunday': instance.sunday,
      'monday': instance.monday,
      'tuesday': instance.tuesday,
      'wednesday': instance.wednesday,
      'thursday': instance.thursday,
      'friday': instance.friday,
      'saturday': instance.saturday,
      'maxpointperbill': instance.maxpointperbill,
    };
