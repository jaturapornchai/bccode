// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'shift_detail_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ShiftDetailModel _$ShiftDetailModelFromJson(Map<String, dynamic> json) =>
    ShiftDetailModel(
      usercode: json['usercode'] as String?,
      username: json['username'] as String?,
      posid: json['posid'] as String?,
      docno: json['docno'] as String?,
      doctype: (json['doctype'] as num?)?.toInt(),
      docdate: json['docdate'] as String?,
      remark: json['remark'] as String?,
      amount: (json['amount'] as num?)?.toDouble(),
      creditcard: (json['creditcard'] as num?)?.toDouble(),
      promptpay: (json['promptpay'] as num?)?.toDouble(),
      transfer: (json['transfer'] as num?)?.toDouble(),
      cheque: (json['cheque'] as num?)?.toDouble(),
      coupon: (json['coupon'] as num?)?.toDouble(),
    );

Map<String, dynamic> _$ShiftDetailModelToJson(ShiftDetailModel instance) =>
    <String, dynamic>{
      'usercode': instance.usercode,
      'username': instance.username,
      'posid': instance.posid,
      'docno': instance.docno,
      'doctype': instance.doctype,
      'docdate': instance.docdate,
      'remark': instance.remark,
      'amount': instance.amount,
      'creditcard': instance.creditcard,
      'promptpay': instance.promptpay,
      'transfer': instance.transfer,
      'cheque': instance.cheque,
      'coupon': instance.coupon,
    };
