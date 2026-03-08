import 'package:json_annotation/json_annotation.dart';

part 'gross_profit_by_product_model.g.dart';

/// Model for Gross Profit By Product detail data
/// Represents profit analysis per product/item
@JsonSerializable()
class GrossProfitByProductModel {
  /// วันที่เอกสาร
  final String? docdate;

  /// เลขที่เอกสาร
  final String? docno;

  /// บาร์โค้ด
  final String? barcode;

  /// บาร์โค้ดหลักอ้างอิง
  final String? mainbarcoderef;

  /// รหัสสินค้า
  final String? itemcode;

  /// ชื่อสินค้าหลายภาษา
  final List<ProductName>? names;

  /// รหัสหน่วย
  final String? unitcode;

  /// ชื่อหน่วยหลายภาษา
  final List<UnitName>? unitnames;

  /// จำนวน
  final double? qty;

  /// ยอดรวมไม่รวม VAT
  final double? sumamountexcludevat;

  /// ต้นทุนรวม
  final double? totalcost;

  /// กำไร
  final double? pal;

  /// เปอร์เซ็นต์กำไร
  @JsonKey(name: 'per_pal')
  final double? perPal;

  /// รหัสหน่วยในเอกสาร
  @JsonKey(name: 'unitcode_doc')
  final String? unitcodeDoc;

  /// ชื่อหน่วยในเอกสารหลายภาษา
  @JsonKey(name: 'unitnames_doc')
  final List<UnitName>? unitnamesDoc;

  /// จำนวนขาย
  @JsonKey(name: 'qty_sale')
  final double? qtySale;

  /// ราคา
  final double? price;

  /// ยอดรวม
  final double? sumamount;

  const GrossProfitByProductModel({
    this.docdate,
    this.docno,
    this.barcode,
    this.mainbarcoderef,
    this.itemcode,
    this.names,
    this.unitcode,
    this.unitnames,
    this.qty,
    this.sumamountexcludevat,
    this.totalcost,
    this.pal,
    this.perPal,
    this.unitcodeDoc,
    this.unitnamesDoc,
    this.qtySale,
    this.price,
    this.sumamount,
  });

  factory GrossProfitByProductModel.fromJson(Map<String, dynamic> json) =>
      _$GrossProfitByProductModelFromJson(json);

  Map<String, dynamic> toJson() => _$GrossProfitByProductModelToJson(this);
}

/// Model for product name in different languages
@JsonSerializable()
class ProductName {
  /// รหัสภาษา (th, en, etc.)
  final String? code;

  /// ชื่อสินค้า
  final String? name;

  /// สร้างอัตโนมัติหรือไม่
  final bool? isauto;

  /// ลบแล้วหรือไม่
  final bool? isdelete;

  const ProductName({this.code, this.name, this.isauto, this.isdelete});

  factory ProductName.fromJson(Map<String, dynamic> json) =>
      _$ProductNameFromJson(json);

  Map<String, dynamic> toJson() => _$ProductNameToJson(this);
}

/// Model for unit name in different languages
@JsonSerializable()
class UnitName {
  /// รหัสภาษา (th, en, etc.)
  final String? code;

  /// ชื่อหน่วย
  final String? name;

  /// สร้างอัตโนมัติหรือไม่
  final bool? isauto;

  /// ลบแล้วหรือไม่
  final bool? isdelete;

  const UnitName({this.code, this.name, this.isauto, this.isdelete});

  factory UnitName.fromJson(Map<String, dynamic> json) =>
      _$UnitNameFromJson(json);

  Map<String, dynamic> toJson() => _$UnitNameToJson(this);
}
