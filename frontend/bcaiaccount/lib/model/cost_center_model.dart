import 'package:json_annotation/json_annotation.dart';
import 'package:smlaicloud/model/global_model.dart';

part 'cost_center_model.g.dart';

@JsonSerializable(explicitToJson: true)
class CostCenterModel {
  String guidfixed;
  String code;
  List<LanguageDataModel> names = <LanguageDataModel>[];

  CostCenterModel({
    String? guidfixed,
    required this.code,
    List<LanguageDataModel>? names,
  })  : guidfixed = guidfixed ?? "",
        names = names ?? <LanguageDataModel>[];

  factory CostCenterModel.fromJson(Map<String, dynamic> json) =>
      _$CostCenterModelFromJson(json);

  Map<String, dynamic> toJson() => _$CostCenterModelToJson(this);
}
