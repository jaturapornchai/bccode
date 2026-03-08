// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'staff_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

StaffModel _$StaffModelFromJson(Map<String, dynamic> json) => StaffModel(
  guidfixed: json['guidfixed'] as String,
  code: json['code'] as String,
  email: json['email'] as String,
  names: (json['names'] as List<dynamic>)
      .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  cashier: json['cashier'] as bool,
  order: json['order'] as bool,
);

Map<String, dynamic> _$StaffModelToJson(StaffModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'code': instance.code,
      'email': instance.email,
      'names': instance.names,
      'cashier': instance.cashier,
      'order': instance.order,
    };
