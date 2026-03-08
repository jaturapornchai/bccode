// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'device_printer_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

DevicePrinterModel _$DevicePrinterModelFromJson(Map<String, dynamic> json) =>
    DevicePrinterModel(
      guidfixed: json['guidfixed'] as String,
      guiddevice: json['guiddevice'] as String,
      devicecode: json['devicecode'] as String,
      devicename: json['devicename'] as String,
      primary: json['primary'] as String,
      spare: json['spare'] as String,
    );

Map<String, dynamic> _$DevicePrinterModelToJson(DevicePrinterModel instance) =>
    <String, dynamic>{
      'guidfixed': instance.guidfixed,
      'guiddevice': instance.guiddevice,
      'devicecode': instance.devicecode,
      'devicename': instance.devicename,
      'primary': instance.primary,
      'spare': instance.spare,
    };

DevicePrinterSaveModel _$DevicePrinterSaveModelFromJson(
  Map<String, dynamic> json,
) => DevicePrinterSaveModel(
  guidfixed: json['guidfixed'] as String,
  guiddevice: json['guiddevice'] as String,
  devicecode: json['devicecode'] as String,
  devicename: json['devicename'] as String,
  printers: (json['printers'] as List<dynamic>)
      .map((e) => e as String)
      .toList(),
);

Map<String, dynamic> _$DevicePrinterSaveModelToJson(
  DevicePrinterSaveModel instance,
) => <String, dynamic>{
  'guidfixed': instance.guidfixed,
  'guiddevice': instance.guiddevice,
  'devicecode': instance.devicecode,
  'devicename': instance.devicename,
  'printers': instance.printers,
};
