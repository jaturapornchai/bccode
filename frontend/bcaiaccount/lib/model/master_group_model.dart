import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'master_group_model.g.dart';

@JsonSerializable(explicitToJson: true)
class MasterGroupModel {
  String guidfixed;
  String code;
  List<LanguageDataModel> names = <LanguageDataModel>[];

  MasterGroupModel({
    required this.guidfixed,
    required this.code,
    List<LanguageDataModel>? names,
  }) : names = names ?? <LanguageDataModel>[];

  factory MasterGroupModel.fromJson(Map<String, dynamic> json) =>
      _$MasterGroupModelFromJson(json);
  Map<String, dynamic> toJson() => _$MasterGroupModelToJson(this);
}
