// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'contact_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ContactModel _$ContactModelFromJson(Map<String, dynamic> json) => ContactModel(
  address: (json['address'] as List<dynamic>?)
      ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  countrycode: json['countrycode'] as String?,
  districtcode: json['districtcode'] as String?,
  latitude: (json['latitude'] as num?)?.toDouble(),
  longitude: (json['longitude'] as num?)?.toDouble(),
  phonenumber: json['phonenumber'] as String?,
  provincecode: json['provincecode'] as String?,
  subdistrictcode: json['subdistrictcode'] as String?,
  zipcode: json['zipcode'] as String?,
);

Map<String, dynamic> _$ContactModelToJson(ContactModel instance) =>
    <String, dynamic>{
      'address': instance.address,
      'countrycode': instance.countrycode,
      'districtcode': instance.districtcode,
      'latitude': instance.latitude,
      'longitude': instance.longitude,
      'phonenumber': instance.phonenumber,
      'provincecode': instance.provincecode,
      'subdistrictcode': instance.subdistrictcode,
      'zipcode': instance.zipcode,
    };

ContactEmployeeModel _$ContactEmployeeModelFromJson(
  Map<String, dynamic> json,
) => ContactEmployeeModel(
  address: json['address'] as String?,
  countrycode: json['countrycode'] as String?,
  districtcode: json['districtcode'] as String?,
  latitude: (json['latitude'] as num?)?.toDouble(),
  longitude: (json['longitude'] as num?)?.toDouble(),
  phonenumber: json['phonenumber'] as String?,
  provincecode: json['provincecode'] as String?,
  subdistrictcode: json['subdistrictcode'] as String?,
  zipcode: json['zipcode'] as String?,
);

Map<String, dynamic> _$ContactEmployeeModelToJson(
  ContactEmployeeModel instance,
) => <String, dynamic>{
  'address': instance.address,
  'countrycode': instance.countrycode,
  'districtcode': instance.districtcode,
  'latitude': instance.latitude,
  'longitude': instance.longitude,
  'phonenumber': instance.phonenumber,
  'provincecode': instance.provincecode,
  'subdistrictcode': instance.subdistrictcode,
  'zipcode': instance.zipcode,
};
