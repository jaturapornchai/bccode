// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'customer_request_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

CustomerRequestModel _$CustomerRequestModelFromJson(
  Map<String, dynamic> json,
) => CustomerRequestModel(
  addressforbilling: CustomerAddressModel.fromJson(
    json['addressforbilling'] as Map<String, dynamic>,
  ),
  addressforshipping: (json['addressforshipping'] as List<dynamic>?)
      ?.map((e) => CustomerAddressModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  branchnumber: json['branchnumber'] as String?,
  code: json['code'] as String,
  email: json['email'] as String?,
  groups: (json['groups'] as List<dynamic>?)?.map((e) => e as String).toList(),
  images: (json['images'] as List<dynamic>?)
      ?.map((e) => ImagesModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  iscreditor: json['iscreditor'] as bool,
  isdebtor: json['isdebtor'] as bool,
  names: (json['names'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  personaltype: (json['personaltype'] as num).toInt(),
  customertype: (json['customertype'] as num).toInt(),
  taxid: json['taxid'] as String?,
  guidfixed: json['guidfixed'] as String,
  fundcode: json['fundcode'] as String?,
  creditday: (json['creditday'] as num?)?.toInt(),
);

Map<String, dynamic> _$CustomerRequestModelToJson(
  CustomerRequestModel instance,
) => <String, dynamic>{
  'addressforbilling': instance.addressforbilling.toJson(),
  'addressforshipping': instance.addressforshipping
      .map((e) => e.toJson())
      .toList(),
  'branchnumber': instance.branchnumber,
  'code': instance.code,
  'email': instance.email,
  'groups': instance.groups,
  'images': instance.images.map((e) => e.toJson()).toList(),
  'iscreditor': instance.iscreditor,
  'isdebtor': instance.isdebtor,
  'names': instance.names.map((e) => e.toJson()).toList(),
  'personaltype': instance.personaltype,
  'customertype': instance.customertype,
  'taxid': instance.taxid,
  'guidfixed': instance.guidfixed,
  'fundcode': instance.fundcode,
  'creditday': instance.creditday,
};
