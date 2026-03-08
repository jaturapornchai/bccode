import 'package:json_annotation/json_annotation.dart';
// import 'package:smlaicloud/model/global_model.dart'; // Not used
import 'package:smlaicloud/model/shelf_product_model.dart';

part 'shelf_model.g.dart';

@JsonSerializable(explicitToJson: true)
class ShelfModel {
  String code;
  String name;
  List<ShelfProductItemModel>? productitems;

  ShelfModel({
    required this.code,
    required this.name,
    List<ShelfProductItemModel>? productitems,
  }) : productitems = productitems ?? [];
  factory ShelfModel.fromJson(Map<String, dynamic> json) =>
      _$ShelfModelFromJson(json);
  Map<String, dynamic> toJson() => _$ShelfModelToJson(this);
}
