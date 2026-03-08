import 'package:json_annotation/json_annotation.dart';
import 'package:smlaicloud/model/global_model.dart';

part 'color_model.g.dart';

@JsonSerializable(explicitToJson: true)
class ColorModel {
  String code;
  String guidfixed;
  String colorselect;
  String colorselecthex;
  String colorsystem;
  String colorsystemhex;
  List<LanguageDataModel> names = <LanguageDataModel>[];

  ColorModel({
    required this.guidfixed,
    required this.code,
    required this.colorselect,
    required this.colorsystem,
    required this.colorselecthex,
    required this.colorsystemhex,
    List<LanguageDataModel>? names,
  }) : names = names ?? <LanguageDataModel>[];

  factory ColorModel.fromJson(Map<String, dynamic> json) =>
      _$ColorModelFromJson(json);

  Map<String, dynamic> toJson() => _$ColorModelToJson(this);
}
