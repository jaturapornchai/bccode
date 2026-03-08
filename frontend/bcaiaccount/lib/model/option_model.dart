import 'package:smlaicloud/model/choice_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'option_model.g.dart';

@JsonSerializable()
class OptionModel {
  final String guidfixed;
  int choicetype = 0;
  final String code;
  int maxselect = 0;
  final String name1;
  final String name2;
  final String name3;
  final String name4;
  final String name5;
  @JsonKey(name: 'required')
  bool isrequired;
  List<ChoiceModel> choices = <ChoiceModel>[];

  OptionModel({
    String? guidfixed,
    required this.choicetype,
    required this.code,
    required this.maxselect,
    required this.name1,
    String? name2,
    String? name3,
    String? name4,
    String? name5,
    required this.isrequired,
    List<ChoiceModel>? choices,
  })  : guidfixed = guidfixed ?? '',
        name2 = name2 ?? '',
        name3 = name3 ?? '',
        name4 = name4 ?? '',
        name5 = name5 ?? '',
        choices = choices ?? <ChoiceModel>[];

  factory OptionModel.fromJson(Map<String, dynamic> json) =>
      _$OptionModelFromJson(json);

  Map<String, dynamic> toJson() => _$OptionModelToJson(this);
}
