// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'journal_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

JournalModel _$JournalModelFromJson(Map<String, dynamic> json) => JournalModel(
  accountdescription: json['accountdescription'] as String?,
  accountgroup: json['accountgroup'] as String?,
  accountperiod: (json['accountperiod'] as num?)?.toInt(),
  accountyear: (json['accountyear'] as num?)?.toInt(),
  amount: (json['amount'] as num?)?.toDouble(),
  batchid: json['batchid'] as String?,
  bookcode: json['bookcode'] as String?,
  docdate: json['docdate'] as String?,
  docformat: json['docformat'] as String?,
  docno: json['docno'] as String?,
  documentref: json['documentref'] as String?,
  exdocrefdate: json['exdocrefdate'] as String?,
  exdocrefno: json['exdocrefno'] as String?,
  journaltype: (json['journaltype'] as num?)?.toInt(),
  journaldetail: (json['journaldetail'] as List<dynamic>?)
      ?.map((e) => JournalDetailModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  taxes: (json['taxes'] as List<dynamic>?)
      ?.map((e) => TaxesModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  vats: (json['vats'] as List<dynamic>?)
      ?.map((e) => VatsModel.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$JournalModelToJson(JournalModel instance) =>
    <String, dynamic>{
      'accountdescription': instance.accountdescription,
      'accountgroup': instance.accountgroup,
      'accountperiod': instance.accountperiod,
      'accountyear': instance.accountyear,
      'amount': instance.amount,
      'batchid': instance.batchid,
      'bookcode': instance.bookcode,
      'docdate': instance.docdate,
      'docformat': instance.docformat,
      'docno': instance.docno,
      'documentref': instance.documentref,
      'exdocrefdate': instance.exdocrefdate,
      'exdocrefno': instance.exdocrefno,
      'journaltype': instance.journaltype,
      'journaldetail': instance.journaldetail,
      'taxes': instance.taxes,
      'vats': instance.vats,
    };

JournalDetailModel _$JournalDetailModelFromJson(Map<String, dynamic> json) =>
    JournalDetailModel(
      accountcode: json['accountcode'] as String?,
      accountname: json['accountname'] as String?,
      creditamount: (json['creditamount'] as num?)?.toDouble(),
      debitamount: (json['debitamount'] as num?)?.toDouble(),
    );

Map<String, dynamic> _$JournalDetailModelToJson(JournalDetailModel instance) =>
    <String, dynamic>{
      'accountcode': instance.accountcode,
      'accountname': instance.accountname,
      'creditamount': instance.creditamount,
      'debitamount': instance.debitamount,
    };

TaxesModel _$TaxesModelFromJson(Map<String, dynamic> json) => TaxesModel(
  address: json['address'] as String?,
  branchcode: json['branchcode'] as String?,
  custname: json['custname'] as String?,
  custtaxid: json['custtaxid'] as String?,
  custtype: (json['custtype'] as num?)?.toInt(),
  details: (json['details'] as List<dynamic>?)
      ?.map((e) => DetailsModelModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  organization: (json['organization'] as num?)?.toInt(),
  taxamount: (json['taxamount'] as num?)?.toDouble(),
  taxdate: json['taxdate'] as String?,
  taxdocno: json['taxdocno'] as String?,
  taxtype: (json['taxtype'] as num?)?.toInt(),
);

Map<String, dynamic> _$TaxesModelToJson(TaxesModel instance) =>
    <String, dynamic>{
      'address': instance.address,
      'branchcode': instance.branchcode,
      'custname': instance.custname,
      'custtaxid': instance.custtaxid,
      'custtype': instance.custtype,
      'details': instance.details,
      'organization': instance.organization,
      'taxamount': instance.taxamount,
      'taxdate': instance.taxdate,
      'taxdocno': instance.taxdocno,
      'taxtype': instance.taxtype,
    };

DetailsModelModel _$DetailsModelModelFromJson(Map<String, dynamic> json) =>
    DetailsModelModel(
      description: json['description'] as String?,
      taxamount: (json['taxamount'] as num?)?.toDouble(),
      taxbase: (json['taxbase'] as num?)?.toDouble(),
      taxrate: (json['taxrate'] as num?)?.toDouble(),
    );

Map<String, dynamic> _$DetailsModelModelToJson(DetailsModelModel instance) =>
    <String, dynamic>{
      'description': instance.description,
      'taxamount': instance.taxamount,
      'taxbase': instance.taxbase,
      'taxrate': instance.taxrate,
    };

VatsModel _$VatsModelFromJson(Map<String, dynamic> json) => VatsModel(
  address: json['address'] as String?,
  branchcode: json['branchcode'] as String?,
  custname: json['custname'] as String?,
  custtaxid: json['custtaxid'] as String?,
  custtype: (json['custtype'] as num?)?.toInt(),
  exceptvat: (json['exceptvat'] as num?)?.toInt(),
  organization: (json['organization'] as num?)?.toInt(),
  remark: json['remark'] as String?,
  vatamount: (json['vatamount'] as num?)?.toDouble(),
  vatbase: (json['vatbase'] as num?)?.toDouble(),
  vatdate: json['vatdate'] as String?,
  vatdocno: json['vatdocno'] as String?,
  vatmode: (json['vatmode'] as num?)?.toInt(),
  vatperiod: (json['vatperiod'] as num?)?.toInt(),
  vatrate: (json['vatrate'] as num?)?.toInt(),
  vatsubmit: json['vatsubmit'] as bool?,
  vattype: (json['vattype'] as num?)?.toInt(),
  vatyear: (json['vatyear'] as num?)?.toInt(),
);

Map<String, dynamic> _$VatsModelToJson(VatsModel instance) => <String, dynamic>{
  'address': instance.address,
  'branchcode': instance.branchcode,
  'custname': instance.custname,
  'custtaxid': instance.custtaxid,
  'custtype': instance.custtype,
  'exceptvat': instance.exceptvat,
  'organization': instance.organization,
  'remark': instance.remark,
  'vatamount': instance.vatamount,
  'vatbase': instance.vatbase,
  'vatdate': instance.vatdate,
  'vatdocno': instance.vatdocno,
  'vatmode': instance.vatmode,
  'vatperiod': instance.vatperiod,
  'vatrate': instance.vatrate,
  'vatsubmit': instance.vatsubmit,
  'vattype': instance.vattype,
  'vatyear': instance.vatyear,
};
