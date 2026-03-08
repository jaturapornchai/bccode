// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'currency_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ExchangeRateEntryModel _$ExchangeRateEntryModelFromJson(
  Map<String, dynamic> json,
) => ExchangeRateEntryModel(
  guidfixed: json['guidfixed'] as String?,
  date: json['date'] as String?,
  rate: (json['rate'] as num?)?.toDouble(),
);

Map<String, dynamic> _$ExchangeRateEntryModelToJson(
  ExchangeRateEntryModel instance,
) => <String, dynamic>{
  'guidfixed': instance.guidfixed,
  'date': instance.date,
  'rate': instance.rate,
};

CurrencyModel _$CurrencyModelFromJson(Map<String, dynamic> json) =>
    CurrencyModel(
      guidfixed: json['guidfixed'] as String?,
      code: json['code'] as String?,
      name: json['name'] as String?,
      symbol: json['symbol'] as String?,
      isdisabled: json['isdisabled'] as bool?,
      exchangeRates: (json['exchange_rates'] as List<dynamic>?)
          ?.map(
            (e) => ExchangeRateEntryModel.fromJson(e as Map<String, dynamic>),
          )
          .toList(),
    );

Map<String, dynamic> _$CurrencyModelToJson(CurrencyModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'code': instance.code,
      'name': instance.name,
      'symbol': instance.symbol,
      'isdisabled': instance.isdisabled,
      'exchange_rates': instance.exchangeRates.map((e) => e.toJson()).toList(),
    };

ExchangeRateHistoryModel _$ExchangeRateHistoryModelFromJson(
  Map<String, dynamic> json,
) => ExchangeRateHistoryModel(
  guidfixed: json['guidfixed'] as String?,
  currency: json['currency'] as String?,
  date: json['date'] as String?,
  rate: (json['rate'] as num?)?.toDouble(),
);

Map<String, dynamic> _$ExchangeRateHistoryModelToJson(
  ExchangeRateHistoryModel instance,
) => <String, dynamic>{
  'guidfixed': instance.guidfixed,
  'currency': instance.currency,
  'date': instance.date,
  'rate': instance.rate,
};
