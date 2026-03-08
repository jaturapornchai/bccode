// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'order_setting_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

OrderSettingModel _$OrderSettingModelFromJson(Map<String, dynamic> json) =>
    OrderSettingModel(
      guidfixed: json['guidfixed'] as String?,
      activepin: json['activepin'] as String?,
      code: json['code'] as String?,
      devicenumber: json['devicenumber'] as String?,
      devicetype: (json['devicetype'] as num?)?.toInt(),
      docformat: json['docformat'] as String?,
      isposactive: json['isposactive'] as bool?,
      settingcode: json['settingcode'] as String?,
      settingname: json['settingname'] as String?,
    );

Map<String, dynamic> _$OrderSettingModelToJson(OrderSettingModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'activepin': instance.activepin,
      'code': instance.code,
      'devicenumber': instance.devicenumber,
      'devicetype': instance.devicetype,
      'docformat': instance.docformat,
      'isposactive': instance.isposactive,
      'settingcode': instance.settingcode,
      'settingname': instance.settingname,
    };
