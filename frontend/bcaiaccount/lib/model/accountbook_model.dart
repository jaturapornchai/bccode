import 'package:json_annotation/json_annotation.dart';

part 'accountbook_model.g.dart';

@JsonSerializable(explicitToJson: true)
class AccountBookModel {
  String? guidfixed;
  String? code;
  String? name1;
  String? name2;
  String? name3;
  String? name4;
  String? name5;
  bool? iscenterbook;

  AccountBookModel({
    String? guidfixed,
    String? code,
    String? name1,
    String? name2,
    String? name3,
    String? name4,
    String? name5,
    bool? iscenterbook,
  })  : guidfixed = guidfixed ?? '',
        code = code ?? '',
        name1 = name1 ?? '',
        name2 = name2 ?? '',
        name3 = name3 ?? '',
        name4 = name4 ?? '',
        name5 = name5 ?? '',
        iscenterbook = iscenterbook ?? false;

  factory AccountBookModel.fromJson(Map<String, dynamic> json) =>
      _$AccountBookModelFromJson(json);

  Map<String, dynamic> toJson() => _$AccountBookModelToJson(this);
}
