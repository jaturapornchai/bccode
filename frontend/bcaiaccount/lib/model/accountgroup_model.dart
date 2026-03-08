import 'package:json_annotation/json_annotation.dart';

part 'accountgroup_model.g.dart';

@JsonSerializable(explicitToJson: true)
class AccountGroupModel {
  String? guidfixed;
  String? code;
  String? name1;
  String? name2;
  String? name3;
  String? name4;
  String? name5;

  AccountGroupModel({
    String? guidfixed,
    String? code,
    String? name1,
    String? name2,
    String? name3,
    String? name4,
    String? name5,
  })  : guidfixed = guidfixed ?? '',
        code = code ?? '',
        name1 = name1 ?? '',
        name2 = name2 ?? '',
        name3 = name3 ?? '',
        name4 = name4 ?? '',
        name5 = name5 ?? '';

  factory AccountGroupModel.fromJson(Map<String, dynamic> json) =>
      _$AccountGroupModelFromJson(json);

  Map<String, dynamic> toJson() => _$AccountGroupModelToJson(this);
}
