// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'cash_drawer_models.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ValidationResult _$ValidationResultFromJson(
  Map<String, dynamic> json,
) => ValidationResult(
  isValid: json['isValid'] as bool? ?? true,
  errors:
      (json['errors'] as List<dynamic>?)?.map((e) => e as String).toList() ??
      const [],
  warnings:
      (json['warnings'] as List<dynamic>?)?.map((e) => e as String).toList() ??
      const [],
);

Map<String, dynamic> _$ValidationResultToJson(ValidationResult instance) =>
    <String, dynamic>{
      'isValid': instance.isValid,
      'errors': instance.errors,
      'warnings': instance.warnings,
    };

CashDrawerTransaction _$CashDrawerTransactionFromJson(
  Map<String, dynamic> json,
) => CashDrawerTransaction(
  guidfixed: json['guidfixed'] as String?,
  usercode: json['usercode'] as String,
  username: json['username'] as String,
  posid: json['posid'] as String,
  docno: json['docno'] as String,
  doctype: $enumDecode(_$CashDrawerTransactionTypeEnumMap, json['doctype']),
  docdate: DateTime.parse(json['docdate'] as String),
  remark: json['remark'] as String?,
  amount: (json['amount'] as num).toDouble(),
  paymentBreakdown: PaymentBreakdown.fromJson(
    json['paymentBreakdown'] as Map<String, dynamic>,
  ),
  billDetails: (json['billDetails'] as List<dynamic>?)
      ?.map((e) => TransactionModel.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$CashDrawerTransactionToJson(
  CashDrawerTransaction instance,
) => <String, dynamic>{
  'guidfixed': instance.guidfixed,
  'usercode': instance.usercode,
  'username': instance.username,
  'posid': instance.posid,
  'docno': instance.docno,
  'doctype': _$CashDrawerTransactionTypeEnumMap[instance.doctype]!,
  'docdate': instance.docdate.toIso8601String(),
  'remark': instance.remark,
  'amount': instance.amount,
  'paymentBreakdown': instance.paymentBreakdown.toJson(),
  'billDetails': instance.billDetails?.map((e) => e.toJson()).toList(),
};

const _$CashDrawerTransactionTypeEnumMap = {
  CashDrawerTransactionType.openShift: 1,
  CashDrawerTransactionType.closeShift: 2,
  CashDrawerTransactionType.addCash: 3,
  CashDrawerTransactionType.withdrawCash: 4,
};

PaymentBreakdown _$PaymentBreakdownFromJson(Map<String, dynamic> json) =>
    PaymentBreakdown(
      cash: (json['cash'] as num?)?.toDouble() ?? 0.0,
      creditCard: (json['creditCard'] as num?)?.toDouble() ?? 0.0,
      promptPay: (json['promptPay'] as num?)?.toDouble() ?? 0.0,
      transfer: (json['transfer'] as num?)?.toDouble() ?? 0.0,
      cheque: (json['cheque'] as num?)?.toDouble() ?? 0.0,
      coupon: (json['coupon'] as num?)?.toDouble() ?? 0.0,
    );

Map<String, dynamic> _$PaymentBreakdownToJson(PaymentBreakdown instance) =>
    <String, dynamic>{
      'cash': instance.cash,
      'creditCard': instance.creditCard,
      'promptPay': instance.promptPay,
      'transfer': instance.transfer,
      'cheque': instance.cheque,
      'coupon': instance.coupon,
    };

ShiftSummary _$ShiftSummaryFromJson(Map<String, dynamic> json) => ShiftSummary(
  docno: json['docno'] as String,
  usercode: json['usercode'] as String,
  username: json['username'] as String,
  posid: json['posid'] as String,
  openShift: json['openShift'] == null
      ? null
      : CashDrawerTransaction.fromJson(
          json['openShift'] as Map<String, dynamic>,
        ),
  closeShift: json['closeShift'] == null
      ? null
      : CashDrawerTransaction.fromJson(
          json['closeShift'] as Map<String, dynamic>,
        ),
  transactions: (json['transactions'] as List<dynamic>)
      .map((e) => CashDrawerTransaction.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$ShiftSummaryToJson(ShiftSummary instance) =>
    <String, dynamic>{
      'docno': instance.docno,
      'usercode': instance.usercode,
      'username': instance.username,
      'posid': instance.posid,
      'openShift': instance.openShift?.toJson(),
      'closeShift': instance.closeShift?.toJson(),
      'transactions': instance.transactions.map((e) => e.toJson()).toList(),
    };

CashDrawerDashboard _$CashDrawerDashboardFromJson(Map<String, dynamic> json) =>
    CashDrawerDashboard(
      totalShifts: (json['totalShifts'] as num?)?.toInt() ?? 0,
      openShifts: (json['openShifts'] as num?)?.toInt() ?? 0,
      closedShifts: (json['closedShifts'] as num?)?.toInt() ?? 0,
      totalAmount: (json['totalAmount'] as num?)?.toDouble() ?? 0.0,
      totalCashIn: (json['totalCashIn'] as num?)?.toDouble() ?? 0.0,
      totalCashOut: (json['totalCashOut'] as num?)?.toDouble() ?? 0.0,
      validationIssues:
          (json['validationIssues'] as List<dynamic>?)
              ?.map((e) => ValidationResult.fromJson(e as Map<String, dynamic>))
              .toList() ??
          const [],
    );

Map<String, dynamic> _$CashDrawerDashboardToJson(
  CashDrawerDashboard instance,
) => <String, dynamic>{
  'totalShifts': instance.totalShifts,
  'openShifts': instance.openShifts,
  'closedShifts': instance.closedShifts,
  'totalAmount': instance.totalAmount,
  'totalCashIn': instance.totalCashIn,
  'totalCashOut': instance.totalCashOut,
  'validationIssues': instance.validationIssues,
};
