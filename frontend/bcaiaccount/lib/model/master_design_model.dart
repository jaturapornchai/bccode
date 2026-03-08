import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'master_design_model.g.dart';

@JsonSerializable(explicitToJson: true)
class MasterDesignModel {
  String guidfixed;
  String code;
  List<LanguageDataModel> names = <LanguageDataModel>[];

  MasterDesignModel({
    required this.guidfixed,
    required this.code,
    List<LanguageDataModel>? names,
  }) : names = names ?? <LanguageDataModel>[];

  factory MasterDesignModel.fromJson(Map<String, dynamic> json) =>
      _$MasterDesignModelFromJson(json);
  Map<String, dynamic> toJson() => _$MasterDesignModelToJson(this);
}
