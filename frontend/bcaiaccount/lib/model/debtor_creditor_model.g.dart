// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'debtor_creditor_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

DebtorCreditorModel _$DebtorCreditorModelFromJson(Map<String, dynamic> json) =>
    DebtorCreditorModel(
      addressforbilling: CustomerAddressModel.fromJson(
        json['addressforbilling'] as Map<String, dynamic>,
      ),
      addressforshipping: (json['addressforshipping'] as List<dynamic>?)
          ?.map((e) => CustomerAddressModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      branchnumber: json['branchnumber'] as String?,
      code: json['code'] as String?,
      email: json['email'] as String?,
      groups: (json['groups'] as List<dynamic>?)
          ?.map((e) => DebtorGroupModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      images: (json['images'] as List<dynamic>?)
          ?.map((e) => ImagesModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      names: (json['names'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      personaltype: (json['personaltype'] as num?)?.toInt(),
      customertype: (json['customertype'] as num?)?.toInt(),
      taxid: json['taxid'] as String?,
      fundcode: json['fundcode'] as String?,
      creditday: (json['creditday'] as num?)?.toInt(),
      guidfixed: json['guidfixed'] as String?,
      ismember: json['ismember'] as bool?,
    );

Map<String, dynamic> _$DebtorCreditorModelToJson(
  DebtorCreditorModel instance,
) => <String, dynamic>{
  'addressforbilling': instance.addressforbilling.toJson(),
  'addressforshipping': instance.addressforshipping
      ?.map((e) => e.toJson())
      .toList(),
  'branchnumber': instance.branchnumber,
  'code': instance.code,
  'email': instance.email,
  'groups': instance.groups?.map((e) => e.toJson()).toList(),
  'images': instance.images?.map((e) => e.toJson()).toList(),
  'names': instance.names?.map((e) => e.toJson()).toList(),
  'personaltype': instance.personaltype,
  'customertype': instance.customertype,
  'taxid': instance.taxid,
  'fundcode': instance.fundcode,
  'creditday': instance.creditday,
  'guidfixed': instance.guidfixed,
  'ismember': instance.ismember,
};
