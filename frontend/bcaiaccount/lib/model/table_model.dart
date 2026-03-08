// ignore_for_file: non_constant_identifier_names

import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'table_model.g.dart';

@JsonSerializable()
class TableModel {
  late String guidfixed;
  late String number;
  late List<LanguageDataModel> names;
  late String zone;
  late int? xorder;
  String? orderendcode;
  int? groupnumber;
  int? zonenumber;
  TableModel({
    required this.guidfixed,
    required this.number,
    required this.names,
    required this.zone,
    int? xorder,
    String? orderendcode,
    int? groupnumber,
    int? zonenumber,
  })  : xorder = xorder ?? 0,
        orderendcode = orderendcode ?? "",
        groupnumber = groupnumber ?? 1,
        zonenumber = zonenumber ?? 1;

  factory TableModel.fromJson(Map<String, dynamic> json) =>
      _$TableModelFromJson(json);
  Map<String, dynamic> toJson() => _$TableModelToJson(this);
}

@JsonSerializable()
class TableXorderModel {
  late String guidfixed;
  late int xorder;
  TableXorderModel({
    required this.guidfixed,
    required this.xorder,
  });

  factory TableXorderModel.fromJson(Map<String, dynamic> json) =>
      _$TableXorderModelFromJson(json);
  Map<String, dynamic> toJson() => _$TableXorderModelToJson(this);
}
