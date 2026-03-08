// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'shop_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ShopModel _$ShopModelFromJson(Map<String, dynamic> json) => ShopModel(
  guidfixed: json['guidfixed'] as String?,
  address: (json['address'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  branchcode: json['branchcode'] as String?,
  images: (json['images'] as List<dynamic>?)
      ?.map((e) => ImagesModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  logo: json['logo'] as String?,
  name1: json['name1'] as String?,
  names: (json['names'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  profilepicture: json['profilepicture'] as String?,
  settings: json['settings'] == null
      ? null
      : Settings.fromJson(json['settings'] as Map<String, dynamic>),
  telephone: json['telephone'] as String?,
  ismainshop: json['ismainshop'] as bool?,
  productcentertype: (json['productcentertype'] as num?)?.toInt(),
  debtorcentertype: (json['debtorcentertype'] as num?)?.toInt(),
  mainshopid: json['mainshopid'] as String?,
  posproductcentertype: (json['posproductcentertype'] as num?)?.toInt(),
  apikey: json['apikey'] as String?,
);

Map<String, dynamic> _$ShopModelToJson(ShopModel instance) => <String, dynamic>{
  'guidfixed': instance.guidfixed,
  'address': instance.address,
  'branchcode': instance.branchcode,
  'images': instance.images,
  'logo': instance.logo,
  'name1': instance.name1,
  'names': instance.names,
  'profilepicture': instance.profilepicture,
  'settings': instance.settings,
  'telephone': instance.telephone,
  'ismainshop': instance.ismainshop,
  'productcentertype': instance.productcentertype,
  'debtorcentertype': instance.debtorcentertype,
  'mainshopid': instance.mainshopid,
  'posproductcentertype': instance.posproductcentertype,
  'apikey': instance.apikey,
};
