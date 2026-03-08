// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'order_template_setting_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

OrderTemplateSettingModel _$OrderTemplateSettingModelFromJson(
  Map<String, dynamic> json,
) => OrderTemplateSettingModel(
  guidfixed: json['guidfixed'] as String?,
  activepin: json['activepin'] as String?,
  branch: json['branch'] == null
      ? null
      : BranchModel.fromJson(json['branch'] as Map<String, dynamic>),
  code: json['code'] as String?,
  devicenumber: json['devicenumber'] as String?,
  devicetype: (json['devicetype'] as num?)?.toInt(),
  docformat: json['docformat'] as String?,
  logourl: json['logourl'] as String?,
  mediaguid: json['mediaguid'] as String?,
  tablenumber: json['tablenumber'] as String?,
  timeforsales: (json['timeforsales'] as List<dynamic>?)
      ?.map((e) => TimeForsaleModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  qrcodes: (json['qrcodes'] as List<dynamic>?)
      ?.map((e) => QrModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  orderdevices: (json['orderdevices'] as List<dynamic>?)
      ?.map((e) => OrderDeviceModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  label: json['label'] as String?,
  salechannels: (json['salechannels'] as List<dynamic>?)
      ?.map((e) => e as String)
      .toList(),
  adminpin: json['adminpin'] as String?,
  categorygroupnumber: (json['categorygroupnumber'] as num?)?.toInt(),
  kitchengroupnumber: (json['kitchengroupnumber'] as num?)?.toInt(),
  tablegroupnumber: (json['tablegroupnumber'] as num?)?.toInt(),
  zonegroupnumber: (json['zonegroupnumber'] as num?)?.toInt(),
  isvatregister: json['isvatregister'] as bool?,
  vattype: (json['vattype'] as num?)?.toInt(),
  vatrate: (json['vatrate'] as num?)?.toDouble(),
  lineoaimg: json['lineoaimg'] as String?,
);

Map<String, dynamic> _$OrderTemplateSettingModelToJson(
  OrderTemplateSettingModel instance,
) => <String, dynamic>{
  'guidfixed': instance.guidfixed,
  'activepin': instance.activepin,
  'branch': instance.branch,
  'code': instance.code,
  'devicenumber': instance.devicenumber,
  'devicetype': instance.devicetype,
  'docformat': instance.docformat,
  'logourl': instance.logourl,
  'mediaguid': instance.mediaguid,
  'tablenumber': instance.tablenumber,
  'timeforsales': instance.timeforsales,
  'qrcodes': instance.qrcodes,
  'orderdevices': instance.orderdevices,
  'label': instance.label,
  'salechannels': instance.salechannels,
  'adminpin': instance.adminpin,
  'categorygroupnumber': instance.categorygroupnumber,
  'kitchengroupnumber': instance.kitchengroupnumber,
  'tablegroupnumber': instance.tablegroupnumber,
  'zonegroupnumber': instance.zonegroupnumber,
  'isvatregister': instance.isvatregister,
  'vattype': instance.vattype,
  'vatrate': instance.vatrate,
  'lineoaimg': instance.lineoaimg,
};

BranchModel _$BranchModelFromJson(Map<String, dynamic> json) => BranchModel(
  code: json['code'] as String?,
  guidfixed: json['guidfixed'] as String?,
  names: (json['names'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$BranchModelToJson(BranchModel instance) =>
    <String, dynamic>{
      'code': instance.code,
      'guidfixed': instance.guidfixed,
      'names': instance.names,
    };

TimeForsaleModel _$TimeForsaleModelFromJson(Map<String, dynamic> json) =>
    TimeForsaleModel(
      from: json['from'] as String?,
      to: json['to'] as String?,
      names: (json['names'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$TimeForsaleModelToJson(TimeForsaleModel instance) =>
    <String, dynamic>{
      'from': instance.from,
      'to': instance.to,
      'names': instance.names,
    };

OrderDeviceModel _$OrderDeviceModelFromJson(Map<String, dynamic> json) =>
    OrderDeviceModel(
      activepin: json['activepin'] as String?,
      id: json['id'] as String?,
      names: (json['names'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$OrderDeviceModelToJson(OrderDeviceModel instance) =>
    <String, dynamic>{
      'activepin': instance.activepin,
      'id': instance.id,
      'names': instance.names,
    };
