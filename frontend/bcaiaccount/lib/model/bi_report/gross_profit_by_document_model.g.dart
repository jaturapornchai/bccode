// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'gross_profit_by_document_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

GrossProfitByDocumentModel _$GrossProfitByDocumentModelFromJson(
  Map<String, dynamic> json,
) => GrossProfitByDocumentModel(
  docdate: json['docdate'] as String?,
  docno: json['docno'] as String?,
  creditorcode: json['creditorcode'] as String?,
  creditornames: (json['creditornames'] as List<dynamic>?)
      ?.map((e) => CreditorName.fromJson(e as Map<String, dynamic>))
      .toList(),
  totalbeforevat: (json['totalbeforevat'] as num?)?.toDouble(),
  detailtotaldiscount: (json['detailtotaldiscount'] as num?)?.toDouble(),
  totalafterdiscount: (json['totalafterdiscount'] as num?)?.toDouble(),
  totalcost: (json['totalcost'] as num?)?.toDouble(),
  pal: (json['pal'] as num?)?.toDouble(),
  perPal: (json['per_pal'] as num?)?.toDouble(),
);

Map<String, dynamic> _$GrossProfitByDocumentModelToJson(
  GrossProfitByDocumentModel instance,
) => <String, dynamic>{
  'docdate': instance.docdate,
  'docno': instance.docno,
  'creditorcode': instance.creditorcode,
  'creditornames': instance.creditornames,
  'totalbeforevat': instance.totalbeforevat,
  'detailtotaldiscount': instance.detailtotaldiscount,
  'totalafterdiscount': instance.totalafterdiscount,
  'totalcost': instance.totalcost,
  'pal': instance.pal,
  'per_pal': instance.perPal,
};

CreditorName _$CreditorNameFromJson(Map<String, dynamic> json) => CreditorName(
  code: json['code'] as String?,
  name: json['name'] as String?,
  isauto: json['isauto'] as bool?,
  isdelete: json['isdelete'] as bool?,
);

Map<String, dynamic> _$CreditorNameToJson(CreditorName instance) =>
    <String, dynamic>{
      'code': instance.code,
      'name': instance.name,
      'isauto': instance.isauto,
      'isdelete': instance.isdelete,
    };
