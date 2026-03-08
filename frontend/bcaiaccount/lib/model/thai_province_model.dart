// ignore_for_file: non_constant_identifier_names

import 'package:json_annotation/json_annotation.dart';

part 'thai_province_model.g.dart';

@JsonSerializable()
class ThaiProvinceModel {
  late int id;
  late String name_th;
  late String name_en;
  late int geography_id;

  ThaiProvinceModel({
    required this.id,
    required this.name_th,
    required this.name_en,
    required this.geography_id,
  });

  factory ThaiProvinceModel.fromJson(Map<String, dynamic> json) =>
      _$ThaiProvinceModelFromJson(json);
  Map<String, dynamic> toJson() => _$ThaiProvinceModelToJson(this);
}
