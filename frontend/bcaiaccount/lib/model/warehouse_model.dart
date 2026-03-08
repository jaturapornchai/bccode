import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/location_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'warehouse_model.g.dart';

@JsonSerializable(explicitToJson: true)
class WarehouseModel {
  String guidfixed;
  String code;
  List<LanguageDataModel> names = <LanguageDataModel>[];
  List<LocationModel> location = <LocationModel>[];

  WarehouseModel({
    required this.guidfixed,
    required this.code,
    List<LanguageDataModel>? names,
    List<LocationModel>? location,
  })  : names = names ?? <LanguageDataModel>[],
        location = location ?? <LocationModel>[];
  factory WarehouseModel.fromJson(Map<String, dynamic> json) =>
      _$WarehouseModelFromJson(json);
  Map<String, dynamic> toJson() => _$WarehouseModelToJson(this);
}
