// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'purchase_partial_summary_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PurchasePartialSummaryModel _$PurchasePartialSummaryModelFromJson(
  Map<String, dynamic> json,
) => PurchasePartialSummaryModel(
  fromDate: json['fromdate'] as String?,
  toDate: json['todate'] as String?,
  totalRecords: (json['total_records'] as num?)?.toInt(),
  totalValue: (json['total_value'] as num?)?.toDouble(),
  totalExceptVat: (json['total_except_vat'] as num?)?.toDouble(),
  totalBeforeVat: (json['total_before_vat'] as num?)?.toDouble(),
  totalVatValue: (json['total_vat_value'] as num?)?.toDouble(),
  totalAmount: (json['total_amount'] as num?)?.toDouble(),
);

Map<String, dynamic> _$PurchasePartialSummaryModelToJson(
  PurchasePartialSummaryModel instance,
) => <String, dynamic>{
  'fromdate': instance.fromDate,
  'todate': instance.toDate,
  'total_records': instance.totalRecords,
  'total_value': instance.totalValue,
  'total_except_vat': instance.totalExceptVat,
  'total_before_vat': instance.totalBeforeVat,
  'total_vat_value': instance.totalVatValue,
  'total_amount': instance.totalAmount,
};
