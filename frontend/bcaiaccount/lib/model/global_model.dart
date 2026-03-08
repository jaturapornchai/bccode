import 'package:json_annotation/json_annotation.dart';

part 'global_model.g.dart';

@JsonSerializable(explicitToJson: true)
class SortDataModel {
  String code;
  int xorder;

  SortDataModel({required this.code, required this.xorder});

  factory SortDataModel.fromJson(Map<String, dynamic> json) =>
      _$SortDataModelFromJson(json);

  Map<String, dynamic> toJson() => _$SortDataModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class XSortModel {
  String guidfixed;
  String code;
  int xorder;

  XSortModel({
    required this.guidfixed,
    required this.xorder,
    required this.code,
  });

  factory XSortModel.fromJson(Map<String, dynamic> json) =>
      _$XSortModelFromJson(json);

  Map<String, dynamic> toJson() => _$XSortModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class LanguageModel {
  String? code;
  String? codeTranslator;
  String? name;
  bool? isuse;
  bool? isdefault;
  bool? isauto;
  bool? isdelete;

  LanguageModel({
    String? code,
    String? codeTranslator,
    String? name,
    bool? isuse,
    bool? isdefault,
    bool? isauto,
    bool? isdelete,
  }) : code = code ?? '',
       codeTranslator = codeTranslator ?? '',
       name = name ?? '',
       isuse = isuse ?? false,
       isdefault = isdefault ?? false,
       isauto = isauto ?? false,
       isdelete = isdelete ?? false;

  factory LanguageModel.fromJson(Map<String, dynamic> json) =>
      _$LanguageModelFromJson(json);

  Map<String, dynamic> toJson() => _$LanguageModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class LanguageDataModel {
  String? code;
  String name;

  LanguageDataModel({this.code, required this.name});

  factory LanguageDataModel.fromJson(Map<String, dynamic> json) =>
      _$LanguageDataModelFromJson(json);

  Map<String, dynamic> toJson() => _$LanguageDataModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class LanguageSystemModel {
  String code;
  String text;

  LanguageSystemModel({required this.code, required this.text});

  factory LanguageSystemModel.fromJson(Map<String, dynamic> json) =>
      _$LanguageSystemModelFromJson(json);
  Map<String, dynamic> toJson() => _$LanguageSystemModelToJson(this);
}

@JsonSerializable()
class ImageUpload {
  String uri;

  ImageUpload({required this.uri});

  factory ImageUpload.fromJson(Map<String, dynamic> json) =>
      _$ImageUploadFromJson(json);

  Map<String, dynamic> toJson() => _$ImageUploadToJson(this);
}

@JsonSerializable()
class ImagesModel {
  String uri;
  int xorder;

  ImagesModel({required this.uri, required this.xorder});

  factory ImagesModel.fromJson(Map<String, dynamic> json) =>
      _$ImagesModelFromJson(json);

  Map<String, dynamic> toJson() => _$ImagesModelToJson(this);
}

class SearchCodeAndNameAndUnitModel {
  String guid;
  String barcode;
  String code;
  List<LanguageDataModel> name;
  List<LanguageDataModel> names; // เพิ่ม names เพื่อความเข้ากันได้
  String unitcode;
  List<LanguageDataModel> unitname;
  String? shelfCode; // เพิ่มฟิลด์สำหรับรหัสชั้นวาง
  String? shelfName; // เพิ่มฟิลด์สำหรับชื่อชั้นวาง

  SearchCodeAndNameAndUnitModel({
    this.guid = "",
    required this.barcode,
    required this.code,
    required this.name,
    required this.unitcode,
    required this.unitname,
    this.shelfCode,
    this.shelfName,
  }) : names = name; // กำหนด names ให้เท่ากับ name
}

@JsonSerializable(explicitToJson: true)
class PromotionListModel {
  int code;
  String name;

  PromotionListModel({required this.code, required this.name});

  factory PromotionListModel.fromJson(Map<String, dynamic> json) =>
      _$PromotionListModelFromJson(json);

  Map<String, dynamic> toJson() => _$PromotionListModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class SlipListModel {
  String code;
  List<LanguageDataModel> names;
  List<LanguageDataModel>? headernames;

  SlipListModel({
    required this.code,
    required this.names,
    List<LanguageDataModel>? headernames,
  }) : headernames = headernames ?? [];

  factory SlipListModel.fromJson(Map<String, dynamic> json) =>
      _$SlipListModelFromJson(json);

  Map<String, dynamic> toJson() => _$SlipListModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class DayOfWeekModel {
  String code;
  String? name;

  DayOfWeekModel({required this.code, String? name}) : name = name ?? "";

  factory DayOfWeekModel.fromJson(Map<String, dynamic> json) =>
      _$DayOfWeekModelFromJson(json);

  Map<String, dynamic> toJson() => _$DayOfWeekModelToJson(this);
}

/// ปัดเศษสตางค์

class MoneyRoundPayModel {
  double begin;
  double end;
  double value;

  MoneyRoundPayModel({
    required this.begin,
    required this.end,
    required this.value,
  });
}

/// ประเภท Qr wallet
@JsonSerializable(explicitToJson: true)
class QrTypeModel {
  int code;
  String name;

  QrTypeModel({required this.code, required this.name});

  factory QrTypeModel.fromJson(Map<String, dynamic> json) =>
      _$QrTypeModelFromJson(json);

  Map<String, dynamic> toJson() => _$QrTypeModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class SearchGuidCodeNameModel {
  String guid;
  String code;
  List<LanguageDataModel> names;
  bool isCancel;
  SearchGuidCodeNameModel({
    required this.guid,
    required this.code,
    required this.names,
    this.isCancel = false,
  });

  factory SearchGuidCodeNameModel.fromJson(Map<String, dynamic> json) =>
      _$SearchGuidCodeNameModelFromJson(json);

  Map<String, dynamic> toJson() => _$SearchGuidCodeNameModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class FiltterBarcodeModel {
  bool? branch;

  FiltterBarcodeModel({bool? branch}) : branch = branch ?? false;

  factory FiltterBarcodeModel.fromJson(Map<String, dynamic> json) =>
      _$FiltterBarcodeModelFromJson(json);

  Map<String, dynamic> toJson() => _$FiltterBarcodeModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class ReportStockBalanceModel {
  @JsonKey(name: 'item_code')
  String itemCode;

  @JsonKey(name: 'barcode_list')
  String barCodeList;

  @JsonKey(name: 'item_name')
  String itemNameName;

  @JsonKey(name: 'unit_code')
  String? unitCode;

  @JsonKey(name: 'unit_name')
  String? unitName;

  @JsonKey(name: 'balance_qty')
  double? balanceQty;

  @JsonKey(name: 'average_cost')
  double? averageCost;

  @JsonKey(name: 'balance_amount')
  double? balanceAmount;

  @JsonKey(name: 'balance_word')
  String? balanceWord;

  @JsonKey(name: 'is_auto_packing')
  int? isAutoPacking;

  @JsonKey(name: 'warehouses')
  List<ReportStockBalanceWarehouseModel>? warehouses;

  ReportStockBalanceModel({
    required this.itemCode,
    required this.barCodeList,
    required this.itemNameName,
    this.unitCode,
    this.unitName,
    this.balanceQty,
    this.averageCost,
    this.balanceAmount,
    this.balanceWord,
    this.isAutoPacking,
    this.warehouses,
  });

  factory ReportStockBalanceModel.fromJson(Map<String, dynamic> json) =>
      _$ReportStockBalanceModelFromJson(json);

  Map<String, dynamic> toJson() => _$ReportStockBalanceModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class ReportStockBalanceWarehouseModel {
  @JsonKey(name: 'warehouse_code')
  String warehouseCode;

  @JsonKey(name: 'balance_qty')
  double balanceQty;

  @JsonKey(name: 'average_cost')
  double averageCost;

  @JsonKey(name: 'balance_word')
  String balanceWord;

  @JsonKey(name: 'balance_amount')
  double balanceAmount;

  @JsonKey(name: 'locations')
  List<ReportStockBalanceLocationModel> locations;

  ReportStockBalanceWarehouseModel({
    required this.warehouseCode,
    required this.balanceQty,
    required this.averageCost,
    required this.balanceWord,
    required this.balanceAmount,
    required this.locations,
  });

  factory ReportStockBalanceWarehouseModel.fromJson(
    Map<String, dynamic> json,
  ) => _$ReportStockBalanceWarehouseModelFromJson(json);

  Map<String, dynamic> toJson() =>
      _$ReportStockBalanceWarehouseModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class ReportStockBalanceLocationModel {
  @JsonKey(name: 'location_code')
  String locationCode;

  @JsonKey(name: 'balance_qty')
  double balanceQty;

  @JsonKey(name: 'balance_word')
  String balanceWord;

  ReportStockBalanceLocationModel({
    required this.locationCode,
    required this.balanceQty,
    required this.balanceWord,
  });

  factory ReportStockBalanceLocationModel.fromJson(Map<String, dynamic> json) =>
      _$ReportStockBalanceLocationModelFromJson(json);

  Map<String, dynamic> toJson() =>
      _$ReportStockBalanceLocationModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class ReportStockBalanceWhBarcodeModel {
  @JsonKey(name: 'whcode')
  String whCode;

  @JsonKey(name: 'item_code')
  String itemCode;

  @JsonKey(name: 'barcode_list')
  String barcodeList;

  @JsonKey(name: 'item_name')
  String itemName;

  @JsonKey(name: 'unit_code')
  String? unitCode;

  @JsonKey(name: 'unit_name')
  String? unitName;

  @JsonKey(name: 'balance_qty')
  double? balanceQty;

  @JsonKey(name: 'balance_word')
  String? balanceWord;

  ReportStockBalanceWhBarcodeModel({
    required this.itemCode,
    required this.whCode,
    required this.barcodeList,
    required this.itemName,
    this.unitCode,
    this.unitName,
    this.balanceQty,
    this.balanceWord,
  });

  factory ReportStockBalanceWhBarcodeModel.fromJson(
    Map<String, dynamic> json,
  ) => _$ReportStockBalanceWhBarcodeModelFromJson(json);

  Map<String, dynamic> toJson() =>
      _$ReportStockBalanceWhBarcodeModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class ReportStockBalanceLocationBarcodeModel {
  @JsonKey(name: 'whcode')
  String whCode;

  @JsonKey(name: 'location_code')
  String locationCode;

  @JsonKey(name: 'item_code')
  String itemCode;

  @JsonKey(name: 'barcode_list')
  String barcodeList;

  @JsonKey(name: 'item_name')
  String itemName;

  @JsonKey(name: 'unit_code')
  String? unitCode;

  @JsonKey(name: 'unit_name')
  String? unitName;

  @JsonKey(name: 'balance_qty')
  double? balanceQty;

  @JsonKey(name: 'balance_word')
  String? balanceWord;

  ReportStockBalanceLocationBarcodeModel({
    required this.whCode,
    required this.locationCode,
    required this.itemCode,
    required this.barcodeList,
    required this.itemName,
    this.unitCode,
    this.unitName,
    this.balanceQty,
    this.balanceWord,
  });

  factory ReportStockBalanceLocationBarcodeModel.fromJson(
    Map<String, dynamic> json,
  ) => _$ReportStockBalanceLocationBarcodeModelFromJson(json);

  Map<String, dynamic> toJson() =>
      _$ReportStockBalanceLocationBarcodeModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class ReportStockMovementModel {
  @JsonKey(name: 'itemCode')
  String itemCode;

  @JsonKey(name: 'name')
  String name;

  @JsonKey(name: 'unitCode')
  String unitCode;

  @JsonKey(name: 'unitName')
  String? unitName;

  @JsonKey(name: 'details')
  List<ReportStockMovementDetailModel> details;

  ReportStockMovementModel({
    required this.itemCode,
    required this.name,
    required this.unitCode,
    required this.unitName,
    required this.details,
  });

  factory ReportStockMovementModel.fromJson(Map<String, dynamic> json) =>
      _$ReportStockMovementModelFromJson(json);
  Map<String, dynamic> toJson() => _$ReportStockMovementModelToJson(this);
}

@JsonSerializable(explicitToJson: true)
class ReportStockMovementDetailModel {
  @JsonKey(name: 'isExtra')
  bool isExtra;

  @JsonKey(name: 'docDateTime')
  String docDateTime;

  @JsonKey(name: 'docNo')
  String docNo;

  @JsonKey(name: 'itemCode')
  String itemCode;

  @JsonKey(name: 'transFlag')
  int transFlag;

  @JsonKey(name: 'unitCode')
  String unitCode;

  @JsonKey(name: 'whCode')
  String whCode;

  @JsonKey(name: 'locationCode')
  String locationCode;

  @JsonKey(name: 'totalQty')
  double totalQty;

  @JsonKey(name: 'price')
  double price;

  @JsonKey(name: 'unitStand')
  double unitStand;

  @JsonKey(name: 'unitDivide')
  double unitDivide;

  @JsonKey(name: 'unitCost')
  double unitCost;

  @JsonKey(name: 'averageCost')
  double averageCost;

  @JsonKey(name: 'calcAmount')
  double calcAmount;

  @JsonKey(name: 'balanceAmount')
  double balanceAmount;

  @JsonKey(name: 'balanceQty')
  double balanceQty;

  @JsonKey(name: 'docRef')
  String docRef;

  ReportStockMovementDetailModel({
    required this.isExtra,
    required this.docDateTime,
    required this.itemCode,
    required this.docNo,
    required this.transFlag,
    required this.unitCode,
    required this.whCode,
    required this.locationCode,
    required this.totalQty,
    required this.price,
    required this.unitStand,
    required this.unitDivide,
    required this.unitCost,
    required this.averageCost,
    required this.calcAmount,
    required this.balanceAmount,
    required this.balanceQty,
    required this.docRef,
  });

  factory ReportStockMovementDetailModel.fromJson(Map<String, dynamic> json) =>
      _$ReportStockMovementDetailModelFromJson(json);

  Map<String, dynamic> toJson() => _$ReportStockMovementDetailModelToJson(this);
}

// Struct สำหรับเก็บข้อมูลสินค้าที่เลือก
class SelectedProductItemStruct {
  final String itemCode;
  final String barcode;
  final String itemName;

  SelectedProductItemStruct({
    required this.itemCode,
    required this.barcode,
    required this.itemName,
  });

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is SelectedProductItemStruct &&
          runtimeType == other.runtimeType &&
          barcode == other.barcode;

  @override
  int get hashCode => barcode.hashCode;
}
