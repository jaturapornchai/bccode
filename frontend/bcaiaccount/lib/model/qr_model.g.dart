// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'qr_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

QrModel _$QrModelFromJson(Map<String, dynamic> json) => QrModel(
  guidfixed: json['guidfixed'] as String?,
  code: json['code'] as String?,
  qrnames: (json['qrnames'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  qrtype: (json['qrtype'] as num?)?.toInt(),
  isactive: json['isactive'] as bool?,
  logo: json['logo'] as String?,
  bankcode: json['bankcode'] as String?,
  banknames: (json['banknames'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  bookbankcode: json['bookbankcode'] as String?,
  bookbanknames: (json['bookbanknames'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  bookbankimages: (json['bookbankimages'] as List<dynamic>?)
      ?.map((e) => ImagesModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  qrcode: json['qrcode'] as String?,
  apikey: json['apikey'] as String?,
  billerCode: json['billerCode'] as String?,
  billerID: json['billerID'] as String?,
  storeID: json['storeID'] as String?,
  terminalID: json['terminalID'] as String?,
  merchantName: json['merchantName'] as String?,
  accessCode: json['accessCode'] as String?,
  bankcharge: json['bankcharge'] as String?,
  customercharge: json['customercharge'] as String?,
  closeqr: (json['closeqr'] as num?)?.toInt(),
  secert: json['secert'] as String?,
  token: json['token'] as String?,
  appid: json['appid'] as String?,
  host: json['host'] as String?,
);

Map<String, dynamic> _$QrModelToJson(QrModel instance) => <String, dynamic>{
  'guidfixed': instance.guidfixed,
  'code': instance.code,
  'qrnames': instance.qrnames?.map((e) => e.toJson()).toList(),
  'qrtype': instance.qrtype,
  'isactive': instance.isactive,
  'logo': instance.logo,
  'bankcode': instance.bankcode,
  'banknames': instance.banknames?.map((e) => e.toJson()).toList(),
  'bookbankcode': instance.bookbankcode,
  'bookbanknames': instance.bookbanknames?.map((e) => e.toJson()).toList(),
  'bookbankimages': instance.bookbankimages?.map((e) => e.toJson()).toList(),
  'qrcode': instance.qrcode,
  'apikey': instance.apikey,
  'billerCode': instance.billerCode,
  'billerID': instance.billerID,
  'storeID': instance.storeID,
  'terminalID': instance.terminalID,
  'merchantName': instance.merchantName,
  'accessCode': instance.accessCode,
  'bankcharge': instance.bankcharge,
  'customercharge': instance.customercharge,
  'secert': instance.secert,
  'token': instance.token,
  'closeqr': instance.closeqr,
  'appid': instance.appid,
  'host': instance.host,
};

QrModelList _$QrModelListFromJson(Map<String, dynamic> json) => QrModelList(
  code: json['code'] as String,
  qrlist: (json['qrlist'] as List<dynamic>)
      .map((e) => QrModel.fromJson(e as Map<String, dynamic>))
      .toList(),
);

Map<String, dynamic> _$QrModelListToJson(QrModelList instance) =>
    <String, dynamic>{
      'code': instance.code,
      'qrlist': instance.qrlist.map((e) => e.toJson()).toList(),
    };
