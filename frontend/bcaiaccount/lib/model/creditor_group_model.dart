import 'package:json_annotation/json_annotation.dart';
import 'package:smlaicloud/model/global_model.dart';

part 'creditor_group_model.g.dart';

@JsonSerializable(explicitToJson: true)
class CreditorGroupModel {
  String guidfixed;

  /// รหัสกลุ่มเจ้าหนี้
  String groupcode;

  /// ชื่อกลุ่มเจ้าหนี้
  List<LanguageDataModel> names = <LanguageDataModel>[];

  CreditorGroupModel({
    required this.guidfixed,
    required this.groupcode,
    List<LanguageDataModel>? names,
  }) : names = names ?? <LanguageDataModel>[];

  factory CreditorGroupModel.fromJson(Map<String, dynamic> json) =>
      _$CreditorGroupModelFromJson(json);

  Map<String, dynamic> toJson() => _$CreditorGroupModelToJson(this);
}
