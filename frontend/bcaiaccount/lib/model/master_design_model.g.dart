// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'master_design_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

MasterDesignModel _$MasterDesignModelFromJson(Map<String, dynamic> json) =>
    MasterDesignModel(
      guidfixed: json['guidfixed'] as String,
      code: json['code'] as String,
      names: (json['names'] as List<dynamic>?)
          ?.map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$MasterDesignModelToJson(MasterDesignModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'code': instance.code,
      'names': instance.names.map((e) => e.toJson()).toList(),
    };
