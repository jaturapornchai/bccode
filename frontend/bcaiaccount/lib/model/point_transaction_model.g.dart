// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'point_transaction_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PointTransactionModel _$PointTransactionModelFromJson(
  Map<String, dynamic> json,
) => PointTransactionModel(
  guidfixed: json['guidfixed'] as String,
  transactiondocno: json['transactiondocno'] as String,
  transactiondate: DateTime.parse(json['transactiondate'] as String),
  debtorcode: json['debtorcode'] as String,
  pointscode: json['pointscode'] as String,
  transactiontype: (json['transactiontype'] as num).toInt(),
  pointamount: (json['pointamount'] as num).toInt(),
  balancebefore: (json['balancebefore'] as num).toInt(),
  balanceafter: (json['balanceafter'] as num).toInt(),
  description: json['description'] as String,
);

Map<String, dynamic> _$PointTransactionModelToJson(
  PointTransactionModel instance,
) => <String, dynamic>{
  'guidfixed': instance.guidfixed,
  'transactiondocno': instance.transactiondocno,
  'transactiondate': instance.transactiondate.toIso8601String(),
  'debtorcode': instance.debtorcode,
  'pointscode': instance.pointscode,
  'transactiontype': instance.transactiontype,
  'pointamount': instance.pointamount,
  'balancebefore': instance.balancebefore,
  'balanceafter': instance.balanceafter,
  'description': instance.description,
};

PointTransactionCreateModel _$PointTransactionCreateModelFromJson(
  Map<String, dynamic> json,
) => PointTransactionCreateModel(
  pointscode: json['pointscode'] as String,
  pointamount: (json['pointamount'] as num).toInt(),
  description: json['description'] as String,
);

Map<String, dynamic> _$PointTransactionCreateModelToJson(
  PointTransactionCreateModel instance,
) => <String, dynamic>{
  'pointscode': instance.pointscode,
  'pointamount': instance.pointamount,
  'description': instance.description,
};
