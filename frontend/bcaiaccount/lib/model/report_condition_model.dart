import 'package:json_annotation/json_annotation.dart';

part 'report_condition_model.g.dart';

@JsonSerializable(explicitToJson: true)
class ConditionLocationSelectModel {
  String code;
  String title;
  bool isSelected = true;

  ConditionLocationSelectModel(this.code, this.title);

  factory ConditionLocationSelectModel.fromJson(Map<String, dynamic> json) =>
      _$ConditionLocationSelectModelFromJson(json);
  Map<String, dynamic> toJson() => _$ConditionLocationSelectModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class ConditionWareHouseSelectModel {
  String code;
  String title;
  bool isSelected = true;
  List<ConditionLocationSelectModel> locations = [];

  ConditionWareHouseSelectModel(this.code, this.title);
  factory ConditionWareHouseSelectModel.fromJson(Map<String, dynamic> json) =>
      _$ConditionWareHouseSelectModelFromJson(json);
  Map<String, dynamic> toJson() => _$ConditionWareHouseSelectModelToJson(this);
}
