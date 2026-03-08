// ignore_for_file: non_constant_identifier_names

import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'printer_model.g.dart';

@JsonSerializable()
class PrinterModel {
  late String guidfixed;
  late String code;
  late List<LanguageDataModel> names;
  late int type;
  late String address;

  PrinterModel(
      {required this.guidfixed,
      required this.code,
      required this.names,
      required this.address,
      required this.type});

  factory PrinterModel.fromJson(Map<String, dynamic> json) =>
      _$PrinterModelFromJson(json);
  Map<String, dynamic> toJson() => _$PrinterModelToJson(this);
}
