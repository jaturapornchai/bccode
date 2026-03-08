// ignore_for_file: non_constant_identifier_names

import 'package:json_annotation/json_annotation.dart';

part 'devices_model.g.dart';

@JsonSerializable()
class DevicesModel {
  late String guidfixed;
  late String code;
  late String name1;
  int type = 0;

  DevicesModel(
      {required this.guidfixed,
      required this.code,
      required this.name1,
      required this.type});

  factory DevicesModel.fromJson(Map<String, dynamic> json) =>
      _$DevicesModelFromJson(json);
  Map<String, dynamic> toJson() => _$DevicesModelToJson(this);
}
