import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'master_group_sub1_model.g.dart';

@JsonSerializable(explicitToJson: true)
class MasterGroupSub1Model {
  String guidfixed;
  String code;
  List<LanguageDataModel> names = <LanguageDataModel>[];
  String groupMainGuid;
  List<LanguageDataModel> groupMainNames = <LanguageDataModel>[];

  MasterGroupSub1Model({
    required this.guidfixed,
    required this.code,
    List<LanguageDataModel>? names,
    required this.groupMainGuid,
    List<LanguageDataModel>? groupMainNames,
  })  : names = names ?? <LanguageDataModel>[],
        groupMainNames = groupMainNames ?? <LanguageDataModel>[];

  factory MasterGroupSub1Model.fromJson(Map<String, dynamic> json) =>
      _$MasterGroupSub1ModelFromJson(json);
  Map<String, dynamic> toJson() => _$MasterGroupSub1ModelToJson(this);
}
