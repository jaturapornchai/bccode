import 'package:json_annotation/json_annotation.dart';

part 'currency_model.g.dart';

@JsonSerializable(explicitToJson: true)
class ExchangeRateEntryModel {
  String guidfixed;
  String date;
  double rate;

  ExchangeRateEntryModel({
    String? guidfixed,
    String? date,
    double? rate,
  })  : guidfixed = guidfixed ?? "",
        date = date ?? "",
        rate = rate ?? 0.0;

  factory ExchangeRateEntryModel.fromJson(Map<String, dynamic> json) =>
      _$ExchangeRateEntryModelFromJson(json);

  Map<String, dynamic> toJson() => _$ExchangeRateEntryModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class CurrencyModel {
  String guidfixed;
  String code;
  String name;
  String symbol;
  bool isdisabled;

  @JsonKey(name: 'exchange_rates')
  List<ExchangeRateEntryModel> exchangeRates;

  CurrencyModel({
    String? guidfixed,
    String? code,
    String? name,
    String? symbol,
    bool? isdisabled,
    List<ExchangeRateEntryModel>? exchangeRates,
  })  : guidfixed = guidfixed ?? "",
        code = code ?? "",
        name = name ?? "",
        symbol = symbol ?? "",
        isdisabled = isdisabled ?? false,
        exchangeRates = exchangeRates ?? [];

  factory CurrencyModel.fromJson(Map<String, dynamic> json) =>
      _$CurrencyModelFromJson(json);

  Map<String, dynamic> toJson() => _$CurrencyModelToJson(this);
}

// For backward compatibility with existing code
@JsonSerializable(explicitToJson: true)
class ExchangeRateHistoryModel {
  @JsonKey(includeIfNull: false)
  String guidfixed;
  String currency;
  String date;
  double rate;

  ExchangeRateHistoryModel({
    String? guidfixed,
    String? currency,
    String? date,
    double? rate,
  })  : guidfixed = guidfixed ?? "",
        currency = currency ?? "",
        date = date ?? "",
        rate = rate ?? 0.0;

  factory ExchangeRateHistoryModel.fromJson(Map<String, dynamic> json) =>
      _$ExchangeRateHistoryModelFromJson(json);

  Map<String, dynamic> toJson() => _$ExchangeRateHistoryModelToJson(this);
}
