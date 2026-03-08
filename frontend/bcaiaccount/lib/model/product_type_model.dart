import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'product_type_model.g.dart';

@JsonSerializable()
class ProductTypeModel {
  String? guidfixed;
  String? code;
  List<LanguageDataModel>? names;

  ProductTypeModel({
    String? guidfixed,
    String? code,
    List<LanguageDataModel>? names,
  })  : guidfixed = guidfixed ?? "",
        code = code ?? "",
        names = names ?? <LanguageDataModel>[];

  factory ProductTypeModel.fromJson(Map<String, dynamic> json) =>
      _$ProductTypeModelFromJson(json);
  Map<String, dynamic> toJson() => _$ProductTypeModelToJson(this);
}
