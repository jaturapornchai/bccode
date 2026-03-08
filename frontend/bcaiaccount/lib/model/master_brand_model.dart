import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'master_brand_model.g.dart';

@JsonSerializable(explicitToJson: true)
class MasterBrandModel {
  String guidfixed;
  String code;
  List<LanguageDataModel> names = <LanguageDataModel>[];

  MasterBrandModel({
    required this.guidfixed,
    required this.code,
    List<LanguageDataModel>? names,
  }) : names = names ?? <LanguageDataModel>[];

  factory MasterBrandModel.fromJson(Map<String, dynamic> json) =>
      _$MasterBrandModelFromJson(json);
  Map<String, dynamic> toJson() => _$MasterBrandModelToJson(this);
}
