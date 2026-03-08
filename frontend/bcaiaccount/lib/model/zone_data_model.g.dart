// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'zone_data_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ZoneDataModel _$ZoneDataModelFromJson(Map<String, dynamic> json) =>
    ZoneDataModel(
      guidfixed: json['guidfixed'] as String,
      code: json['code'] as String,
      names: (json['names'] as List<dynamic>)
          .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      printers: json['printers'] == null
          ? null
          : PrinterModel.fromJson(json['printers'] as Map<String, dynamic>),
      groupnumber: (json['groupnumber'] as num?)?.toInt(),
    );

Map<String, dynamic> _$ZoneDataModelToJson(ZoneDataModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'code': instance.code,
      'names': instance.names,
      'printers': instance.printers,
      'groupnumber': instance.groupnumber,
    };
