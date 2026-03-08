import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'master_category_model.g.dart';

@JsonSerializable(explicitToJson: true)
class MasterCategoryModel {
  String guidfixed;
  String code;
  List<LanguageDataModel> names = <LanguageDataModel>[];

  MasterCategoryModel({
    required this.guidfixed,
    required this.code,
    List<LanguageDataModel>? names,
  }) : names = names ?? <LanguageDataModel>[];

  factory MasterCategoryModel.fromJson(Map<String, dynamic> json) =>
      _$MasterCategoryModelFromJson(json);
  Map<String, dynamic> toJson() => _$MasterCategoryModelToJson(this);
}
