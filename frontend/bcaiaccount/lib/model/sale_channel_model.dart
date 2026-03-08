import 'package:json_annotation/json_annotation.dart';

part 'sale_channel_model.g.dart';

@JsonSerializable()
class SaleChannelModel {
  String? guidfixed;
  String? code;
  double? gp;
  int? gptype;
  String? imageuri;
  String? name;
  int? price;

  SaleChannelModel({
    String? guidfixed,
    String? code,
    double? gp,
    int? gptype,
    String? imageuri,
    String? name,
    int? price,
  })  : guidfixed = guidfixed ?? '',
        code = code ?? '',
        gp = gp ?? 0.0,
        gptype = gptype ?? 0,
        imageuri = imageuri ?? '',
        name = name ?? '',
        price = price ?? 0;

  factory SaleChannelModel.fromJson(Map<String, dynamic> json) =>
      _$SaleChannelModelFromJson(json);

  Map<String, dynamic> toJson() => _$SaleChannelModelToJson(this);
}
