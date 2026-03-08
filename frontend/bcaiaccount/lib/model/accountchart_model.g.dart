// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'accountchart_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

AccountChartModel _$AccountChartModelFromJson(Map<String, dynamic> json) =>
    AccountChartModel(
      shopid: json['shopid'] as String?,
      accountcode: json['accountcode'] as String?,
      guidfixed: json['guidfixed'] as String?,
      accountname: json['accountname'] as String?,
      accountcategory: (json['accountcategory'] as num?)?.toInt(),
      accountgroup: json['accountgroup'] as String?,
      accountlevel: (json['accountlevel'] as num?)?.toInt(),
      consolidateaccountcode: json['consolidateaccountcode'] as String?,
      accountbalancetype: (json['accountbalancetype'] as num?)?.toInt(),
      iscenterchart: json['iscenterchart'] as bool?,
    );

Map<String, dynamic> _$AccountChartModelToJson(AccountChartModel instance) =>
    <String, dynamic>{
      'shopid': instance.shopid,
      'accountcode': instance.accountcode,
      'guidfixed': instance.guidfixed,
      'accountname': instance.accountname,
      'accountcategory': instance.accountcategory,
      'accountgroup': instance.accountgroup,
      'accountlevel': instance.accountlevel,
      'consolidateaccountcode': instance.consolidateaccountcode,
      'accountbalancetype': instance.accountbalancetype,
      'iscenterchart': instance.iscenterchart,
    };
