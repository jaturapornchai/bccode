// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'config_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

DayWorkTimeModel _$DayWorkTimeModelFromJson(Map<String, dynamic> json) =>
    DayWorkTimeModel(
      dayNumber: (json['dayNumber'] as num).toInt(),
      openTime: (json['openTime'] as List<dynamic>)
          .map((e) => e as String)
          .toList(),
      closeTime: (json['closeTime'] as List<dynamic>)
          .map((e) => e as String)
          .toList(),
      isClose: json['isClose'] as bool,
    );

Map<String, dynamic> _$DayWorkTimeModelToJson(DayWorkTimeModel instance) =>
    <String, dynamic>{
      'dayNumber': instance.dayNumber,
      'openTime': instance.openTime,
      'closeTime': instance.closeTime,
      'isClose': instance.isClose,
    };

DayCloseModel _$DayCloseModelFromJson(Map<String, dynamic> json) =>
    DayCloseModel(
      dates: (json['dates'] as List<dynamic>).map((e) => e as String).toList(),
      openTime: (json['openTime'] as List<dynamic>)
          .map((e) => e as String)
          .toList(),
      closeTime: (json['closeTime'] as List<dynamic>)
          .map((e) => e as String)
          .toList(),
      isClose: json['isClose'] as bool,
    );

Map<String, dynamic> _$DayCloseModelToJson(DayCloseModel instance) =>
    <String, dynamic>{
      'dates': instance.dates,
      'openTime': instance.openTime,
      'closeTime': instance.closeTime,
      'isClose': instance.isClose,
    };

ConfigSystemModel _$ConfigSystemModelFromJson(Map<String, dynamic> json) =>
    ConfigSystemModel(
      languageList: (json['languageList'] as List<dynamic>)
          .map((e) => e as String)
          .toList(),
      vatrate: (json['vatrate'] as num?)?.toDouble(),
      vattypesale: (json['vattypesale'] as num?)?.toInt(),
      vattypepurchase: (json['vattypepurchase'] as num?)?.toInt(),
      inquirytypesale: (json['inquirytypesale'] as num?)?.toInt(),
      inquirytypepurchase: (json['inquirytypepurchase'] as num?)?.toInt(),
      headerreceiptpos: json['headerreceiptpos'] as String?,
      footerreceiptpos: json['footerreceiptpos'] as String?,
    );

Map<String, dynamic> _$ConfigSystemModelToJson(ConfigSystemModel instance) =>
    <String, dynamic>{
      'languageList': instance.languageList,
      'vatrate': instance.vatrate,
      'vattypesale': instance.vattypesale,
      'vattypepurchase': instance.vattypepurchase,
      'inquirytypesale': instance.inquirytypesale,
      'inquirytypepurchase': instance.inquirytypepurchase,
      'headerreceiptpos': instance.headerreceiptpos,
      'footerreceiptpos': instance.footerreceiptpos,
    };

CompanyModel _$CompanyModelFromJson(Map<String, dynamic> json) => CompanyModel(
  names: (json['names'] as List<dynamic>)
      .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  taxID: json['taxID'] as String,
  branchNames: (json['branchNames'] as List<dynamic>)
      .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  addresses: (json['addresses'] as List<dynamic>)
      .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  phones: (json['phones'] as List<dynamic>).map((e) => e as String).toList(),
  emailOwners: (json['emailOwners'] as List<dynamic>)
      .map((e) => e as String)
      .toList(),
  emailStaffs: (json['emailStaffs'] as List<dynamic>)
      .map((e) => e as String)
      .toList(),
  latitude: json['latitude'] as String,
  longitude: json['longitude'] as String,
  usebranch: json['usebranch'] as bool?,
  usedepartment: json['usedepartment'] as bool?,
  images: (json['images'] as List<dynamic>)
      .map((e) => ImagesModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  logo: json['logo'] as String?,
);

Map<String, dynamic> _$CompanyModelToJson(CompanyModel instance) =>
    <String, dynamic>{
      'names': instance.names.map((e) => e.toJson()).toList(),
      'taxID': instance.taxID,
      'branchNames': instance.branchNames.map((e) => e.toJson()).toList(),
      'addresses': instance.addresses.map((e) => e.toJson()).toList(),
      'phones': instance.phones,
      'emailOwners': instance.emailOwners,
      'emailStaffs': instance.emailStaffs,
      'latitude': instance.latitude,
      'longitude': instance.longitude,
      'usebranch': instance.usebranch,
      'usedepartment': instance.usedepartment,
      'images': instance.images.map((e) => e.toJson()).toList(),
      'logo': instance.logo,
    };

WorkingTimeModel _$WorkingTimeModelFromJson(Map<String, dynamic> json) =>
    WorkingTimeModel(
      dayWorkTimes: (json['dayWorkTimes'] as List<dynamic>)
          .map((e) => DayWorkTimeModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      dayCloses: (json['dayCloses'] as List<dynamic>)
          .map((e) => DayCloseModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$WorkingTimeModelToJson(WorkingTimeModel instance) =>
    <String, dynamic>{
      'dayWorkTimes': instance.dayWorkTimes.map((e) => e.toJson()).toList(),
      'dayCloses': instance.dayCloses.map((e) => e.toJson()).toList(),
    };

DeviceConfigModel _$DeviceConfigModelFromJson(Map<String, dynamic> json) =>
    DeviceConfigModel(
      listDataFontSize: (json['listDataFontSize'] as num).toDouble(),
      listDataLineSpace: (json['listDataLineSpace'] as num).toDouble(),
      itemDisplaySku: json['itemDisplaySku'] as bool,
      itemDisplayPrice: json['itemDisplayPrice'] as bool,
    );

Map<String, dynamic> _$DeviceConfigModelToJson(DeviceConfigModel instance) =>
    <String, dynamic>{
      'listDataFontSize': instance.listDataFontSize,
      'listDataLineSpace': instance.listDataLineSpace,
      'itemDisplaySku': instance.itemDisplaySku,
      'itemDisplayPrice': instance.itemDisplayPrice,
    };
