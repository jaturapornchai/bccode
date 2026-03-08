import 'package:smlaicloud/model/global_model.dart';
import 'package:json_annotation/json_annotation.dart';

part 'shelf_product_model.g.dart';

/// Model สำหรับสินค้าที่จะเพิ่มเข้าชั้นวาง (POST request)
@JsonSerializable()
class ShelfProductBulkAddModel {
  List<ShelfProductItemModel> products;

  ShelfProductBulkAddModel({
    required this.products,
  });

  factory ShelfProductBulkAddModel.fromJson(Map<String, dynamic> json) =>
      _$ShelfProductBulkAddModelFromJson(json);

  Map<String, dynamic> toJson() => _$ShelfProductBulkAddModelToJson(this);
}

/// Model สำหรับแต่ละสินค้าในชั้นวาง
@JsonSerializable()
class ShelfProductItemModel {
  String barcode;
  String guidfixed;
  List<LanguageDataModel> names;
  String? unitcode;
  List<LanguageDataModel>? unitnames;

  ShelfProductItemModel({
    required this.barcode,
    required this.guidfixed,
    required this.names,
    this.unitcode,
    List<LanguageDataModel>? unitnames,
  }) : unitnames = unitnames ?? <LanguageDataModel>[];

  factory ShelfProductItemModel.fromJson(Map<String, dynamic> json) =>
      _$ShelfProductItemModelFromJson(json);

  Map<String, dynamic> toJson() => _$ShelfProductItemModelToJson(this);
}

/// Model สำหรับลบสินค้าออกจากชั้นวาง (DELETE request)
@JsonSerializable()
class ShelfProductBulkDeleteModel {
  List<String> productguidfixedlist;

  ShelfProductBulkDeleteModel({
    required this.productguidfixedlist,
  });

  factory ShelfProductBulkDeleteModel.fromJson(Map<String, dynamic> json) =>
      _$ShelfProductBulkDeleteModelFromJson(json);

  Map<String, dynamic> toJson() => _$ShelfProductBulkDeleteModelToJson(this);
}

/// Model สำหรับข้อมูลสินค้าที่มีอยู่ในชั้นวางแล้ว (สำหรับแสดงผล)
@JsonSerializable()
class ShelfProductDisplayModel {
  String barcode;
  String guidfixed;
  List<LanguageDataModel> names;
  bool isSelected;
  String unitcode;
  List<LanguageDataModel> unitnames;

  ShelfProductDisplayModel({
    required this.barcode,
    required this.guidfixed,
    required this.names,
    this.isSelected = false,
    required this.unitcode,
    required this.unitnames,
  });

  factory ShelfProductDisplayModel.fromJson(Map<String, dynamic> json) =>
      _$ShelfProductDisplayModelFromJson(json);

  Map<String, dynamic> toJson() => _$ShelfProductDisplayModelToJson(this);

  ShelfProductDisplayModel copyWith({
    String? barcode,
    String? guidfixed,
    List<LanguageDataModel>? names,
    bool? isSelected,
    String? unitcode,
    List<LanguageDataModel>? unitnames,
  }) {
    return ShelfProductDisplayModel(
      barcode: barcode ?? this.barcode,
      guidfixed: guidfixed ?? this.guidfixed,
      names: names ?? this.names,
      isSelected: isSelected ?? this.isSelected,
      unitcode: unitcode ?? this.unitcode,
      unitnames: unitnames ?? this.unitnames,
    );
  }
}
