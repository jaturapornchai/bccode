// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'report_main_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ReportSaleByDateModel _$ReportSaleByDateModelFromJson(
  Map<String, dynamic> json,
) => ReportSaleByDateModel(
  docdatetime: json['docdatetime'] as String?,
  detailtotalamount: (json['detailtotalamount'] as num?)?.toDouble(),
  totaldiscount: (json['totaldiscount'] as num?)?.toDouble(),
  totalexceptvat: (json['totalexceptvat'] as num?)?.toDouble(),
  totalbeforevat: (json['totalbeforevat'] as num?)?.toDouble(),
  totalvatvalue: (json['totalvatvalue'] as num?)?.toDouble(),
  totalamount: (json['totalamount'] as num?)?.toDouble(),
);

Map<String, dynamic> _$ReportSaleByDateModelToJson(
  ReportSaleByDateModel instance,
) => <String, dynamic>{
  'docdatetime': instance.docdatetime,
  'detailtotalamount': instance.detailtotalamount,
  'totaldiscount': instance.totaldiscount,
  'totalexceptvat': instance.totalexceptvat,
  'totalbeforevat': instance.totalbeforevat,
  'totalvatvalue': instance.totalvatvalue,
  'totalamount': instance.totalamount,
};

ReportReceiveMoneyModel _$ReportReceiveMoneyModelFromJson(
  Map<String, dynamic> json,
) => ReportReceiveMoneyModel(
  date: json['date'] as String?,
  data: json['data'] == null
      ? null
      : ReportReceiveMoneyDetailModel.fromJson(
          json['data'] as Map<String, dynamic>,
        ),
);

Map<String, dynamic> _$ReportReceiveMoneyModelToJson(
  ReportReceiveMoneyModel instance,
) => <String, dynamic>{'date': instance.date, 'data': instance.data.toJson()};

ReportReceiveMoneyDetailModel _$ReportReceiveMoneyDetailModelFromJson(
  Map<String, dynamic> json,
) => ReportReceiveMoneyDetailModel(
  cashAmount: (json['cashAmount'] as num?)?.toDouble(),
  creditAmount: (json['creditAmount'] as num?)?.toDouble(),
  transferAmount: (json['transferAmount'] as num?)?.toDouble(),
  couponAmount: (json['couponAmount'] as num?)?.toDouble(),
  chequeAmount: (json['chequeAmount'] as num?)?.toDouble(),
  totalAmount: (json['totalAmount'] as num?)?.toDouble(),
);

Map<String, dynamic> _$ReportReceiveMoneyDetailModelToJson(
  ReportReceiveMoneyDetailModel instance,
) => <String, dynamic>{
  'cashAmount': instance.cashAmount,
  'creditAmount': instance.creditAmount,
  'transferAmount': instance.transferAmount,
  'couponAmount': instance.couponAmount,
  'chequeAmount': instance.chequeAmount,
  'totalAmount': instance.totalAmount,
};
