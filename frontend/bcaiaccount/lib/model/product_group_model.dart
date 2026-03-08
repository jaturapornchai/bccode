import 'package:json_annotation/json_annotation.dart';
import 'package:smlaicloud/model/global_model.dart';

part 'product_group_model.g.dart';

@JsonSerializable(explicitToJson: true)
class ProductGroupModel {
  String guidfixed;
  String code; //อ้างอิง
  List<LanguageDataModel> names; // ชื่อ (หลายภาษา)

  ProductGroupModel({
    required this.guidfixed,
    required this.code,
    required this.names,
  });

  factory ProductGroupModel.fromJson(Map<String, dynamic> json) =>
      _$ProductGroupModelFromJson(json);

  Map<String, dynamic> toJson() => _$ProductGroupModelToJson(this);
}
