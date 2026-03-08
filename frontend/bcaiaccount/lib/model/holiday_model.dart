// ignore_for_file: non_constant_identifier_names

import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'holiday_model.g.dart';

@JsonSerializable()
class HolidayModel {
  late String guidfixed;
  late String date;
  List<LanguageDataModel> desc = <LanguageDataModel>[];

  HolidayModel({
    required this.guidfixed,
    required this.date,
    required this.desc,
  });

  factory HolidayModel.fromJson(Map<String, dynamic> json) =>
      _$HolidayModelFromJson(json);
  Map<String, dynamic> toJson() => _$HolidayModelToJson(this);
}
