import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/shelf_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'warehouse_location_update_model.g.dart';

@JsonSerializable()
class WarehouseLocationUpdateModel {
  String warehousecode;
  String locationcode;
  List<LanguageDataModel> locationnames = <LanguageDataModel>[];
  List<ShelfModel> shelf = <ShelfModel>[];

  WarehouseLocationUpdateModel({
    required this.warehousecode,
    required this.locationcode,
    List<LanguageDataModel>? locationnames,
    List<ShelfModel>? shelf,
  })  : locationnames = locationnames ?? <LanguageDataModel>[],
        shelf = shelf ?? <ShelfModel>[];

  factory WarehouseLocationUpdateModel.fromJson(Map<String, dynamic> json) =>
      _$WarehouseLocationUpdateModelFromJson(json);

  Map<String, dynamic> toJson() => _$WarehouseLocationUpdateModelToJson(this);
}
