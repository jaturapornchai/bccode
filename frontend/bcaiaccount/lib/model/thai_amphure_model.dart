// ignore_for_file: non_constant_identifier_names

import 'package:json_annotation/json_annotation.dart';

part 'thai_amphure_model.g.dart';

@JsonSerializable()
class ThaiAmphureModel {
  late int id;
  late String name_th;
  late String name_en;
  late int province_id;

  ThaiAmphureModel({
    required this.id,
    required this.name_th,
    required this.name_en,
    required this.province_id,
  });

  factory ThaiAmphureModel.fromJson(Map<String, dynamic> json) =>
      _$ThaiAmphureModelFromJson(json);
  Map<String, dynamic> toJson() => _$ThaiAmphureModelToJson(this);
}
