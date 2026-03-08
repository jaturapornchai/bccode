import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/shelf_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'warehouse_location_model.g.dart';

@JsonSerializable()
class WarehouseLocationModel {
  String guidfixed;
  String warehousecode;
  List<LanguageDataModel> warehousenames = <LanguageDataModel>[];
  String locationcode;
  List<LanguageDataModel> locationnames = <LanguageDataModel>[];
  List<ShelfModel> shelf = <ShelfModel>[];

  WarehouseLocationModel({
    required this.guidfixed,
    required this.warehousecode,
    List<LanguageDataModel>? warehousenames,
    required this.locationcode,
    List<LanguageDataModel>? locationnames,
    List<ShelfModel>? shelf,
  })  : warehousenames = warehousenames ?? <LanguageDataModel>[],
        locationnames = locationnames ?? <LanguageDataModel>[],
        shelf = shelf ?? <ShelfModel>[];

  factory WarehouseLocationModel.fromJson(Map<String, dynamic> json) =>
      _$WarehouseLocationModelFromJson(json);

  Map<String, dynamic> toJson() => _$WarehouseLocationModelToJson(this);
}
