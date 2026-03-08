// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'member_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

MemberModel _$MemberModelFromJson(Map<String, dynamic> json) => MemberModel(
  guidfixed: json['guidfixed'] as String?,
  address: json['address'] as String?,
  branchcode: json['branchcode'] as String?,
  branchtype: (json['branchtype'] as num).toInt(),
  contacttype: (json['contacttype'] as num).toInt(),
  name: json['name'] as String,
  personaltype: (json['personaltype'] as num).toInt(),
  surname: json['surname'] as String?,
  taxid: json['taxid'] as String?,
  telephone: json['telephone'] as String?,
  zipcode: json['zipcode'] as String?,
);

Map<String, dynamic> _$MemberModelToJson(MemberModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'address': instance.address,
      'branchcode': instance.branchcode,
      'branchtype': instance.branchtype,
      'contacttype': instance.contacttype,
      'name': instance.name,
      'personaltype': instance.personaltype,
      'surname': instance.surname,
      'taxid': instance.taxid,
      'telephone': instance.telephone,
      'zipcode': instance.zipcode,
    };
