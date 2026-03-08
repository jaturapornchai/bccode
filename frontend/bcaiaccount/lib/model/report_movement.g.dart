// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'report_movement.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ReportMovementModel _$ReportMovementModelFromJson(Map<String, dynamic> json) =>
    ReportMovementModel(
      balance: (json['balance'] as num?)?.toDouble(),
      details: (json['details'] as List<dynamic>?)
          ?.map((e) => ReportMovementDetail.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$ReportMovementModelToJson(
  ReportMovementModel instance,
) => <String, dynamic>{
  'balance': instance.balance,
  'details': instance.details?.map((e) => e.toJson()).toList(),
};

ReportMovementDetail _$ReportMovementDetailFromJson(
  Map<String, dynamic> json,
) => ReportMovementDetail(
  docdate: json['docdate'] as String?,
  docno: json['docno'] as String?,
  barcode: json['barcode'] as String?,
  qty: json['qty'] as String?,
  transflag: (json['transflag'] as num?)?.toInt(),
  calcflag: (json['calcflag'] as num?)?.toInt(),
  balance: (json['balance'] as num?)?.toDouble(),
);

Map<String, dynamic> _$ReportMovementDetailToJson(
  ReportMovementDetail instance,
) => <String, dynamic>{
  'docdate': instance.docdate,
  'docno': instance.docno,
  'barcode': instance.barcode,
  'qty': instance.qty,
  'transflag': instance.transflag,
  'calcflag': instance.calcflag,
  'balance': instance.balance,
};
