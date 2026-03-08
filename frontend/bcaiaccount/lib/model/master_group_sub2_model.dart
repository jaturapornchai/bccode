import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'master_group_sub2_model.g.dart';

@JsonSerializable(explicitToJson: true)
class MasterGroupSub2Model {
  String guidfixed;
  String code;
  List<LanguageDataModel> names = <LanguageDataModel>[];
  String groupMainGuid;
  List<LanguageDataModel> groupMainNames = <LanguageDataModel>[];
  String groupSubGuid;
  List<LanguageDataModel> groupSubNames = <LanguageDataModel>[];

  MasterGroupSub2Model({
    required this.guidfixed,
    required this.code,
    List<LanguageDataModel>? names,
    required this.groupMainGuid,
    List<LanguageDataModel>? groupMainNames,
    required this.groupSubGuid,
    List<LanguageDataModel>? groupSubNames,
  })  : names = names ?? <LanguageDataModel>[],
        groupMainNames = groupMainNames ?? <LanguageDataModel>[],
        groupSubNames = groupSubNames ?? <LanguageDataModel>[];

  factory MasterGroupSub2Model.fromJson(Map<String, dynamic> json) =>
      _$MasterGroupSub2ModelFromJson(json);
  Map<String, dynamic> toJson() => _$MasterGroupSub2ModelToJson(this);
}
