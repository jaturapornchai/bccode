// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'doc_format_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

DocFormatModel _$DocFormatModelFromJson(Map<String, dynamic> json) =>
    DocFormatModel(
      guidfixed: json['guidfixed'] as String?,
      dateformate: json['dateformate'] as String,
      details: (json['details'] as List<dynamic>)
          .map((e) => DetailModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      doccode: json['doccode'] as String,
      docnumber: (json['docnumber'] as num).toInt(),
      module: json['module'] as String,
      description: json['description'] as String?,
      docformat: json['docformat'] as String,
      isautoformat: json['isautoformat'] as bool,
      yeartype: (json['yeartype'] as num).toInt(),
      accountgroup: json['accountgroup'] as String?,
      bookcode: json['bookcode'] as String?,
      ischeck: json['ischeck'] as bool?,
      totaldocument: (json['totaldocument'] as num?)?.toInt(),
      successdocument: (json['successdocument'] as num?)?.toDouble(),
    );

Map<String, dynamic> _$DocFormatModelToJson(DocFormatModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'dateformate': instance.dateformate,
      'details': instance.details,
      'doccode': instance.doccode,
      'docnumber': instance.docnumber,
      'module': instance.module,
      'description': instance.description,
      'docformat': instance.docformat,
      'isautoformat': instance.isautoformat,
      'yeartype': instance.yeartype,
      'accountgroup': instance.accountgroup,
      'bookcode': instance.bookcode,
      'ischeck': instance.ischeck,
      'totaldocument': instance.totaldocument,
      'successdocument': instance.successdocument,
    };

DetailModel _$DetailModelFromJson(Map<String, dynamic> json) => DetailModel(
  actioncode: json['actioncode'] as String,
  credit: json['credit'] as String,
  debit: json['debit'] as String,
  detail: json['detail'] as String,
  isentryselfaccount: json['isentryselfaccount'] as bool,
  accountcredit: json['accountcredit'] == null
      ? null
      : AccountChartModel.fromJson(
          json['accountcredit'] as Map<String, dynamic>,
        ),
  accountdebit: json['accountdebit'] == null
      ? null
      : AccountChartModel.fromJson(
          json['accountdebit'] as Map<String, dynamic>,
        ),
);

Map<String, dynamic> _$DetailModelToJson(DetailModel instance) =>
    <String, dynamic>{
      'actioncode': instance.actioncode,
      'credit': instance.credit,
      'debit': instance.debit,
      'detail': instance.detail,
      'isentryselfaccount': instance.isentryselfaccount,
      'accountcredit': instance.accountcredit,
      'accountdebit': instance.accountdebit,
    };

DefaultDocFormatModel _$DefaultDocFormatModelFromJson(
  Map<String, dynamic> json,
) => DefaultDocFormatModel(
  dateformate: json['dateformate'] as String,
  doccode: json['doccode'] as String,
  docnumber: (json['docnumber'] as num).toInt(),
  name: json['name'] as String,
);

Map<String, dynamic> _$DefaultDocFormatModelToJson(
  DefaultDocFormatModel instance,
) => <String, dynamic>{
  'dateformate': instance.dateformate,
  'doccode': instance.doccode,
  'docnumber': instance.docnumber,
  'name': instance.name,
};
