// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'report_product_balance_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ReportProductBalanceModel _$ReportProductBalanceModelFromJson(
  Map<String, dynamic> json,
) => ReportProductBalanceModel(
  barcode: json['barcode'] as String?,
  names: (json['names'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  unitcode: json['unitcode'] as String?,
  standunit: json['standunit'] as String?,
  balanceqty: json['balanceqty'] as String?,
  averagecost: json['averagecost'] as String?,
  balanceamount: json['balanceamount'] as String?,
);

Map<String, dynamic> _$ReportProductBalanceModelToJson(
  ReportProductBalanceModel instance,
) => <String, dynamic>{
  'barcode': instance.barcode,
  'names': instance.names?.map((e) => e.toJson()).toList(),
  'unitcode': instance.unitcode,
  'standunit': instance.standunit,
  'balanceqty': instance.balanceqty,
  'averagecost': instance.averagecost,
  'balanceamount': instance.balanceamount,
};
