import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'master_class_model.g.dart';

@JsonSerializable(explicitToJson: true)
class MasterClassModel {
  String guidfixed;
  String code;
  List<LanguageDataModel> names = <LanguageDataModel>[];

  MasterClassModel({
    required this.guidfixed,
    required this.code,
    List<LanguageDataModel>? names,
  }) : names = names ?? <LanguageDataModel>[];

  factory MasterClassModel.fromJson(Map<String, dynamic> json) =>
      _$MasterClassModelFromJson(json);
  Map<String, dynamic> toJson() => _$MasterClassModelToJson(this);
}
