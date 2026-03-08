import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'bank_model.g.dart';

@JsonSerializable(explicitToJson: true)
class BankModel {
  String code;
  String guidfixed;
  String logo;
  List<LanguageDataModel> names = <LanguageDataModel>[];

  BankModel({
    String? guidfixed,
    String? code,
    String? logo,
    List<LanguageDataModel>? names,
  })  : names = names ?? <LanguageDataModel>[],
        guidfixed = guidfixed ?? "",
        code = code ?? "",
        logo = logo ?? "";

  factory BankModel.fromJson(Map<String, dynamic> json) =>
      _$BankModelFromJson(json);

  Map<String, dynamic> toJson() => _$BankModelToJson(this);
}
