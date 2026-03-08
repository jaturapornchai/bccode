// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'employee_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

EmployeeModel _$EmployeeModelFromJson(Map<String, dynamic> json) =>
    EmployeeModel(
      guidfixed: json['guidfixed'] as String,
      code: json['code'] as String,
      name: json['name'] as String?,
      email: json['email'] as String?,
      profilepicture: json['profilepicture'] as String?,
      isenabled: json['isenabled'] as bool?,
      contact: json['contact'] == null
          ? null
          : ContactEmployeeModel.fromJson(
              json['contact'] as Map<String, dynamic>,
            ),
      pincode: json['pincode'] as String?,
      isusepos: json['isusepos'] as bool?,
      branches: (json['branches'] as List<dynamic>?)
          ?.map((e) => CompanyBranchModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$EmployeeModelToJson(EmployeeModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'code': instance.code,
      'name': instance.name,
      'email': instance.email,
      'profilepicture': instance.profilepicture,
      'isenabled': instance.isenabled,
      'contact': instance.contact,
      'pincode': instance.pincode,
      'isusepos': instance.isusepos,
      'branches': instance.branches,
    };
