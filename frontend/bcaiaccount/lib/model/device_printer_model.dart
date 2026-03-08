// ignore_for_file: non_constant_identifier_names

import 'package:json_annotation/json_annotation.dart';

part 'device_printer_model.g.dart';

@JsonSerializable()
class DevicePrinterModel {
  late String guidfixed;
  late String guiddevice;
  late String devicecode;
  late String devicename;
  late String primary;
  late String spare;

  DevicePrinterModel({
    required this.guidfixed,
    required this.guiddevice,
    required this.devicecode,
    required this.devicename,
    required this.primary,
    required this.spare,
  });

  factory DevicePrinterModel.fromJson(Map<String, dynamic> json) =>
      _$DevicePrinterModelFromJson(json);
  Map<String, dynamic> toJson() => _$DevicePrinterModelToJson(this);
}

@JsonSerializable()
class DevicePrinterSaveModel {
  late String guidfixed;
  late String guiddevice;
  late String devicecode;
  late String devicename;
  late List<String> printers;

  DevicePrinterSaveModel({
    required this.guidfixed,
    required this.guiddevice,
    required this.devicecode,
    required this.devicename,
    required this.printers,
  });

  factory DevicePrinterSaveModel.fromJson(Map<String, dynamic> json) =>
      _$DevicePrinterSaveModelFromJson(json);
  Map<String, dynamic> toJson() => _$DevicePrinterSaveModelToJson(this);
}
