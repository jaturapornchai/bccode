// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'customer_address_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

CustomerAddressModel _$CustomerAddressModelFromJson(
  Map<String, dynamic> json,
) => CustomerAddressModel(
  guid: json['guid'] as String?,
  address: (json['address'] as List<dynamic>?)
      ?.map((e) => e as String)
      .toList(),
  countrycode: json['countrycode'] as String?,
  provincecode: json['provincecode'] as String?,
  districtcode: json['districtcode'] as String?,
  subdistrictcode: json['subdistrictcode'] as String?,
  zipcode: json['zipcode'] as String?,
  contactnames: (json['contactnames'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  phoneprimary: json['phoneprimary'] as String?,
  phonesecondary: json['phonesecondary'] as String?,
  latitude: (json['latitude'] as num?)?.toDouble(),
  longitude: (json['longitude'] as num?)?.toDouble(),
);

Map<String, dynamic> _$CustomerAddressModelToJson(
  CustomerAddressModel instance,
) => <String, dynamic>{
  'guid': instance.guid,
  'address': instance.address,
  'countrycode': instance.countrycode,
  'provincecode': instance.provincecode,
  'districtcode': instance.districtcode,
  'subdistrictcode': instance.subdistrictcode,
  'zipcode': instance.zipcode,
  'contactnames': instance.contactnames?.map((e) => e.toJson()).toList(),
  'phoneprimary': instance.phoneprimary,
  'phonesecondary': instance.phonesecondary,
  'latitude': instance.latitude,
  'longitude': instance.longitude,
};
