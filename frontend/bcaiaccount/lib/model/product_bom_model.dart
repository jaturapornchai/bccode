import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'product_bom_model.g.dart'; // This is generated automatically by json_serializable

@JsonSerializable()
class ProductBomModel {
  String? guidfixed;
  List<LanguageDataModel>? names;
  String? itemunitcode;
  List<LanguageDataModel>? itemunitnames;
  String? barcode;
  bool? condition;
  int? dividevalue;
  int? standvalue;
  double? qty;
  String? imageuri;
  List<ProductBomModel>? bom;

  ProductBomModel({
    String? guidfixed,
    List<LanguageDataModel>? names,
    String? itemunitcode,
    List<LanguageDataModel>? itemunitnames,
    String? barcode,
    bool? condition,
    int? dividevalue,
    int? standvalue,
    double? qty,
    String? imageuri,
    List<ProductBomModel>? bom,
  })  : guidfixed = guidfixed ?? '',
        names = names ?? [],
        itemunitcode = itemunitcode ?? '',
        itemunitnames = itemunitnames ?? [],
        barcode = barcode ?? '',
        condition = condition ?? false,
        dividevalue = dividevalue ?? 0,
        standvalue = standvalue ?? 0,
        qty = qty ?? 0.0,
        imageuri = imageuri ?? '',
        bom = bom ?? [];

  factory ProductBomModel.fromJson(Map<String, dynamic> json) =>
      _$ProductBomModelFromJson(json);
  Map<String, dynamic> toJson() => _$ProductBomModelToJson(this);
}
