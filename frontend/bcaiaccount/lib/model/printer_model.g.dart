// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'printer_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PrinterModel _$PrinterModelFromJson(Map<String, dynamic> json) => PrinterModel(
  guidfixed: json['guidfixed'] as String,
  code: json['code'] as String,
  names: (json['names'] as List<dynamic>)
      .map((e) => LanguageDataModel.fromJson(e as Map<String, dynamic>))
      .toList(),
  address: json['address'] as String,
  type: (json['type'] as num).toInt(),
);

Map<String, dynamic> _$PrinterModelToJson(PrinterModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'code': instance.code,
      'names': instance.names,
      'type': instance.type,
      'address': instance.address,
    };
