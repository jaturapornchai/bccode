import 'package:json_annotation/json_annotation.dart';

part 'master_model.g.dart';

/// คลังสินค้า
@JsonSerializable(explicitToJson: true)
class WareHouseModel {
  String code; // รหัสคลังสินค้า
  String imageuri; // uri รูปภาพ
  List<LocationModel> locatecodes; // รหัสตำแหน่ง

  WareHouseModel({
    required this.code,
    required this.imageuri,
    required this.locatecodes,
  });

  factory WareHouseModel.fromJson(Map<String, dynamic> json) =>
      _$WareHouseModelFromJson(json);

  Map<String, dynamic> toJson() => _$WareHouseModelToJson(this);
}

/// บริเวณที่เก็บสินค้า
@JsonSerializable(explicitToJson: true)
class LocationModel {
  String code; // รหัสบริเวณ
  String imageuri; // uri รูปภาพ
  double width; // ความกว้าง
  double height; // ความสูง
  double depth; // ความลึก
  String description; // รายละเอียด
  List<String> recommendshelfs; // แนะนำที่วาง

  LocationModel({
    required this.code,
    required this.imageuri,
    required this.width,
    required this.height,
    required this.depth,
    required this.description,
    required this.recommendshelfs,
  });

  factory LocationModel.fromJson(Map<String, dynamic> json) =>
      _$LocationModelFromJson(json);

  Map<String, dynamic> toJson() => _$LocationModelToJson(this);
}

/// Shelf สินค้า
@JsonSerializable(explicitToJson: true)
class ProductShelfModel {
  String code; // รหัสชั้นวาง
  String imageuri; // uri รูปภาพ
  double width; // ความกว้าง
  double height; // ความสูง
  double depth; // ความลึก
  String description; // รายละเอียด

  ProductShelfModel({
    required this.code,
    required this.imageuri,
    required this.width,
    required this.height,
    required this.depth,
    required this.description,
  });

  factory ProductShelfModel.fromJson(Map<String, dynamic> json) =>
      _$ProductShelfModelFromJson(json);

  Map<String, dynamic> toJson() => _$ProductShelfModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class UploadImageModel {
  String uri;

  UploadImageModel({
    required this.uri,
  });

  factory UploadImageModel.fromJson(Map<String, dynamic> json) =>
      _$UploadImageModelFromJson(json);

  Map<String, dynamic> toJson() => _$UploadImageModelToJson(this);
}
