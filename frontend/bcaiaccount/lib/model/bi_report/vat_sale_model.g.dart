// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'vat_sale_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

VatSaleModel _$VatSaleModelFromJson(Map<String, dynamic> json) => VatSaleModel(
  taxdocdate: json['taxdocdate'] as String,
  taxdocno: json['taxdocno'] as String,
  docdate: json['docdate'] as String,
  docno: json['docno'] as String,
  creditorcode: json['creditorcode'] as String,
  creditorname: (json['creditorname'] as List<dynamic>)
      .map((e) => CreditorName.fromJson(e as Map<String, dynamic>))
      .toList(),
  branchcode: json['branchcode'] as String,
  taxid: json['taxid'] as String,
  totalbeforevat: (json['totalbeforevat'] as num).toDouble(),
  totalvatvalue: (json['totalvatvalue'] as num).toDouble(),
  totalexceptvat: (json['totalexceptvat'] as num).toDouble(),
  totalaftervat: (json['totalaftervat'] as num).toDouble(),
  detailtotaldiscount: (json['detailtotaldiscount'] as num).toDouble(),
  totalamount: (json['totalamount'] as num).toDouble(),
  status: json['status'] as String,
);

Map<String, dynamic> _$VatSaleModelToJson(VatSaleModel instance) =>
    <String, dynamic>{
      'taxdocdate': instance.taxdocdate,
      'taxdocno': instance.taxdocno,
      'docdate': instance.docdate,
      'docno': instance.docno,
      'creditorcode': instance.creditorcode,
      'creditorname': instance.creditorname,
      'branchcode': instance.branchcode,
      'taxid': instance.taxid,
      'totalbeforevat': instance.totalbeforevat,
      'totalvatvalue': instance.totalvatvalue,
      'totalexceptvat': instance.totalexceptvat,
      'totalaftervat': instance.totalaftervat,
      'detailtotaldiscount': instance.detailtotaldiscount,
      'totalamount': instance.totalamount,
      'status': instance.status,
    };

CreditorName _$CreditorNameFromJson(Map<String, dynamic> json) => CreditorName(
  code: json['code'] as String,
  name: json['name'] as String,
  isauto: json['isauto'] as bool,
  isdelete: json['isdelete'] as bool,
);

Map<String, dynamic> _$CreditorNameToJson(CreditorName instance) =>
    <String, dynamic>{
      'code': instance.code,
      'name': instance.name,
      'isauto': instance.isauto,
      'isdelete': instance.isdelete,
    };
