import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/kitchen_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'kitchen_product_model.g.dart';

@JsonSerializable()
class KitchenProductModel {
  String guidfixed;
  String barcode;
  List<LanguageDataModel> names;
  bool? isdisable;
  List<KitchenModel>? kitchens;

  KitchenProductModel({
    required this.guidfixed,
    required this.barcode,
    required this.names,
    bool? isdisable,
    String? kitchenname,
    List<KitchenModel>? kitchens,
  })  : isdisable = isdisable ?? false,
        kitchens = kitchens ?? [];

  factory KitchenProductModel.fromJson(Map<String, dynamic> json) =>
      _$KitchenProductModelFromJson(json);
  Map<String, dynamic> toJson() => _$KitchenProductModelToJson(this);
}

@JsonSerializable()
class ProductInKitchenModel {
  String? barcode;
  List<KitchenModel>? kitchens;

  ProductInKitchenModel({
    String? barcode,
    List<KitchenModel>? kitchens,
  })  : barcode = barcode ?? "",
        kitchens = kitchens ?? [];

  factory ProductInKitchenModel.fromJson(Map<String, dynamic> json) =>
      _$ProductInKitchenModelFromJson(json);
  Map<String, dynamic> toJson() => _$ProductInKitchenModelToJson(this);
}
