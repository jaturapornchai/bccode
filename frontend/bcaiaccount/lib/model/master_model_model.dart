import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'master_model_model.g.dart';

@JsonSerializable(explicitToJson: true)
class MasterModelModel {
  String guidfixed;
  String code;
  List<LanguageDataModel> names = <LanguageDataModel>[];

  MasterModelModel({
    required this.guidfixed,
    required this.code,
    List<LanguageDataModel>? names,
  }) : names = names ?? <LanguageDataModel>[];

  factory MasterModelModel.fromJson(Map<String, dynamic> json) =>
      _$MasterModelModelFromJson(json);
  Map<String, dynamic> toJson() => _$MasterModelModelToJson(this);
}
