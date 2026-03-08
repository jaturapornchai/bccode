// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'payment_daily_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PaymentDailyModel _$PaymentDailyModelFromJson(Map<String, dynamic> json) =>
    PaymentDailyModel(
      docDate: json['doc_date'] as String,
      totalAmount: _numToDouble(json['total_amount']),
      roundAmount: _numToDouble(json['round_amount']),
      totalValue: _numToDouble(json['total_value']),
      payCashAmount: _numToDouble(json['pay_cashamount']),
      sumTransfer: _numToDouble(json['sum_transfer']),
      sumCreditCard: _numToDouble(json['sum_creditcard']),
      sumCheque: _numToDouble(json['sum_cheque']),
      sumCoupon: _numToDouble(json['sum_coupon']),
      sumQRCode: _numToDouble(json['sum_qrcode']),
      sumCredit: _numToDouble(json['sum_credit']),
      totalPayment: _numToDouble(json['total_payment']),
      transactions:
          (json['transactions'] as List<dynamic>?)
              ?.map((e) => TransactionModel.fromJson(e as Map<String, dynamic>))
              .toList() ??
          const [],
    );

Map<String, dynamic> _$PaymentDailyModelToJson(PaymentDailyModel instance) =>
    <String, dynamic>{
      'doc_date': instance.docDate,
      'total_amount': instance.totalAmount,
      'round_amount': instance.roundAmount,
      'total_value': instance.totalValue,
      'pay_cashamount': instance.payCashAmount,
      'sum_transfer': instance.sumTransfer,
      'sum_creditcard': instance.sumCreditCard,
      'sum_cheque': instance.sumCheque,
      'sum_coupon': instance.sumCoupon,
      'sum_qrcode': instance.sumQRCode,
      'sum_credit': instance.sumCredit,
      'total_payment': instance.totalPayment,
      'transactions': instance.transactions?.map((e) => e.toJson()).toList(),
    };

PaymentDailySummaryModel _$PaymentDailySummaryModelFromJson(
  Map<String, dynamic> json,
) => PaymentDailySummaryModel(
  fromDate: json['fromdate'] as String,
  toDate: json['todate'] as String,
  totalAmount: _numToDouble(json['total_amount']),
  totalPayment: _numToDouble(json['total_payment']),
);

Map<String, dynamic> _$PaymentDailySummaryModelToJson(
  PaymentDailySummaryModel instance,
) => <String, dynamic>{
  'fromdate': instance.fromDate,
  'todate': instance.toDate,
  'total_amount': instance.totalAmount,
  'total_payment': instance.totalPayment,
};

TransactionModel _$TransactionModelFromJson(Map<String, dynamic> json) =>
    TransactionModel(
      docDate: _stringFromJson(json['docdate']),
      docTime: _stringFromJson(json['doctime']),
      docNo: _stringFromJson(json['docno']),
      custName: _stringFromJson(json['custname']),
      totalAmount: _numToDouble(json['totalamount']),
      roundAmount: _numToDouble(json['roundamount']),
      totalValue: _numToDouble(json['totalvalue']),
      payCashAmount: _numToDouble(json['paycashamount']),
      sumMoneyTransfer: _numToDouble(json['summoneytransfer']),
      sumCreditCard: _numToDouble(json['sumcreditcard']),
      sumCheque: _numToDouble(json['sumcheque']),
      sumCoupon: _numToDouble(json['sumcoupon']),
      sumQRCode: _numToDouble(json['sumqrcode']),
      sumCredit: _numToDouble(json['sumcredit']),
      totalPayment: _numToDouble(json['totalpayment']),
    );

Map<String, dynamic> _$TransactionModelToJson(TransactionModel instance) =>
    <String, dynamic>{
      'docdate': instance.docDate,
      'doctime': instance.docTime,
      'docno': instance.docNo,
      'custname': instance.custName,
      'totalamount': instance.totalAmount,
      'roundamount': instance.roundAmount,
      'totalvalue': instance.totalValue,
      'paycashamount': instance.payCashAmount,
      'summoneytransfer': instance.sumMoneyTransfer,
      'sumcreditcard': instance.sumCreditCard,
      'sumcheque': instance.sumCheque,
      'sumcoupon': instance.sumCoupon,
      'sumqrcode': instance.sumQRCode,
      'sumcredit': instance.sumCredit,
      'totalpayment': instance.totalPayment,
    };
