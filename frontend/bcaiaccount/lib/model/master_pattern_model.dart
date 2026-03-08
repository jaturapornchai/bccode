import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'master_pattern_model.g.dart';

@JsonSerializable(explicitToJson: true)
class MasterPatternModel {
  String guidfixed;
  String code;
  List<LanguageDataModel> names = <LanguageDataModel>[];

  MasterPatternModel({
    required this.guidfixed,
    required this.code,
    List<LanguageDataModel>? names,
  }) : names = names ?? <LanguageDataModel>[];

  factory MasterPatternModel.fromJson(Map<String, dynamic> json) =>
      _$MasterPatternModelFromJson(json);
  Map<String, dynamic> toJson() => _$MasterPatternModelToJson(this);
}
