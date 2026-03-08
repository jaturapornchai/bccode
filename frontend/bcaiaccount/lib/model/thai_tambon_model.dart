// ignore_for_file: non_constant_identifier_names

import 'package:json_annotation/json_annotation.dart';

part 'thai_tambon_model.g.dart';

@JsonSerializable()
class ThaiTambonModel {
  late int id;
  late String name_th;
  late String name_en;
  late int amphure_id;
  late int zip_code;

  ThaiTambonModel({
    required this.id,
    required this.name_th,
    required this.name_en,
    required this.amphure_id,
    required this.zip_code,
  });

  factory ThaiTambonModel.fromJson(Map<String, dynamic> json) =>
      _$ThaiTambonModelFromJson(json);
  Map<String, dynamic> toJson() => _$ThaiTambonModelToJson(this);
}
