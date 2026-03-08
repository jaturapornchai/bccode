// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'purchase_partial_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PurchasePartialCreditorName _$PurchasePartialCreditorNameFromJson(
  Map<String, dynamic> json,
) => PurchasePartialCreditorName(
  code: json['code'] as String,
  name: json['name'] as String,
  isauto: json['isauto'] as bool,
  isdelete: json['isdelete'] as bool,
);

Map<String, dynamic> _$PurchasePartialCreditorNameToJson(
  PurchasePartialCreditorName instance,
) => <String, dynamic>{
  'code': instance.code,
  'name': instance.name,
  'isauto': instance.isauto,
  'isdelete': instance.isdelete,
};

PurchasePartialBranchName _$PurchasePartialBranchNameFromJson(
  Map<String, dynamic> json,
) => PurchasePartialBranchName(
  code: json['code'] as String,
  name: json['name'] as String,
  isauto: json['isauto'] as bool,
  isdelete: json['isdelete'] as bool,
);

Map<String, dynamic> _$PurchasePartialBranchNameToJson(
  PurchasePartialBranchName instance,
) => <String, dynamic>{
  'code': instance.code,
  'name': instance.name,
  'isauto': instance.isauto,
  'isdelete': instance.isdelete,
};

PurchasePartialModel _$PurchasePartialModelFromJson(
  Map<String, dynamic> json,
) => PurchasePartialModel(
  guidfixed: json['guidfixed'] as String,
  docdate: json['docdate'] as String,
  docno: json['docno'] as String,
  docrefno: json['docrefno'] as String,
  creditorcode: json['creditorcode'] as String,
  creditornames: (json['creditornames'] as List<dynamic>)
      .map(
        (e) => PurchasePartialCreditorName.fromJson(e as Map<String, dynamic>),
      )
      .toList(),
  totalvalue: (json['totalvalue'] as num).toDouble(),
  totalexceptvat: (json['totalexceptvat'] as num).toDouble(),
  totalbeforevat: (json['totalbeforevat'] as num).toDouble(),
  totalvatvalue: (json['totalvatvalue'] as num).toDouble(),
  totalamount: (json['totalamount'] as num).toDouble(),
  branchcode: json['branchcode'] as String,
  branchnames: (json['branchnames'] as List<dynamic>)
      .map((e) => PurchasePartialBranchName.fromJson(e as Map<String, dynamic>))
      .toList(),
  transactions: json['transactions'] as List<dynamic>,
);

Map<String, dynamic> _$PurchasePartialModelToJson(
  PurchasePartialModel instance,
) => <String, dynamic>{
  'guidfixed': instance.guidfixed,
  'docdate': instance.docdate,
  'docno': instance.docno,
  'docrefno': instance.docrefno,
  'creditorcode': instance.creditorcode,
  'creditornames': instance.creditornames,
  'totalvalue': instance.totalvalue,
  'totalexceptvat': instance.totalexceptvat,
  'totalbeforevat': instance.totalbeforevat,
  'totalvatvalue': instance.totalvatvalue,
  'totalamount': instance.totalamount,
  'branchcode': instance.branchcode,
  'branchnames': instance.branchnames,
  'transactions': instance.transactions,
};
