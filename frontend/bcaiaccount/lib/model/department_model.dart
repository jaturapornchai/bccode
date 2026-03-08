import 'package:json_annotation/json_annotation.dart';
import 'package:smlaicloud/model/global_model.dart';

part 'department_model.g.dart';

@JsonSerializable(explicitToJson: true)
class DepartmentModel {
  String guidfixed;
  String code;
  List<LanguageDataModel> names = <LanguageDataModel>[];

  DepartmentModel({
    String? guidfixed,
    required this.code,
    List<LanguageDataModel>? names,
  })  : guidfixed = guidfixed ?? "",
        names = names ?? <LanguageDataModel>[];

  factory DepartmentModel.fromJson(Map<String, dynamic> json) =>
      _$DepartmentModelFromJson(json);

  Map<String, dynamic> toJson() => _$DepartmentModelToJson(this);
}
