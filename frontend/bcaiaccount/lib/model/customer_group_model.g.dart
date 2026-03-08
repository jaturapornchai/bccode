// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'customer_group_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

CustomerGroupModel _$CustomerGroupModelFromJson(Map<String, dynamic> json) =>
    CustomerGroupModel(
      guidfixed: json['guidfixed'] as String,
      groupcode: json['groupcode'] as String,
      names: (json['names'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$CustomerGroupModelToJson(CustomerGroupModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'groupcode': instance.groupcode,
      'names': instance.names.map((e) => e.toJson()).toList(),
    };
