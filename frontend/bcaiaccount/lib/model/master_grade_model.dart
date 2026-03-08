import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'master_grade_model.g.dart';

@JsonSerializable(explicitToJson: true)
class MasterGradeModel {
  String guidfixed;
  String code;
  List<LanguageDataModel> names = <LanguageDataModel>[];

  MasterGradeModel({
    required this.guidfixed,
    required this.code,
    List<LanguageDataModel>? names,
  }) : names = names ?? <LanguageDataModel>[];

  factory MasterGradeModel.fromJson(Map<String, dynamic> json) =>
      _$MasterGradeModelFromJson(json);
  Map<String, dynamic> toJson() => _$MasterGradeModelToJson(this);
}
