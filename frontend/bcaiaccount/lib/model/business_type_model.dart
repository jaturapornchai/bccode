import 'package:json_annotation/json_annotation.dart';
import 'package:smlaicloud/model/global_model.dart';

part 'business_type_model.g.dart';

@JsonSerializable(explicitToJson: true)
class BusinessTypeModel {
  String? guidfixed;
  String? code;
  List<LanguageDataModel>? names;
  bool? isdefault;

  BusinessTypeModel({
    String? guidfixed,
    String? code,
    List<LanguageDataModel>? names,
    bool? isdefault,
  })  : names = names ?? <LanguageDataModel>[],
        guidfixed = guidfixed ?? "",
        code = code ?? "",
        isdefault = isdefault ?? false;

  factory BusinessTypeModel.fromJson(Map<String, dynamic> json) =>
      _$BusinessTypeModelFromJson(json);

  Map<String, dynamic> toJson() => _$BusinessTypeModelToJson(this);
}
