import 'package:json_annotation/json_annotation.dart';

part 'payment_daily_model.g.dart';

// Helper function to convert null or num to double
double _numToDouble(dynamic value) {
  if (value == null) return 0.0;
  if (value is int) return value.toDouble();
  if (value is double) return value;
  if (value is String) return double.tryParse(value) ?? 0.0;
  return 0.0;
}

// Helper function to convert null to empty string
String _stringFromJson(dynamic value) {
  if (value == null) return '';
  return value.toString();
}

@JsonSerializable(explicitToJson: true)
class PaymentDailyModel {
  @JsonKey(name: 'doc_date')
  final String docDate;
  @JsonKey(name: 'total_amount', fromJson: _numToDouble)
  final double totalAmount;
  @JsonKey(name: 'round_amount', fromJson: _numToDouble)
  final double roundAmount;
  @JsonKey(name: 'total_value', fromJson: _numToDouble)
  final double totalValue;
  @JsonKey(name: 'pay_cashamount', fromJson: _numToDouble)
  final double payCashAmount;
  @JsonKey(name: 'sum_transfer', fromJson: _numToDouble)
  final double sumTransfer;
  @JsonKey(name: 'sum_creditcard', fromJson: _numToDouble)
  final double sumCreditCard;
  @JsonKey(name: 'sum_cheque', fromJson: _numToDouble)
  final double sumCheque;
  @JsonKey(name: 'sum_coupon', fromJson: _numToDouble)
  final double sumCoupon;
  @JsonKey(name: 'sum_qrcode', fromJson: _numToDouble)
  final double sumQRCode;
  @JsonKey(name: 'sum_credit', fromJson: _numToDouble)
  final double sumCredit;
  @JsonKey(name: 'total_payment', fromJson: _numToDouble)
  final double totalPayment;
  @JsonKey(name: 'transactions')
  final List<TransactionModel>? transactions;

  PaymentDailyModel({
    required this.docDate,
    required this.totalAmount,
    required this.roundAmount,
    required this.totalValue,
    required this.payCashAmount,
    required this.sumTransfer,
    required this.sumCreditCard,
    required this.sumCheque,
    required this.sumCoupon,
    required this.sumQRCode,
    required this.sumCredit,
    required this.totalPayment,
    this.transactions = const [],
  });

  factory PaymentDailyModel.fromJson(Map<String, dynamic> json) =>
      _$PaymentDailyModelFromJson(json);

  Map<String, dynamic> toJson() => _$PaymentDailyModelToJson(this);
}

/// PaymentDailySummaryModel
@JsonSerializable(explicitToJson: true)
class PaymentDailySummaryModel {
  @JsonKey(name: 'fromdate')
  final String fromDate;
  @JsonKey(name: 'todate')
  final String toDate;
  @JsonKey(name: 'total_amount', fromJson: _numToDouble)
  final double totalAmount;
  @JsonKey(name: 'total_payment', fromJson: _numToDouble)
  final double totalPayment;

  PaymentDailySummaryModel({
    required this.fromDate,
    required this.toDate,
    required this.totalAmount,
    required this.totalPayment,
  });

  factory PaymentDailySummaryModel.fromJson(Map<String, dynamic> json) =>
      _$PaymentDailySummaryModelFromJson(json);

  Map<String, dynamic> toJson() => _$PaymentDailySummaryModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class TransactionModel {
  @JsonKey(name: 'docdate', fromJson: _stringFromJson)
  final String docDate;
  @JsonKey(name: 'doctime', fromJson: _stringFromJson)
  final String docTime;
  @JsonKey(name: 'docno', fromJson: _stringFromJson)
  final String docNo;
  @JsonKey(name: 'custname', fromJson: _stringFromJson)
  final String custName;
  @JsonKey(name: 'totalamount', fromJson: _numToDouble)
  final double totalAmount;
  @JsonKey(name: 'roundamount', fromJson: _numToDouble)
  final double roundAmount;
  @JsonKey(name: 'totalvalue', fromJson: _numToDouble)
  final double totalValue;
  @JsonKey(name: 'paycashamount', fromJson: _numToDouble)
  final double payCashAmount;
  @JsonKey(name: 'summoneytransfer', fromJson: _numToDouble)
  final double sumMoneyTransfer;
  @JsonKey(name: 'sumcreditcard', fromJson: _numToDouble)
  final double sumCreditCard;
  @JsonKey(name: 'sumcheque', fromJson: _numToDouble)
  final double sumCheque;
  @JsonKey(name: 'sumcoupon', fromJson: _numToDouble)
  final double sumCoupon;
  @JsonKey(name: 'sumqrcode', fromJson: _numToDouble)
  final double sumQRCode;
  @JsonKey(name: 'sumcredit', fromJson: _numToDouble)
  final double sumCredit;
  @JsonKey(name: 'totalpayment', fromJson: _numToDouble)
  final double totalPayment;

  TransactionModel({
    required this.docDate,
    required this.docTime,
    required this.docNo,
    required this.custName,
    required this.totalAmount,
    required this.roundAmount,
    required this.totalValue,
    required this.payCashAmount,
    required this.sumMoneyTransfer,
    required this.sumCreditCard,
    required this.sumCheque,
    required this.sumCoupon,
    required this.sumQRCode,
    required this.sumCredit,
    required this.totalPayment,
  });

  factory TransactionModel.fromJson(Map<String, dynamic> json) =>
      _$TransactionModelFromJson(json);
  Map<String, dynamic> toJson() => _$TransactionModelToJson(this);
}
