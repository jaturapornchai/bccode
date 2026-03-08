// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'wht_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

WHTEntryModel _$WHTEntryModelFromJson(Map<String, dynamic> json) =>
    WHTEntryModel(
      description: json['description'] as String?,
      rate: (json['rate'] as num?)?.toDouble(),
      taxbase: (json['taxbase'] as num?)?.toDouble(),
      amount: (json['amount'] as num?)?.toDouble(),
      note: json['note'] as String?,
    );

Map<String, dynamic> _$WHTEntryModelToJson(WHTEntryModel instance) =>
    <String, dynamic>{
      'description': instance.description,
      'rate': instance.rate,
      'taxbase': instance.taxbase,
      'amount': instance.amount,
      'note': instance.note,
    };
