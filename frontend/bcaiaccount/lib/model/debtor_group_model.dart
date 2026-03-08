import 'package:json_annotation/json_annotation.dart';
import 'package:smlaicloud/model/global_model.dart';

part 'debtor_group_model.g.dart';

@JsonSerializable(explicitToJson: true)
class DebtorGroupModel {
  String guidfixed;

  /// รหัสกลุ่มลูกหนี้
  String groupcode;

  /// ชื่อกลุ่มลูกหนี้
  List<LanguageDataModel> names = <LanguageDataModel>[];

  DebtorGroupModel({
    required this.guidfixed,
    required this.groupcode,
    List<LanguageDataModel>? names,
  }) : names = names ?? <LanguageDataModel>[];

  factory DebtorGroupModel.fromJson(Map<String, dynamic> json) =>
      _$DebtorGroupModelFromJson(json);

  Map<String, dynamic> toJson() => _$DebtorGroupModelToJson(this);
}
