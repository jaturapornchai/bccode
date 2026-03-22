import 'package:json_annotation/json_annotation.dart';
import 'package:smlaicloud/model/global_model.dart';

part 'job_project_model.g.dart';

@JsonSerializable(explicitToJson: true)
class JobProjectModel {
  String guidfixed;
  String code;
  String parentcode;
  List<LanguageDataModel> names = <LanguageDataModel>[];

  JobProjectModel({
    String? guidfixed,
    required this.code,
    String? parentcode,
    List<LanguageDataModel>? names,
  })  : guidfixed = guidfixed ?? "",
        parentcode = parentcode ?? "",
        names = names ?? <LanguageDataModel>[];

  factory JobProjectModel.fromJson(Map<String, dynamic> json) =>
      _$JobProjectModelFromJson(json);

  Map<String, dynamic> toJson() => _$JobProjectModelToJson(this);
}
